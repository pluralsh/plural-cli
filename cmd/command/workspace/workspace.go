package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/helm"
	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
	"github.com/urfave/cli"
	"helm.sh/helm/v3/pkg/action"
)

type Plural struct {
	client.Plural
	HelmConfiguration *action.Configuration
}

func Command(clients client.Plural, helmConfiguration *action.Configuration) cli.Command {
	p := Plural{
		Plural:            clients,
		HelmConfiguration: helmConfiguration,
	}
	return cli.Command{
		Name:        "workspace",
		Aliases:     []string{"wkspace"},
		Usage:       "Commands for managing installations in your workspace",
		Subcommands: p.workspaceCommands(),
		Category:    "Workspace",
	}
}

func (p *Plural) workspaceCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "kube-init",
			Usage:  "generates kubernetes credentials for this subworkspace",
			Action: common.LatestVersion(kubeInit),
		},
		{
			Name:      "readme",
			Usage:     "generate chart readme for an app",
			ArgsUsage: "{app}",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "dry-run",
					Usage: "output to stdout instead of to a file",
				},
			},
			Action: common.LatestVersion(func(c *cli.Context) error { return common.AppReadme(c.Args().Get(0), c.Bool("dry-run")) }),
		},
		{
			Name:      "helm",
			Usage:     "upgrade/installs the helm chart for this subworkspace",
			ArgsUsage: "{name}",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "skip",
					Usage: "helm sub-chart to skip. can be passed multiple times",
				},
				cli.StringSliceFlag{
					Name:  "set",
					Usage: "helm value to set. can be passed multiple times",
				},
				cli.StringSliceFlag{
					Name:  "setJSON",
					Usage: "JSON helm value to set. can be passed multiple times",
				},
				cli.BoolFlag{
					Name:  "wait",
					Usage: "have helm wait until all pods are in ready state",
				},
			},
			Action: common.LatestVersion(common.InitKubeconfig(p.bounceHelm)),
		},
		{
			Name:      "helm-diff",
			Usage:     "diffs the helm release for this subworkspace",
			ArgsUsage: "{name}",
			Action:    common.LatestVersion(p.diffHelm),
		},
		{
			Name:      "helm-deps",
			Usage:     "updates the helm dependencies for this workspace",
			ArgsUsage: "{path}",
			Action:    common.LatestVersion(updateDeps),
		},
		{
			Name:      "terraform-diff",
			Usage:     "diffs the helm release for this subworkspace",
			ArgsUsage: "{name}",
			Action:    common.LatestVersion(p.diffTerraform),
		},
		{
			Name:      "crds",
			Usage:     "installs the crds for this repo",
			ArgsUsage: "{name}",
			Action:    common.LatestVersion(common.InitKubeconfig(p.createCrds)),
		},
		{
			Name:      "helm-template",
			Usage:     "templates the helm values to stdout",
			ArgsUsage: "{name}",
			Action:    common.LatestVersion(common.RequireArgs(p.templateHelm, []string{"{name}"})),
		},
		{
			Name:      "helm-mapkubeapis",
			Usage:     "updates in-place Helm release metadata that contains deprecated or removed Kubernetes APIs to a new instance with supported Kubernetes APIs",
			ArgsUsage: "{name}",
			Action:    common.LatestVersion(common.RequireArgs(p.mapkubeapis, []string{"{name}"})),
		},
	}
}

func kubeInit(_ *cli.Context) error {
	_, found := utils.ProjectRoot()
	if !found {
		return fmt.Errorf("Project not initialized, run `plural init` to set up a workspace")
	}

	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}

	return prov.KubeConfig()
}

func (p *Plural) bounceHelm(c *cli.Context) error {
	name := c.Args().Get(0)
	minimal, err := wkspace.Minimal(name, p.HelmConfiguration)
	if err != nil {
		return err
	}

	var skipArgs []string
	if c.IsSet("skip") {
		for _, skipChart := range c.StringSlice("skip") {
			skipString := fmt.Sprintf("%s.enabled=false", skipChart)
			skipArgs = append(skipArgs, skipString)
		}
	}
	var setArgs []string
	if c.IsSet("set") {
		setArgs = append(setArgs, c.StringSlice("set")...)
	}

	var setJSONArgs []string
	if c.IsSet("setJSON") {
		setJSONArgs = append(setJSONArgs, c.StringSlice("setJSON")...)
	}

	return minimal.BounceHelm(c.IsSet("wait"), skipArgs, setArgs, setJSONArgs)
}

func (p *Plural) diffHelm(c *cli.Context) error {
	name := c.Args().Get(0)
	minimal, err := wkspace.Minimal(name, p.HelmConfiguration)
	if err != nil {
		return err
	}

	return minimal.DiffHelm()
}

func (p *Plural) diffTerraform(c *cli.Context) error {
	name := c.Args().Get(0)
	minimal, err := wkspace.Minimal(name, p.HelmConfiguration)
	if err != nil {
		return err
	}

	return minimal.DiffTerraform()
}

func (p *Plural) createCrds(_ *cli.Context) error {
	err := p.InitKube()
	if err != nil {
		return err
	}
	if empty, err := utils.IsEmpty("crds"); err != nil || empty {
		return nil
	}

	return filepath.Walk("crds", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		err = p.Kube.Apply(path, true)
		if err != nil {
			errStr := fmt.Sprint(err)
			if strings.Contains(errStr, "invalid apiVersion \"client.authentication.k8s.io/v1alpha1\"") {
				return fmt.Errorf("failed with %s, this is usually due to your aws cli version being out of date", errStr)
			}
			return err
		}

		return nil
	})
}

func updateDeps(c *cli.Context) error {
	path := c.Args().Get(0)
	if path == "" {
		path = "."
	}

	return helm.UpdateDependencies(path)
}

func (p *Plural) templateHelm(c *cli.Context) error {
	name := c.Args().Get(0)
	minimal, err := wkspace.Minimal(name, p.HelmConfiguration)
	if err != nil {
		return err
	}

	return minimal.TemplateHelm()
}

func (p *Plural) mapkubeapis(c *cli.Context) error {
	name := c.Args().Get(0)
	minimal, err := wkspace.Minimal(name, p.HelmConfiguration)
	if err != nil {
		return err
	}

	return minimal.MapKubeApis()
}
