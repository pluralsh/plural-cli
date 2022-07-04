package main

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

func workspaceCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "kube-init",
			Usage:  "generates kubernetes credentials for this subworkspace",
			Action: kubeInit,
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
				cli.BoolFlag{
					Name:  "wait",
					Usage: "have helm wait until all pods are in ready state",
				},
			},
			Action: bounceHelm,
		},
		{
			Name:      "helm-diff",
			Usage:     "diffs the helm release for this subworkspace",
			ArgsUsage: "NAME",
			Action:    diffHelm,
		},
		{
			Name:      "helm-deps",
			Usage:     "updates the helm dependencies for this workspace",
			ArgsUsage: "PATH",
			Action:    updateDeps,
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

	args := []string{}
	if c.IsSet("wait") {
		args = append(args, "--wait")
	}
	if c.IsSet("skip") {
		for _, skipChart := range c.StringSlice("skip") {
			skipString := fmt.Sprintf("%s.enabled=false", skipChart)
			skip := []string{"--set", skipString}
			args = append(args, skip...)
		}
	}

	return minimal.BounceHelm(args...)
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

		if err = utils.Exec("kubectl", "create", "-f", path); err != nil {
			err = utils.Exec("kubectl", "replace", "-f", path)
		}

		if err != nil {
			errStr := fmt.Sprint(err)
			if strings.Contains(errStr, "invalid apiVersion \"client.authentication.k8s.io/v1alpha1\"") {
				return fmt.Errorf("kubectl failed with %s, this is usually due to your aws cli version being out of date", errStr)
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