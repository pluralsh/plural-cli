package main

import (
	"fmt"
	"path/filepath"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
)

func opsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "terminate",
			Usage:     "terminates a worker node in your cluster",
			ArgsUsage: "NAME",
			Action:    handleTerminateNode,
		},
	}
}

func handleTerminateNode(c *cli.Context) error {
	name := c.Args().Get(0)
	provider, err := getProvider()
	if err != nil {
		return err
	}

	kube, err := utils.Kubernetes()
	if err != nil {
		return err
	}

	node, err := kube.Node(name)
	if err != nil {
		return err
	}

	return provider.Decommision(node)
}

func getProvider() (provider.Provider, error) {
	root, found := utils.ProjectRoot()
	if !found {
		return nil, fmt.Errorf("Project not initialized, run `plural init` to set up a workspace")
	}

	manifestPath := filepath.Join(root, "manifest.yaml")
	return provider.Bootstrap(manifestPath, true)
}