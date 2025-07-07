package up

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

type templatePair struct {
	from      string
	to        string
	overwrite bool
	cloud     bool
	cloudless bool
}

//nolint:gocyclo
func (ctx *Context) Generate(gitRef string) (dir string, err error) {
	dir, err = os.MkdirTemp("", "sampledir")
	ctx.dir = dir
	hasDomain := ctx.Manifest.AppDomain != ""
	if err != nil {
		return
	}

	if err = git.PathClone("https://github.com/pluralsh/bootstrap.git", gitRef, dir); err != nil {
		return
	}

	prov := ctx.Provider.Name()
	tpls := []templatePair{
		{from: ctx.path("charts/runtime/values.yaml.tpl"), to: "./helm-values/runtime.yaml", overwrite: true},
		{from: ctx.path("helm/certmanager.yaml"), to: "./helm-values/certmanager.yaml", overwrite: true},
		{from: ctx.path("helm/flux.yaml"), to: "./helm-values/flux.yaml", overwrite: true},
		{from: ctx.path(fmt.Sprintf("templates/providers/bootstrap/%s.tf", prov)), to: "terraform/mgmt/provider.tf"},
		{from: ctx.path(fmt.Sprintf("templates/setup/providers/%s.tf", prov)), to: "terraform/mgmt/mgmt.tf"},
		{from: ctx.path("templates/setup/console.tf"), to: "terraform/mgmt/console.tf", cloudless: true},
		{from: ctx.path(fmt.Sprintf("templates/providers/apps/%s.tf", prov)), to: "terraform/apps/provider.tf", cloudless: true},
		{from: ctx.path("templates/providers/apps/cloud.tf"), to: "terraform/apps/provider.tf", cloud: true},
		{from: ctx.path("templates/setup/cd.tf"), to: "terraform/apps/cd.tf"},
		{from: ctx.path("README.md"), to: "README.md", overwrite: true},
	}

	for _, tpl := range tpls {
		if utils.Exists(tpl.to) && !tpl.overwrite {
			fmt.Printf("%s already exists, skipping for now...\n", tpl.to)
			continue
		}

		if tpl.cloudless && ctx.Cloud {
			continue
		}

		if tpl.cloud && !ctx.Cloud {
			continue
		}

		if err = ctx.templateFrom(tpl.from, tpl.to); err != nil {
			err = fmt.Errorf("failed to template %s: %w", tpl.from, err)
			return
		}
	}

	copies := []templatePair{
		{from: ctx.path("terraform/modules/clusters"), to: "terraform/modules/clusters"},
		{from: ctx.path(fmt.Sprintf("terraform/clouds/%s", prov)), to: "terraform/mgmt/cluster"},
		{from: ctx.path("setup"), to: "bootstrap"},
		{from: ctx.path(fmt.Sprintf("terraform/core-infra/%s", prov)), to: "terraform/core-infra"},
		{from: ctx.path("templates"), to: "templates"},
		{from: ctx.path("resources"), to: "resources"},
		{from: ctx.path("services"), to: "services"},
		{from: ctx.path("helm"), to: "helm"},
	}

	if ctx.Cloud {
		copies = append(copies, templatePair{from: ctx.path("o11y"), to: "bootstrap/o11y"})
	}

	if hasDomain {
		copies = append(copies, templatePair{from: ctx.path("network"), to: "bootstrap/network"})
	}

	for _, copy := range copies {
		if utils.Exists(copy.to) && !copy.overwrite {
			continue
		}

		if err = utils.CopyDir(copy.from, copy.to); err != nil {
			return
		}
	}

	postTemplates := []templatePair{
		{from: "terraform/core-infra/network.tf", to: "terraform/core-infra/network.tf"},
	}

	if hasDomain {
		postTemplates = append(postTemplates, templatePair{from: "terraform/core-infra/dns.tf", to: "terraform/core-infra/dns.tf"})
	}

	for _, tpl := range postTemplates {
		if err = ctx.templateFrom(tpl.from, tpl.to); err != nil {
			err = fmt.Errorf("failed to template %s: %w", tpl.from, err)
			return
		}
	}

	toRemove := make([]string, 0)
	if ctx.Cloud {
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

	ctx.changeDelims()
	overwrites := []templatePair{
		{from: "resources/monitoring/services", to: "resources/monitoring/services"},
		{from: "resources/policy/services", to: "resources/policy/services"},
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
				if err = ctx.templateFrom(file, destFile); err != nil {
					err = fmt.Errorf("failed to template %s: %w", file, err)
					return dir, err
				}
			}

			continue
		}

		if err = ctx.templateFrom(tpl.from, tpl.to); err != nil {
			return
		}
	}

	return
}

func (ctx *Context) afterSetup() error {
	prov := ctx.Provider.Name()
	overwrites := []templatePair{
		{from: ctx.path(fmt.Sprintf("templates/setup/stacks/%s.yaml", prov)), to: "bootstrap/stacks/serviceaccount.yaml"},
		{from: "bootstrap/stacks/mgmt.yaml", to: "bootstrap/stacks/mgmt.yaml"},
		{from: "bootstrap/stacks/core-infra.yaml", to: "bootstrap/stacks/core-infra.yaml"},
	}

	ctx.Delims = nil
	for _, tpl := range overwrites {
		if err := ctx.templateFrom(tpl.from, tpl.to); err != nil {
			err = fmt.Errorf("failed to template %s: %w", tpl.from, err)
			return err
		}
	}

	redacts := []templatePair{
		{from: "./terraform/mgmt/provider.tf"},
	}

	for _, redact := range redacts {
		if err := ctx.redact(redact.from); err != nil {
			return err
		}
	}

	uncomments := []templatePair{
		{from: "./terraform/mgmt/cluster/eks.tf"},
	}

	for _, uncomment := range uncomments {
		if utils.Exists(uncomment.from) {
			if err := ctx.uncomment(uncomment.from); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ctx *Context) path(p string) string {
	return filepath.Join(ctx.dir, p)
}
