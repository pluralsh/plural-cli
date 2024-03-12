package plural

import (
	"github.com/AlecAivazis/survey/v2"
	consoleclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/algorithms"
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

func (s *Plural) listNotifications() *algorithms.Pager[*consoleclient.NotificationSinkEdgeFragment] {
	fetch := func(page *string, size int64) ([]*consoleclient.NotificationSinkEdgeFragment, *algorithms.PageInfo, error) {
		resp, err := s.ConsoleClient.ListNotificationSinks(page, &size)
		if err != nil {
			return nil, nil, err
		}
		pageInfo := &algorithms.PageInfo{
			HasNext:  resp.PageInfo.HasNextPage,
			After:    resp.PageInfo.EndCursor,
			PageSize: size,
		}
		return resp.Edges, pageInfo, nil
	}
	return algorithms.NewPager[*consoleclient.NotificationSinkEdgeFragment](defaultPageSize, fetch)
}

func (p *Plural) handleListNotificationSinks(_ *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	result := make([]*consoleclient.NotificationSinkFragment, 0)

	pager := p.listNotifications()

	for pager.HasNext() {
		resp, err := pager.NextPage()
		if err != nil {
			return err
		}

		for _, n := range resp {
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
