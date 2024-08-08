package cd

import (
	"fmt"
	"github.com/pluralsh/plural-cli/pkg/common"

	"github.com/AlecAivazis/survey/v2"
	"github.com/samber/lo"
	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/stacks"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

func (p *Plural) cdStacks() cli.Command {
	return cli.Command{
		Name:        "stacks",
		Subcommands: p.cdStacksCommands(),
		Usage:       "manage CD stacks",
	}
}

func (p *Plural) cdStacksCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "gen-backend",
			Action: common.LatestVersion(p.handleGenerateBackend),
			Usage:  "generate '_override.tf' to configure a custom terraform backend",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "address", Usage: "terraform backend address", Required: false},
				cli.StringFlag{Name: "lock-address", Usage: "terraform backend lock address", Required: false},
				cli.StringFlag{Name: "unlock-address", Usage: "terraform backend unlock address", Required: false},
			},
		},
	}
}

func (p *Plural) handleGenerateBackend(_ *cli.Context) error {
	if !config.Exists() {
		return fmt.Errorf("plural config not found. Run 'plural cd login' to log in first")
	}

	cfg := config.Read()
	if len(cfg.Token) == 0 || len(cfg.Email) == 0 {
		return fmt.Errorf("not logged in. Run 'plural cd login' to log in first")
	}

	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	id := ""
	err := p.askStackID(&id)
	if err != nil {
		return err
	}

	stateUrls, err := stacks.GetTerraformStateUrls(p.ConsoleClient, id)
	if err != nil {
		return err
	}

	dir := ""
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the path to the directory where '_override.tf' file should be stored [defaults to current working dir]:",
		Default: ".",
	}, &dir); err != nil {
		return err
	}

	fileName, err := stacks.GenerateOverrideTemplate(&stacks.OverrideTemplateInput{
		Address:       lo.FromPtr(stateUrls.Address),
		LockAddress:   lo.FromPtr(stateUrls.Lock),
		UnlockAddress: lo.FromPtr(stateUrls.Unlock),
		Actor:         cfg.Email,
		DeployToken:   consoleToken,
	}, dir)
	if err != nil {
		return err
	}

	return git.AppendGitIgnore(dir, []string{fileName})
}

func (p *Plural) askStackID(id *string) (err error) {
	return survey.AskOne(
		&survey.Input{Message: "Enter the stack id:"},
		id,
		survey.WithValidator(survey.Required),
	)
}
