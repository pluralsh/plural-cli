package plural

import (
	"fmt"
	"os"

	"github.com/pluralsh/plural/pkg/cd"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/console"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func init() {
	consoleToken = ""
	consoleURL = ""
}

const (
	operatorNamespace = "plrl-deploy-operator"
)

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
		{
			Name:   "install",
			Action: p.handleInstallDeploymentsOperator,
			Usage:  "install deployments operator",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "console url", Required: true},
				cli.StringFlag{Name: "token", Usage: "deployment token", Required: true},
			},
		},
		{
			Name:   "control-plane",
			Action: p.handleInstallControlPlane,
			Usage:  "sets up the plural console in an existing k8s cluster",
		},
		{
			Name:   "uninstall",
			Action: p.handleUninstallOperator,
			Usage:  "uninstalls the deployment operator from the current cluster",
		},
		{
			Name:   "login",
			Action: handleCdLogin,
			Usage:  "logs into your plural console",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "console url", Required: true},
				cli.StringFlag{Name: "token", Usage: "console access token"},
			},
		},
	}
}

func (p *Plural) handleInstallDeploymentsOperator(c *cli.Context) error {
	return p.doInstallOperator(c.String("url"), c.String("token"))
}

func (p *Plural) handleUninstallOperator(_ *cli.Context) error {
	err := p.InitKube()
	if err != nil {
		return err
	}
	return console.UninstallAgent(operatorNamespace)
}

func (p *Plural) doInstallOperator(url, token string) error {
	err := p.InitKube()
	if err != nil {
		return err
	}
	err = p.Kube.CreateNamespace(operatorNamespace)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	err = console.InstallAgent(url, token, operatorNamespace)
	if err == nil {
		utils.Success("deployment operator installed successfully\n")
	}
	return err
}

func handleCdLogin(c *cli.Context) (err error) {
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
	utils.Highlight("helm upgrade --install --create-namespace -f values.secret.yaml console -n plrl-console")
	return nil
}
