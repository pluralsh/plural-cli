package tui

import (
	"context"
	"os"

	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	accessbridge "github.com/pluralsh/plural-cli/pkg/bridge/access"
	clustersbridge "github.com/pluralsh/plural-cli/pkg/bridge/clusters"
	notificationsbridge "github.com/pluralsh/plural-cli/pkg/bridge/notifications"
	pipelinesbridge "github.com/pluralsh/plural-cli/pkg/bridge/pipelines"
	repositoriesbridge "github.com/pluralsh/plural-cli/pkg/bridge/repositories"
	servicesbridge "github.com/pluralsh/plural-cli/pkg/bridge/services"
	welcomebridge "github.com/pluralsh/plural-cli/pkg/bridge/welcome"
	"github.com/pluralsh/plural-cli/pkg/common"
	tuiapp "github.com/pluralsh/plural-cli/tui/app"
)

// Command is the only route that starts the full-screen terminal application.
func Command() cli.Command {
	return command(func(ctx context.Context) error {
		welcome := welcomebridge.NewService(welcomebridge.NewLocalSource(common.Version))
		auth := bridge.NewAuthService(bridge.PluralAuthFactory{}, 0)
		access := accessbridge.NewLocalManager("", auth, nil)
		services := servicesbridge.NewService(access)
		clusters := clustersbridge.NewService(access)
		repositories := repositoriesbridge.NewService(access)
		pipelines := pipelinesbridge.NewService(access)
		notifications := notificationsbridge.NewService(access)
		return tuiapp.Run(ctx, os.Stdin, os.Stdout, tuiapp.Dependencies{
			Welcome:       welcome,
			Access:        access,
			Services:      services,
			Clusters:      clusters,
			Repositories:  repositories,
			Pipelines:     pipelines,
			Notifications: notifications,
		})
	})
}

func command(run func(context.Context) error) cli.Command {
	return cli.Command{
		Name:  "tui",
		Usage: "opens the interactive terminal application",
		Action: func(*cli.Context) error {
			return run(context.Background())
		},
	}
}
