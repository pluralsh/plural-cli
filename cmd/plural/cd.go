package plural

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/pluralsh/plural-cli/pkg/cd"
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

func (p *Plural) cdCommands() []cli.Command {
	return []cli.Command{
		p.cdProviders(),
		p.cdCredentials(),
		p.cdClusters(),
		p.cdServices(),
		p.cdRepositories(),
		p.cdPipelines(),
		p.cdPipelineGates(),
		{
			Name:   "install",
			Action: p.handleInstallDeploymentsOperator,
			Usage:  "install deployments operator",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "console url", Required: true},
				cli.StringFlag{Name: "token", Usage: "deployment token", Required: true},
				cli.BoolFlag{Name: "force", Usage: "ignore checking if the current cluster is correct"},
			},
		},
		{
			Name:   "control-plane",
			Action: p.handleInstallControlPlane,
			Usage:  "sets up the plural console in an existing k8s cluster",
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
				cli.StringFlag{Name: "url", Usage: "console url", Required: true},
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

	return p.doInstallOperator(c.String("url"), c.String("token"))
}

func (p *Plural) handleUninstallOperator(_ *cli.Context) error {
	err := p.InitKube()
	if err != nil {
		return err
	}
	return console.UninstallAgent(console.OperatorNamespace)
}

func (p *Plural) doInstallOperator(url, token string) error {
	err := p.InitKube()
	if err != nil {
		return err
	}
	err = p.Kube.CreateNamespace(console.OperatorNamespace)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	err = console.InstallAgent(url, token, console.OperatorNamespace)
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
	return confirm(fmt.Sprintf("Are you sure you want to install deploy operator for the cluster:\nName: %s\nHandle: %s\nProvider: %s\n", myCluster.MyCluster.Name, handle, provider), "PLURAL_INSTALL_AGENT_CONFIRM"), nil
}

func (p *Plural) handleCdLogin(c *cli.Context) (err error) {
	url := c.String("url")
	token := c.String("token")
	if token == "" {
		token, err = utils.ReadPwd("Enter your console access token")
		if err != nil {
			return
		}
	}
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
