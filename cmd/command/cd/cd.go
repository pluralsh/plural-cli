package cd

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
	"helm.sh/helm/v3/pkg/action"

	"github.com/pluralsh/plural-cli/pkg/cd"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

func init() {
	consoleToken = ""
	consoleURL = ""
}

var consoleToken string
var consoleURL string

type Plural struct {
	client.Plural
	HelmConfiguration *action.Configuration
}

func SetConsoleURL(url string) {
	consoleURL = url
}

func Command(clients client.Plural, helmConfiguration *action.Configuration) cli.Command {
	return cli.Command{
		Name:        "deployments",
		Aliases:     []string{"cd"},
		Usage:       "view and manage plural deployments",
		Subcommands: Commands(clients, helmConfiguration),
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
		Category: "CD",
	}
}

func Commands(clients client.Plural, helmConfiguration *action.Configuration) []cli.Command {
	p := Plural{
		HelmConfiguration: helmConfiguration,
		Plural:            clients,
	}
	return []cli.Command{
		p.cdProviders(),
		p.cdCredentials(),
		p.cdClusters(),
		p.cdServices(),
		p.cdContexts(),
		p.cdRepositories(),
		p.cdPipelines(),
		p.cdNotifications(),
		p.cdSettings(),
		{
			Name:   "install",
			Action: p.handleInstallDeploymentsOperator,
			Usage:  "install deployments operator",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "console url", Required: true},
				cli.StringFlag{Name: "token", Usage: "deployment token", Required: true},
				cli.StringFlag{Name: "cluster-id", Usage: "cluster id to install the operator for", Required: false},
				cli.StringFlag{Name: "values", Usage: "values file to use for the deployment agent helm chart", Required: false},
				cli.StringFlag{Name: "chart-loc", Usage: "URL or filepath of helm chart tar file. Use if not wanting to install helm chart from default plural repository.", Required: false},
				cli.BoolFlag{Name: "force", Usage: "ignore checking if the current cluster is correct"},
			},
		},
		{
			Name:    "control-plane",
			Aliases: []string{"helm-values"},
			Action:  p.handleInstallControlPlane,
			Usage:   "sets up the plural console in an existing k8s cluster",
		},
		{
			Name:   "control-plane-values",
			Action: p.handlePrintControlPlaneValues,
			Usage:  "dumps a values file for installing the plural console",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "domain", Usage: "The plural domain to use for this console", Required: true},
				cli.StringFlag{Name: "dsn", Usage: "The Postgres DSN to use for database connections", Required: true},
				cli.StringFlag{Name: "name", Usage: "The name given to the cluster", Required: true},
				cli.StringFlag{Name: "file", Usage: "The file to dump values to", Required: true},
			},
		},
		{
			Name:   "uninstall",
			Action: p.handleUninstallOperator,
			Usage:  "uninstalls the deployment operator from the current cluster",
		},
		{
			Name:   "login",
			Action: p.HandleCdLogin,
			Usage:  "logs into your plural console",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "console url"},
				cli.StringFlag{Name: "token", Usage: "console access token"},
			},
		},
		{
			Name:      "eject",
			Action:    common.RequireArgs(p.handleEject, []string{"{cluster-id}"}),
			Usage:     "ejects cluster scaffolds",
			ArgsUsage: "{cluster-id}",
			Hidden:    true, // TODO: enable once logic is finished
		},
	}
}

func (p *Plural) handleInstallDeploymentsOperator(c *cli.Context) error {
	cliClusterId := c.String("cluster-id")
	if !c.Bool("force") {
		confirm, clusterId, err := confirmCluster(c.String("url"), c.String("token"))
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}
		cliClusterId = clusterId
	}

	if cliClusterId == "" {
		return fmt.Errorf("cluster id must be provided")
	}

	// we don't care if this fails to init as this command can be auth-less
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		utils.Warn("Console client was not initialized, reason: %s", err.Error())
	}

	return p.DoInstallOperator(c.String("url"), c.String("token"), c.String("values"), c.String("chart-loc"), cliClusterId)
}

