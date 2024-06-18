package plural

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
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
			Action: latestVersion(p.handleGenerateBackend),
			Usage:  "generate '_override.tf' to configure a custom terraform backend",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "address", Usage: "terraform backend address", Required: false},
				cli.StringFlag{Name: "lock-address", Usage: "terraform backend lock address", Required: false},
				cli.StringFlag{Name: "unlock-address", Usage: "terraform backend unlock address", Required: false},
			},
		},
	}
}

var backendSurvey = []*survey.Question{
	{
		Name:     "address",
		Prompt:   &survey.Input{Message: "Enter the address of terraform backend:"},
		Validate: survey.Required,
	},
	{
		Name:     "lockAddress",
		Prompt:   &survey.Input{Message: "Enter the lock address of terraform backend:"},
		Validate: survey.Required,
	},
	{
		Name:     "unlockAddress",
		Prompt:   &survey.Input{Message: "Enter the unlock address of terraform backend:"},
		Validate: survey.Required,
	},
	{
		Name: "dir",
		Prompt: &survey.Input{
			Message: "Enter the path to the directory where '_override.tf' file should be stored [defaults to current working dir]:",
			Default: ".",
		},
		Validate: survey.Required,
	},
}

func (p *Plural) handleGenerateBackend(_ *cli.Context) error {
	if !config.Exists() {
		return fmt.Errorf("plural config not found. Run 'plural cd login' to log in first")
	}

	cfg := config.Read()
	if len(cfg.Token) == 0 || len(cfg.Email) == 0 {
		return fmt.Errorf("not logged in. Run 'plural cd login' to log in first")
	}

	var resp struct {
		Address       string
		LockAddress   string
		UnlockAddress string
		Dir           string
	}

	if err := survey.Ask(backendSurvey, &resp); err != nil {
		return err
	}

	fileName, err := stacks.GenerateOverrideTemplate(&stacks.OverrideTemplateInput{
		Address:       resp.Address,
		LockAddress:   resp.LockAddress,
		UnlockAddress: resp.UnlockAddress,
		Actor:         cfg.Email,
		DeployToken:   cfg.Token,
	}, resp.Dir)
	if err != nil {
		return err
	}

	return git.AppendGitIgnore(resp.Dir, []string{fileName})
}
