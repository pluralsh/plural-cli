package shell

import (
	"os"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/config"
	pkgcrypto "github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/urfave/cli"
)

func Command() cli.Command {
	return cli.Command{
		Name:        "shell",
		Usage:       "manages your cloud shell",
		Subcommands: shellCommands(),
		Category:    "Workspace",
	}
}

func shellCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "sync",
			Usage:  "syncs the setup in your cloud shell locally",
			Action: common.LatestVersion(handleShellSync),
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
			Name:   "purge",
			Usage:  "deletes your cloud shell",
			Action: common.LatestVersion(handleShellPurge),
		},
	}
}

func handleShellSync(c *cli.Context) error {
	if !config.Exists() {
		if err := common.HandleLogin(c); err != nil {
			return err
		}
	}
	client := api.NewClient()

	shell, err := client.GetShell()
	if err != nil {
		return api.GetErrorResponse(err, "GetShell")
	}

	if err := pkgcrypto.Setup(shell.AesKey); err != nil {
		return err
	}

	utils.Highlight("Cloning your workspace repo locally:\n")
	if err := utils.Exec("git", "clone", shell.GitUrl); err != nil {
		return err
	}

	dir := git.RepoName(shell.GitUrl)
	if err := os.Chdir(dir); err != nil {
		return err
	}
	if err := common.CryptoInit(c); err != nil {
		return err
	}

	return common.HandleUnlock(c)
}

var destoryShellConfirm = "Are you sure you want to destroy your cloud shell (you should either `plural destroy` anything deployed or `plural shell sync` to sync the contents locally)?"

func handleShellPurge(c *cli.Context) error {
	if ok := common.Confirm(destoryShellConfirm, "PLURAL_SHELL_PURGE_CONFIRM"); !ok {
		return nil
	}

	client := api.NewClient()
	err := client.DeleteShell()
	return api.GetErrorResponse(err, "DeleteShell")
}