func (p *Plural) handleUninstallOperator(_ *cli.Context) error {
	err := p.InitKube()
	if err != nil {
		return err
	}
	return console.UninstallAgent(console.OperatorNamespace)
}

func confirmCluster(url, token string) (bool, string, error) {
	consoleClient, err := console.NewConsoleClient(token, url)
	if err != nil {
		return false, "", err
	}

	myCluster, err := consoleClient.MyCluster()
	if err != nil {
		return false, "", err
	}

	clusterFragment, err := consoleClient.GetCluster(&myCluster.MyCluster.ID, nil)
	if err != nil {
		return false, "", err
	}

	handle := "-"
	provider := "-"
	if clusterFragment.Handle != nil {
		handle = *clusterFragment.Handle
	}
	if clusterFragment.Distro != nil {
		provider = string(*clusterFragment.Distro)
	}
	return common.Confirm(fmt.Sprintf("Are you sure you want to install deploy operator for the cluster:\nName: %s\nHandle: %s\nProvider: %s\n", myCluster.MyCluster.Name, handle, provider), "PLURAL_INSTALL_AGENT_CONFIRM"), myCluster.MyCluster.ID, nil
}

func (p *Plural) HandleCdLogin(c *cli.Context) (err error) {
	prior := console.ReadConfig()
	if prior.Url != "" {
		if !strings.EqualFold(common.GetHostnameFromURL(prior.Url), common.GetHostnameFromURL(consoleURL)) {
			if common.Affirm(
				fmt.Sprintf("You've already configured your console at %s, continue using those credentials?", prior.Url),
				"PLURAL_CD_USE_EXISTING_CREDENTIALS",
			) {
				return
			}
		}
	}

	url := consoleURL
	if url == "" {
		url = c.String("url")
	}
	if url == "" {
		url, err = utils.ReadLine("Enter the url of your console: ")
		if err != nil {
			return
		}
	}

	token := c.String("token")
	if token == "" {
		token, err = utils.ReadPwd("Enter your console access token: ")
		if err != nil {
			return
		}
	}

	url = console.NormalizeUrl(url)
	conf := console.Config{Url: url, Token: token}
	return conf.Save()
}

func (p *Plural) handleInstallControlPlane(_ *cli.Context) error {
	conf := config.Read()
	vals, err := cd.CreateControlPlane(conf)
	if err != nil {
		return err
	}

	fmt.Print("\n\n")
	utils.Highlight("===> writing values.secret.yaml, you should keep this in a secure location for future helm upgrades\n\n")
	if err := os.WriteFile("values.secret.yaml", []byte(vals), 0644); err != nil {
		return err
	}

	fmt.Println("After confirming everything looks correct in values.secret.yaml, run the following command to install:")
	utils.Highlight("helm repo add plrl-console https://pluralsh.github.io/console\n")
	utils.Highlight("helm upgrade --install --create-namespace -f values.secret.yaml console plrl-console/console -n plrl-console\n")
	return nil
}

func (p *Plural) handlePrintControlPlaneValues(c *cli.Context) error {
	p.InitPluralClient()
	conf := config.Read()
	vals, err := cd.ControlPlaneValues(conf, c.String("file"), c.String("domain"), c.String("dsn"), c.String("name"))
	if err != nil {
		return err
	}

	return os.WriteFile(c.String("file"), []byte(vals), 0644)
}

func (p *Plural) handleEject(c *cli.Context) (err error) {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	clusterId := c.Args().First()
	cluster, err := p.ConsoleClient.GetCluster(&clusterId, nil)
	if err != nil {
		return err
	}

	if cluster == nil {
		return fmt.Errorf("could not find cluster with given id")
	}

	return cd.Eject(cluster)
}
