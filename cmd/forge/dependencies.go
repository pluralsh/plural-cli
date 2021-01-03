package main

import (
	"fmt"
	"github.com/michaeljguarino/forge/pkg/api"
	"github.com/michaeljguarino/forge/pkg/wkspace"
	"github.com/urfave/cli"
)

func topsort(c *cli.Context) error {
	client := api.NewClient()
	installations, _ := client.GetInstallations()
	repoName := c.Args().Get(0)
	sorted, err := wkspace.Dependencies(repoName, installations)
	if err != nil {
		return err
	}

	for _, inst := range sorted {
		fmt.Println(inst.Repository.Name)
	}
	return nil
}
