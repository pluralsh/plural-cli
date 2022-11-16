package main

import (
	"path/filepath"

	"github.com/pluralsh/plural/pkg/output"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/urfave/cli/v2"
)

func outputCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:      "terraform",
			Usage:     "generates terraform output",
			ArgsUsage: "REPO",
			Action:    latestVersion(handleTerraformOutput),
		},
	}
}

func outputPath(root, app string) string {
	return pathing.SanitizeFilepath(filepath.Join(root, app, "output.yaml"))
}

func handleTerraformOutput(c *cli.Context) (err error) {
	root, _ := utils.ProjectRoot()
	app := c.Args().Get(0)
	path := outputPath(root, app)
	out, err := output.Read(path)
	if err != nil {
		out = output.New()
	}

	tfOut, err := output.TerraformOutput(pathing.SanitizeFilepath(filepath.Join(root, app, "terraform")))
	if err != nil {
		return
	}

	out.Terraform = tfOut
	err = out.Save(app, path)
	return
}
