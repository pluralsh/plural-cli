package main

import (
	"fmt"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path/filepath"
)

func workspaceCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "docker-credentials",
			Usage:     "create a docker credentials secret for this workspace",
			ArgsUsage: "NAME",
			Action:    ensureDockerCredentials,
		},
		{
			Name:      "helm-init",
			Usage:     "pushes a helm chart",
			ArgsUsage: "NAME",
			Action:    helmInit,
		},
		{
			Name:      "kube-init",
			Usage:     "generates kubernetes credentials for this subworkspace",
			ArgsUsage: "NAME",
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
			ArgsUsage: "REPO",
			Action:    createCrds,
		},
	}
}

func ensureDockerCredentials(c *cli.Context) error {
	name := c.Args().Get(0)
	minimal, err := wkspace.Minimal(name)
	if err != nil {
		return err
	}

	return minimal.EnsurePullCredentials()
}

func helmInit(c *cli.Context) error {
	name := c.Args().Get(0)
	minimal, err := wkspace.Minimal(name)
	if err != nil {
		return err
	}

	return minimal.HelmInit(false)
}

func kubeInit(c *cli.Context) error {
	root, found := utils.ProjectRoot()
	if !found {
		return fmt.Errorf("Project not initialized, run `plural init` to set up a workspace")
	}

	manifestPath := filepath.Join(root, "manifest.yaml")
	prov, err := provider.Bootstrap(manifestPath, true)
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

	cmd := exec.Command("kubectl", "apply", "-f", "crds")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
