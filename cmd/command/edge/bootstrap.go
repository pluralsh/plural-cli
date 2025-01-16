package edge

import (
	"context"
	"fmt"
	"time"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/samber/lo"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/wait"
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
	return p.DoInstallOperator(url, *cluster.CreateCluster.DeployToken, c.String("values"))
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
