package edge

import (
	"context"

	pkgedge "github.com/pluralsh/plural-cli/pkg/edge"
	"github.com/urfave/cli"
)

func (p *Plural) handleEdgeBootstrap(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	return pkgedge.Bootstrap(context.Background(), p.ConsoleClient, func(url, token, chartLoc, clusterID string) error {
		return p.DoInstallOperator(url, token, "", chartLoc, clusterID)
	}, pkgedge.BootstrapOptions{
		MachineID: c.String("machine-id"),
		ChartLoc:  c.String("chart-loc"),
	}, nil)
}
