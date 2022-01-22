package main

import (
	"os"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/urfave/cli"
)

func shellCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "sync",
			Usage:     "syncs the setup in your cloud shell locally",
			Action:    handleShellSync,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "endpoint",
					Usage: "the endpoint for the plural installation you're working with",
				},
				cli.StringFlag{
					Name:  "service-account",
					Usage: "email for the service account you'd like to use for this workspace",
				},
			},
		},
		{
			Name:      "purge",
			Usage:     "deletes your cloud shell",
			Action:    handleShellPurge,
		},
	}
}

func handleShellSync(c *cli.Context) error {
	if !config.Exists() {
		if err := handleLogin(c); err != nil {
			return err
		}
	}
	client := api.NewClient()

	shell, err := client.GetShell()
	if err != nil {
		return err
	}

	if err := crypto.Setup(shell.AesKey); err != nil {
		return err
	}

	utils.Highlight("Cloning your workspace repo locally:\n")
	if err := utils.Exec("git", "clone", shell.GitUrl); err != nil {
		return err
	}

	dir := utils.RepoName(shell.GitUrl)
	os.Chdir(dir)
	if err := cryptoInit(c); err != nil {
		return err
	}

	return handleUnlock(c)
}

func handleShellPurge(c *cli.Context) error {
	client := api.NewClient()
	return client.DeleteShell()
}