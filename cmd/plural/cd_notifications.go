package plural

import (
	"github.com/AlecAivazis/survey/v2"
	consoleclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/samber/lo"
	"github.com/urfave/cli"
)

const defaultPageSize = 100

func (p *Plural) cdNotifications() cli.Command {
	return cli.Command{
		Name: "notifications",
		Subcommands: []cli.Command{
			{
				Name:        "sinks",
				Subcommands: p.cdNotificationSinkCommands(),
				Usage:       "manage CD notification sinks",
			},
		},
		Usage: "manage CD notifications",
	}
}

func (p *Plural) cdNotificationSinkCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "list",
			Action: latestVersion(p.handleListNotificationSinks),
			Usage:  "list notification sinks",
		},
		{
			Name:      "upsert",
			ArgsUsage: "NAME",
			Action:    latestVersion(requireArgs(p.handleCreateNotificationSinks, []string{"NAME"})),
			Usage:     "upsert notification sink",
		},
	}
}

func (p *Plural) handleCreateNotificationSinks(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	sinkType := ""
	name := c.Args().Get(0)
	prompt := &survey.Select{
		Message: "Select one of the following type:",
		Options: []string{consoleclient.SinkTypeSLACk.String(), consoleclient.SinkTypeTeams.String()},
	}
	if err := survey.AskOne(prompt, &sinkType, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	url := ""
	if err := survey.AskOne(&survey.Input{Message: "Enter an URL address"}, &url); err != nil {
		return err
	}

	configuration := consoleclient.SinkConfigurationAttributes{}
	urlSinkAttributes := &consoleclient.URLSinkAttributes{
		URL: url,
	}
	if consoleclient.SinkTypeSLACk == consoleclient.SinkType(sinkType) {
		configuration.Slack = urlSinkAttributes
	} else {
		configuration.Teams = urlSinkAttributes
	}

	attr := consoleclient.NotificationSinkAttributes{
		Name:          name,
		Type:          consoleclient.SinkType(sinkType),
		Configuration: configuration,
	}
	result, err := p.ConsoleClient.CreateNotificationSinks(attr)
	if err != nil {
		return err
	}

	headers := []string{"Id", "Name", "Type", "URL"}
	return utils.PrintTable([]consoleclient.NotificationSinkFragment{*result}, headers, func(ns consoleclient.NotificationSinkFragment) ([]string, error) {
		url := ""
		if ns.Configuration.Teams != nil {
			url = ns.Configuration.Teams.URL
		}
		if ns.Configuration.Slack != nil {
			url = ns.Configuration.Slack.URL
		}
		return []string{ns.ID, ns.Name, ns.Type.String(), url}, nil
	})
}

func (p *Plural) handleListNotificationSinks(_ *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	result := make([]*consoleclient.NotificationSinkFragment, 0)
	var after *string
	hasNextPage := true
	for hasNextPage {
		resp, err := p.ConsoleClient.ListNotificationSinks(after, lo.ToPtr(int64(defaultPageSize)))
		if err != nil {
			return err
		}

		hasNextPage = resp.PageInfo.HasNextPage
		after = resp.PageInfo.EndCursor

		for _, n := range resp.Edges {
			result = append(result, n.Node)
		}
	}

	headers := []string{"Id", "Name", "Type", "URL"}
	return utils.PrintTable(result, headers, func(ns *consoleclient.NotificationSinkFragment) ([]string, error) {
		url := ""
		if ns.Configuration.Teams != nil {
			url = ns.Configuration.Teams.URL
		}
		if ns.Configuration.Slack != nil {
			url = ns.Configuration.Slack.URL
		}
		return []string{ns.ID, ns.Name, ns.Type.String(), url}, nil
	})
}
