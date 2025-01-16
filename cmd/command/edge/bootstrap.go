package edge

import (
	"context"
	"fmt"
	"strings"
	"time"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/algorithms"
	"github.com/samber/lo"
	"github.com/urfave/cli"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/yaml"
)

func (p *Plural) handleEdgeBootstrap(c *cli.Context) error {
	machineID := c.String("machine-id")
	project := c.String("project")

	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	registrationAttributes, err := p.getClusterRegistrationAttributes(machineID, project)
	if err != nil {
		return err
	}

	_, err = p.ConsoleClient.CreateClusterRegistration(*registrationAttributes) // TODO: Handle the case when it already exists, i.e. after reboot.
	if err != nil {
		return err
	}

	var complete bool
	var registration *gqlclient.ClusterRegistrationFragment
	_ = wait.PollUntilContextCancel(context.Background(), 1*time.Minute, true, func(_ context.Context) (done bool, err error) { // TODO: Add backoff?
		complete, registration = p.ConsoleClient.IsClusterRegistrationComplete(machineID)
		return complete, nil
	})

	cluster, err := p.ConsoleClient.CreateCluster(p.getClusterAttributes(registration))
	//if err != nil {
	//	if errors.Like(err, "handle") && common.Affirm("Do you want to reinstall the deployment operator?", "PLURAL_INSTALL_AGENT_CONFIRM_IF_EXISTS") {
	//		handle := lo.ToPtr(attrs.Name)
	//		if attrs.Handle != nil {
	//			handle = attrs.Handle
	//		}
	//		return p.ReinstallOperator(c, nil, handle)
	//	}
	//
	//	return err
	//}

	if cluster.CreateCluster.DeployToken == nil {
		return fmt.Errorf("could not fetch deploy token from cluster")
	}

	url := p.ConsoleClient.ExtUrl()
	if agentUrl, err := p.ConsoleClient.AgentUrl(cluster.CreateCluster.ID); err == nil {
		url = agentUrl
	}

	utils.Highlight("installing agent on %s with url %s\n", c.String("name"), p.ConsoleClient.Url())
	return p.doInstallOperator(url, *cluster.CreateCluster.DeployToken, c.String("values"))
}

func (p *Plural) getClusterRegistrationAttributes(machineID, project string) (*gqlclient.ClusterRegistrationCreateAttributes, error) {
	attrs := gqlclient.ClusterRegistrationCreateAttributes{MachineID: machineID}

	if project != "" {
		p, err := p.ConsoleClient.GetProject(project)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, fmt.Errorf("cannot find %s project", project)
		}
		attrs.ProjectID = lo.ToPtr(p.ID)
	}

	return &attrs, nil
}

func (p *Plural) getClusterAttributes(registration *gqlclient.ClusterRegistrationFragment) gqlclient.ClusterAttributes {
	attributes := gqlclient.ClusterAttributes{
		Name:   registration.Name,
		Handle: &registration.Handle,
		// TODO: Tags: registration.Tags,
		// TODO: Metadata: registration.Metadata,
	}

	if registration.Project != nil {
		attributes.ProjectID = &registration.Project.ID
	}

	return attributes
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
