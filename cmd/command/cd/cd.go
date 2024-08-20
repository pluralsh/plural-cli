package cd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/cd"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/algorithms"
	"github.com/urfave/cli"
	"helm.sh/helm/v3/pkg/action"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/yaml"
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
		p.cdStacks(),
		{
			Name:   "install",
			Action: p.handleInstallDeploymentsOperator,
			Usage:  "install deployments operator",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "console url", Required: true},
				cli.StringFlag{Name: "token", Usage: "deployment token", Required: true},
				cli.StringFlag{Name: "values", Usage: "values file to use for the deployment agent helm chart", Required: false},
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
			Action: p.handleCdLogin,
			Usage:  "logs into your plural console",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "console url"},
				cli.StringFlag{Name: "token", Usage: "console access token"},
			},
		},
		{
			Name:      "eject",
			Action:    p.handleEject,
			Usage:     "ejects cluster scaffolds",
			ArgsUsage: "<cluster-id>",
			// TODO: enable once logic is finished
			Hidden: true,
		},
	}
}

func (p *Plural) handleInstallDeploymentsOperator(c *cli.Context) error {
	if !c.Bool("force") {
		confirm, err := confirmCluster(c.String("url"), c.String("token"))
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}
	}

	// we don't care if this fails to init as this command can be auth-less
	_ = p.InitConsoleClient(consoleToken, consoleURL)

	return p.doInstallOperator(c.String("url"), c.String("token"), c.String("values"))
}

func (p *Plural) handleUninstallOperator(_ *cli.Context) error {
	err := p.InitKube()
	if err != nil {
		return err
	}
	return console.UninstallAgent(console.OperatorNamespace)
}

func (p *Plural) doInstallOperator(url, token, values string) error {
	err := p.InitKube()
	if err != nil {
		return err
	}
	alreadyExists, err := console.IsAlreadyAgentInstalled(p.Kube.GetClient())
	if err != nil {
		return err
	}
	if alreadyExists && !common.Confirm("the deployment operator is already installed. Do you want to replace it", "PLURAL_INSTALL_AGENT_CONFIRM_IF_EXISTS") {
		utils.Success("deployment operator is already installed, skip installation\n")
		return nil
	}

	err = p.Kube.CreateNamespace(console.OperatorNamespace, false)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	vals := map[string]interface{}{}
	globalVals := map[string]interface{}{}
	version := ""

	if p.ConsoleClient != nil {
		settings, err := p.ConsoleClient.GetGlobalSettings()
		if err == nil && settings != nil {
			version = strings.Trim(settings.AgentVsn, "v")
			if settings.AgentHelmValues != nil {
				if err := yaml.Unmarshal([]byte(*settings.AgentHelmValues), &globalVals); err != nil {
					return err
				}
			}
		}
	}

	if values != "" {
		if err := utils.YamlFile(values, &vals); err != nil {
			return err
		}
	}
	vals = algorithms.Merge(vals, globalVals)
	err = console.InstallAgent(url, token, console.OperatorNamespace, version, vals)
	if err == nil {
		utils.Success("deployment operator installed successfully\n")
	}
	return err
}

func confirmCluster(url, token string) (bool, error) {
	consoleClient, err := console.NewConsoleClient(token, url)
	if err != nil {
		return false, err
	}

	myCluster, err := consoleClient.MyCluster()
	if err != nil {
		return false, err
	}

	clusterFragment, err := consoleClient.GetCluster(&myCluster.MyCluster.ID, nil)
	if err != nil {
		return false, err
	}

	handle := "-"
	provider := "-"
	if clusterFragment.Handle != nil {
		handle = *clusterFragment.Handle
	}
	if clusterFragment.Provider != nil {
		provider = clusterFragment.Provider.Name
	}
	return common.Confirm(fmt.Sprintf("Are you sure you want to install deploy operator for the cluster:\nName: %s\nHandle: %s\nProvider: %s\n", myCluster.MyCluster.Name, handle, provider), "PLURAL_INSTALL_AGENT_CONFIRM"), nil
}

func (p *Plural) handleCdLogin(c *cli.Context) (err error) {
	url := c.String("url")
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
	if !c.Args().Present() {
		return fmt.Errorf("clusterid cannot be empty")
	}

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
