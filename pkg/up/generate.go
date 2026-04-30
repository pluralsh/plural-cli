package up

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

const consoleValuesTemplateURL = "https://raw.githubusercontent.com/pluralsh/console/refs/heads/master/templates/values.yaml.liquid"

type templatePair struct {
	from      string
	to        string
	overwrite bool
	cloud     bool
	cloudless bool
}

//nolint:gocyclo
func (c *Context) Generate(gitRef string) (dir string, err error) {
	if c.Provider.Name() == api.BYOK && c.Cloud {
		return "", nil
	}
	dir, err = os.MkdirTemp("", "sampledir")
	c.dir = dir
	hasDomain := c.Manifest.AppDomain != ""
	if err != nil {
		return
	}

	if err = git.PathClone("https://github.com/pluralsh/bootstrap.git", gitRef, dir); err != nil {
		return
	}

	prov := c.Provider.Name()
	tpls := []templatePair{
		{from: c.path("charts/runtime/values.yaml.tpl"), to: "./temp/helm/runtime.yaml", overwrite: true},
		{from: c.path("charts/runtime/values.yaml.liquid.tpl"), to: "./helm/runtime.yaml.liquid", overwrite: true},
		{from: c.path(fmt.Sprintf("templates/providers/bootstrap/%s.tf", prov)), to: "terraform/mgmt/provider.tf"},
		{from: c.path(fmt.Sprintf("templates/setup/providers/%s.tf", prov)), to: "terraform/mgmt/mgmt.tf"},
		{from: c.path("templates/setup/console.tf"), to: "terraform/mgmt/console.tf", cloudless: true},
		{from: c.path(fmt.Sprintf("templates/providers/apps/%s.tf", prov)), to: "terraform/apps/provider.tf", cloudless: true},
		{from: c.path("templates/providers/apps/cloud.tf"), to: "terraform/apps/provider.tf", cloud: true},
		{from: c.path("templates/setup/cd.tf"), to: "terraform/apps/cd.tf"},
		{from: c.path("README.md"), to: "README.md", overwrite: true},
	}

	if prov == api.ProviderGCP {
		tpls = append(tpls, templatePair{from: c.path("templates/setup/config_secrets_gcp.tf"), to: "terraform/mgmt/config_secrets.tf", cloudless: true})
	} else {
		tpls = append(tpls, templatePair{from: c.path("templates/setup/config_secrets.tf"), to: "terraform/mgmt/config_secrets.tf", cloudless: true})
	}

	for _, tpl := range tpls {
		if utils.Exists(tpl.to) && !tpl.overwrite {
			fmt.Printf("%s already exists, skipping for now...\n", tpl.to)
			continue
		}

		if tpl.cloudless && c.Cloud {
			continue
		}

		if tpl.cloud && !c.Cloud {
			continue
		}

		if err = c.templateFrom(tpl.from, tpl.to); err != nil {
			err = fmt.Errorf("failed to template %s: %w", tpl.from, err)
			return
		}
	}

	copies := []templatePair{
		{from: c.path("terraform/modules/clusters"), to: "terraform/modules/clusters", overwrite: true},
		{from: c.path(fmt.Sprintf("terraform/clouds/%s", prov)), to: "terraform/mgmt/cluster", overwrite: true},
		{from: c.path("setup"), to: "bootstrap", overwrite: true},
		{from: c.path(fmt.Sprintf("terraform/core-infra/%s", prov)), to: "terraform/core-infra"},
		{from: c.path("templates"), to: "templates", overwrite: true},
		{from: c.path("services"), to: "services", overwrite: true},
		{from: c.path("helm"), to: "helm", overwrite: true},
	}

	if c.Cloud {
		copies = append(copies, templatePair{from: c.path("o11y"), to: "bootstrap/o11y"})
	}

	if hasDomain {
		copies = append(copies, templatePair{from: c.path("network"), to: "bootstrap/network"})
	}

	for _, copy := range copies {
		if utils.Exists(copy.to) && !copy.overwrite {
			continue
		}

		if err = utils.CopyDir(copy.from, copy.to); err != nil {
			return
		}
	}

	if err = utils.DownloadFile(filepath.Join("helm", "console.yaml.liquid"), consoleValuesTemplateURL); err != nil {
		return "", fmt.Errorf("fetch console values template: %w", err)
	}

	postTemplates := []templatePair{
		{from: "terraform/core-infra/network.tf", to: "terraform/core-infra/network.tf"},
	}

	if hasDomain {
		postTemplates = append(postTemplates, templatePair{from: "terraform/core-infra/dns.tf", to: "terraform/core-infra/dns.tf"})
	}

	for _, tpl := range postTemplates {
		if !utils.Exists(tpl.from) {
			continue
		}
		if err = c.templateFrom(tpl.from, tpl.to); err != nil {
			err = fmt.Errorf("failed to template %s: %w (you might need to regenerate your repo from scratch if partially applied)", tpl.from, err)
			return
		}
	}

	toRemove := make([]string, 0)
	if c.Cloud {
		toRemove = append(toRemove, "bootstrap/console.yaml")
	}

	if !hasDomain {
		toRemove = append(toRemove, "terraform/core-infra/dns.tf")
	}

	if prov != "aws" {
		toRemove = append(toRemove, "bootstrap/network/aws-load-balancer.yaml")
	}

	for _, f := range toRemove {
		os.Remove(f)
	}

	c.changeDelims()
	overwrites := []templatePair{
		{from: "bootstrap", to: "bootstrap"},
	}

	for _, tpl := range overwrites {
		if utils.IsDir(tpl.from) {
			files, err := utils.ListDirectory(tpl.from)
			if err != nil {
				return dir, err
			}

			for _, file := range files {
				destFile, err := filepath.Rel(tpl.from, file)
				if err != nil {
					return dir, err
				}

				destFile = filepath.Join(tpl.to, destFile)
				if err = c.templateFrom(file, destFile); err != nil {
					err = fmt.Errorf("failed to template %s: %w", file, err)
					return dir, err
				}
			}

			continue
		}

		if err = c.templateFrom(tpl.from, tpl.to); err != nil {
			return
		}
	}

	return
}

func (c *Context) afterSetup() error {
	prov := c.Provider.Name()
	overwrites := []templatePair{
		{from: c.path(fmt.Sprintf("templates/setup/stacks/%s.yaml", prov)), to: "bootstrap/stacks/serviceaccount.yaml"},
		{from: "bootstrap/stacks/mgmt.yaml", to: "bootstrap/stacks/mgmt.yaml"},
		{from: "bootstrap/stacks/core-infra.yaml", to: "bootstrap/stacks/core-infra.yaml"},
	}

	c.Delims = nil
	for _, tpl := range overwrites {
		if err := c.templateFrom(tpl.from, tpl.to); err != nil {
			err = fmt.Errorf("failed to template %s: %w", tpl.from, err)
			return err
		}
	}

	redacts := []templatePair{
		{from: "./terraform/mgmt/provider.tf"},
	}

	for _, redact := range redacts {
		if err := c.redact(redact.from); err != nil {
			return err
		}
	}

	uncomments := []templatePair{
		{from: "./terraform/mgmt/cluster/eks.tf"},
	}

	for _, uncomment := range uncomments {
		if utils.Exists(uncomment.from) {
			if err := c.uncomment(uncomment.from); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Context) path(p string) string {
	return filepath.Join(c.dir, p)
}
