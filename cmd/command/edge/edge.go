package edge

import (
	"fmt"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/console/errors"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/algorithms"
	"github.com/samber/lo"
	"github.com/urfave/cli"
	"helm.sh/helm/v3/pkg/action"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/yaml"
)

var consoleToken string
var consoleURL string

type Plural struct {
	client.Plural
	HelmConfiguration *action.Configuration
}

func init() {
	consoleToken = ""
	consoleURL = ""
}

func Command(clients client.Plural, helmConfiguration *action.Configuration) cli.Command {
	return cli.Command{
		Name:        "edge",
		Usage:       "manage edge clusters",
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
		Category: "Edge",
	}
}

func Commands(clients client.Plural, helmConfiguration *action.Configuration) []cli.Command {
	p := Plural{
		HelmConfiguration: helmConfiguration,
		Plural:            clients,
	}
	return []cli.Command{
		{
			Name:   "bootstrap",
			Action: p.handleClusterBootstrap,
			Usage:  "registers edge cluster and installs agent onto it using the current kubeconfig",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "machine-id",
					Usage:    "the unique id of the edge device on which this cluster runs",
					Required: true,
				},
				cli.StringFlag{
					Name:     "project",
					Usage:    "the project this cluster will belong to",
					Required: false, // TODO: It can be inferred from bootstrap token.
				},
			},
		},
	}
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

func (p *Plural) ReinstallOperator(c *cli.Context, id, handle *string) error {
	deployToken, err := p.ConsoleClient.GetDeployToken(id, handle)
	if err != nil {
		return err
	}

	url := p.ConsoleClient.ExtUrl()
	if cluster, err := p.ConsoleClient.GetCluster(id, handle); err == nil {
		if agentUrl, err := p.ConsoleClient.AgentUrl(cluster.ID); err == nil {
			url = agentUrl
		}
	}

	return p.doInstallOperator(url, deployToken, c.String("values"))
}

func (p *Plural) handleClusterBootstrap(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	attrs := gqlclient.ClusterAttributes{Name: c.String("name")}
	if c.String("handle") != "" {
		attrs.Handle = lo.ToPtr(c.String("handle"))
	}

	if c.String("project") != "" {
		project, err := p.ConsoleClient.GetProject(c.String("project"))
		if err != nil {
			return nil
		}
		if project == nil {
			return fmt.Errorf("Could not find project %s", c.String("project"))
		}

		attrs.ProjectID = lo.ToPtr(project.ID)
	}

	if c.IsSet("tag") {
		attrs.Tags = lo.Map(c.StringSlice("tag"), func(tag string, index int) *gqlclient.TagAttributes {
			tags := strings.Split(tag, "=")
			if len(tags) == 2 {
				return &gqlclient.TagAttributes{
					Name:  tags[0],
					Value: tags[1],
				}
			}
			return nil
		})
		attrs.Tags = lo.Filter(attrs.Tags, func(t *gqlclient.TagAttributes, ind int) bool { return t != nil })
	}

	existing, err := p.ConsoleClient.CreateCluster(attrs)
	if err != nil {
		if errors.Like(err, "handle") && common.Affirm("Do you want to reinstall the deployment operator?", "PLURAL_INSTALL_AGENT_CONFIRM_IF_EXISTS") {
			handle := lo.ToPtr(attrs.Name)
			if attrs.Handle != nil {
				handle = attrs.Handle
			}
			return p.ReinstallOperator(c, nil, handle)
		}

		return err
	}

	if existing.CreateCluster.DeployToken == nil {
		return fmt.Errorf("could not fetch deploy token from cluster")
	}

	url := p.ConsoleClient.ExtUrl()
	if agentUrl, err := p.ConsoleClient.AgentUrl(existing.CreateCluster.ID); err == nil {
		url = agentUrl
	}

	deployToken := *existing.CreateCluster.DeployToken
	utils.Highlight("installing agent on %s with url %s\n", c.String("name"), p.ConsoleClient.Url())
	return p.doInstallOperator(url, deployToken, c.String("values"))
}
