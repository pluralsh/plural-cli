package main

import (
	"fmt"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
)

func workspaceCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "kube-init",
			Usage:     "generates kubernetes credentials for this subworkspace",
			Action:    kubeInit,
		},
		{
			Name:      "helm",
			Usage:     "upgrade/installs the helm chart for this subworkspace",
			ArgsUsage: "NAME",
			Action:    bounceHelm,
		},
		{
			Name:      "helm-diff",
			Usage:     "diffs the helm release for this subworkspace",
			ArgsUsage: "NAME",
			Action:    diffHelm,
		},
		{
			Name:      "terraform-diff",
			Usage:     "diffs the helm release for this subworkspace",
			ArgsUsage: "NAME",
			Action:    diffTerraform,
		},
		{
			Name:      "crds",
			Usage:     "installs the crds for this repo",
			ArgsUsage: "NAME",
			Action:    createCrds,
		},
	}
}

func kubeInit(c *cli.Context) error {
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

func bounceHelm(c *cli.Context) error {
	name := c.Args().Get(0)
	minimal, err := wkspace.Minimal(name)
	if err != nil {
		return err
	}

	return minimal.BounceHelm()
}

func diffHelm(c *cli.Context) error {
	name := c.Args().Get(0)
	minimal, err := wkspace.Minimal(name)
	if err != nil {
		return err
	}

	return minimal.DiffHelm()
}

func diffTerraform(c *cli.Context) error {
	name := c.Args().Get(0)
	minimal, err := wkspace.Minimal(name)
	if err != nil {
		return err
	}

	return minimal.DiffTerraform()
}

func createCrds(c *cli.Context) error {
	if empty, err := utils.IsEmpty("crds"); err != nil || empty {
		return err
	}
	return filepath.Walk("crds", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if err := utils.Exec("kubectl", "create", "-f", path); err != nil {
			return utils.Exec("kubectl", "replace", "-f", path)
		}

		return nil
	})
}
