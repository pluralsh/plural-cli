package main

import (
	"fmt"

	"github.com/pluralsh/plural/pkg/terraform"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
)

func terraformCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "init",
			Usage:  "command initializes a working directory containing Terraform configuration files",
			Action: terraformVersion(latestVersion(terraformInit)),
		},
	}
}

func terraformInit(c *cli.Context) error {
	_, found := utils.ProjectRoot()
	if !found {
		return fmt.Errorf("Project not initialized, run `plural init` to set up a workspace")
	}

	return terraform.Init()
}
