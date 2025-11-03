package stacks

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/stacks"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/samber/lo"
	"github.com/urfave/cli"
)

func init() {
	consoleToken = ""
	consoleURL = ""
}

var consoleToken string
var consoleURL string

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:        "stacks",
		Aliases:     []string{"s"},
		Usage:       "manage infrastructure stacks",
		Subcommands: p.stacksCommands(),
		Category:    "CD",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "token",
				Usage:       "console token",
				EnvVar:      "PLURAL_CONSOLE_TOKEN",
				Destination: &consoleToken,
			},
			cli.StringFlag{
				Name:        "url",
				Usage:       "console url address",
				EnvVar:      "PLURAL_CONSOLE_URL",
				Destination: &consoleURL,
			},
		},
	}
}

func (p *Plural) stacksCommands() []cli.Command {
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

	stackNames := make(map[string]string)
	infrastructureStacks, err := p.Plural.ConsoleClient.ListaStacks()
	if err != nil {
		return api.GetErrorResponse(err, "ListaStacks")
	}
	if infrastructureStacks == nil || infrastructureStacks.InfrastructureStacks == nil || len(infrastructureStacks.InfrastructureStacks.Edges) == 0 {
		return fmt.Errorf("returned objects list [ListStacks] is nil")
	}
	for _, node := range infrastructureStacks.InfrastructureStacks.Edges {
		stackNames[node.Node.Name] = lo.FromPtr(node.Node.ID)
	}
	var name string
	prompt := &survey.Select{
		Message: "Select a stack to generate a backend for:",
		Options: lo.Keys(stackNames),
	}
	opts := []survey.AskOpt{survey.WithValidator(survey.Required)}
	if err := survey.AskOne(prompt, &name, opts...); err != nil {
		return err
	}

	stateUrls, err := stacks.GetTerraformStateUrls(p.ConsoleClient, stackNames[name])
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
