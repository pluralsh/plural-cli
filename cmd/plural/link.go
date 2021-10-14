package main

import (
	"strings"

	"github.com/urfave/cli"
	"github.com/pluralsh/plural/pkg/manifest"
)

func linkCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "link",
			Usage:     "links a local package into an installation repo",
			ArgsUsage: "TOOL REPO NAME:PATH",
			Action:    handleLink,
		},
		{
			Name:      "unlink",
			Usage:     "unlinks a linked package",
			ArgsUsage: "REPO TOOL:NAME",
			Action:    handleUnlink,
		},
	}
}

func handleLink(c *cli.Context) error {
	tool, repo, spec := c.Args().Get(0), c.Args().Get(1), c.Args().Get(2)
	parsed := strings.Split(spec, ":")

	manPath, err := manifest.ManifestPath(repo)
	if err != nil {
		return err
	}

	man, err := manifest.Read(manPath)
	if err != nil {
		return err
	}

	man.AddLink(tool, parsed[0], parsed[1])

	return man.Write(manPath)
}

func handleUnlink(c *cli.Context) error {
	repo, spec := c.Args().Get(0), c.Args().Get(1)

	manPath, err := manifest.ManifestPath(repo)
	if err != nil {
		return err
	}

	man, err := manifest.Read(manPath)
	if err != nil {
		return err
	}

	if spec == "all" {
		man.UnlinkAll()
	} else {
		parsed := strings.Split(spec, ":")
		man.Unlink(parsed[0], parsed[1])
	}

	return man.Write(manPath)
}