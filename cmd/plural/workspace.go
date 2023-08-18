package plural

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural/pkg/helm"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/urfave/cli"
)

func (p *Plural) workspaceCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "kube-init",
			Usage:  "generates kubernetes credentials for this subworkspace",
			Action: latestVersion(kubeInit),
		},
		{
			Name:      "helm",
			Usage:     "upgrade/installs the helm chart for this subworkspace",
			ArgsUsage: "NAME",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "skip",
					Usage: "helm sub-chart to skip. can be passed multiple times",
				},
				cli.StringSliceFlag{
					Name:  "set",
					Usage: "helm value to set. can be passed multiple times",
				},
				cli.BoolFlag{
					Name:  "wait",
					Usage: "have helm wait until all pods are in ready state",
				},
			},
			Action: latestVersion(initKubeconfig(p.bounceHelm)),
		},
		{
			Name:      "helm-diff",
			Usage:     "diffs the helm release for this subworkspace",
			ArgsUsage: "NAME",
			Action:    latestVersion(p.diffHelm),
		},
		{
			Name:      "helm-deps",
			Usage:     "updates the helm dependencies for this workspace",
			ArgsUsage: "PATH",
			Action:    latestVersion(updateDeps),
		},
		{
			Name:      "terraform-diff",
			Usage:     "diffs the helm release for this subworkspace",
			ArgsUsage: "NAME",
			Action:    latestVersion(p.diffTerraform),
		},
		{
			Name:      "crds",
			Usage:     "installs the crds for this repo",
			ArgsUsage: "NAME",
			Action:    latestVersion(initKubeconfig(p.createCrds)),
		},
		{
			Name:      "helm-template",
			Usage:     "templates the helm values to stdout",
			ArgsUsage: "NAME",
			Action:    latestVersion(requireArgs(p.templateHelm, []string{"NAME"})),
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

	skipArgs := []string{}
	if c.IsSet("skip") {
		for _, skipChart := range c.StringSlice("skip") {
			skipString := fmt.Sprintf("%s.enabled=false", skipChart)
			skipArgs = append(skipArgs, skipString)
		}
	}
	setArgs := []string{}
	if c.IsSet("set") {
		setArgs = append(setArgs, c.StringSlice("set")...)
	}

	return minimal.BounceHelm(c.IsSet("wait"), skipArgs, setArgs)
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
