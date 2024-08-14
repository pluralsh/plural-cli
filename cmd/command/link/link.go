package link

import (
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/urfave/cli"
)

func Commands() []cli.Command {
	return []cli.Command{
		{
			Name:      "link",
			Usage:     "links a local package into an installation repo",
			ArgsUsage: "TOOL REPO",
			Action:    handleLink,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Usage: "the name of the artifact to link",
				},
				cli.StringFlag{
					Name:  "path, f",
					Usage: "local path to that artifact (can be relative)",
				},
			},
		},
		{
			Name:      "unlink",
			Usage:     "unlinks a linked package",
			ArgsUsage: "REPO TOOL NAME",
			Action:    handleUnlink,
		},
	}
}

func handleLink(c *cli.Context) error {
	tool, repo := c.Args().Get(0), c.Args().Get(1)
	name, path := c.String("name"), c.String("path")

	if name == "" {
		name = filepath.Base(path)
	}

	manPath, err := manifest.ManifestPath(repo)
	if err != nil {
		return err
	}

	man, err := manifest.Read(manPath)
	if err != nil {
		return err
	}

	man.AddLink(tool, name, path)

	return man.Write(manPath)
}

func handleUnlink(c *cli.Context) error {
	repo, tool := c.Args().Get(0), c.Args().Get(1)

	manPath, err := manifest.ManifestPath(repo)
	if err != nil {
		return err
	}

	man, err := manifest.Read(manPath)
	if err != nil {
		return err
	}

	if tool == "all" {
		man.UnlinkAll()
	} else {
		man.Unlink(tool, c.Args().Get(2))
	}

	return man.Write(manPath)
}
