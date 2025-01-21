package edge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

	utils.Highlight("registering new cluster on %s machine\n", machineID)
	if _, err = p.ConsoleClient.CreateClusterRegistration(*registrationAttributes); err != nil {
		if !strings.Contains(err.Error(), "machine_id has already been taken") {
			return err
		}
		utils.Highlight("cluster registration already exists\n")
	}

	utils.Highlight("waiting for registration to be completed\n")
	var complete bool
	var registration *gqlclient.ClusterRegistrationFragment
	_ = wait.PollUntilContextCancel(context.Background(), 30*time.Second, true, func(_ context.Context) (done bool, err error) {
		complete, registration = p.ConsoleClient.IsClusterRegistrationComplete(machineID)
		return complete, nil
	})

	clusterAttributes, err := p.getClusterAttributes(registration)
	if err != nil {
		return err
	}

	utils.Highlight("creating %s cluster\n", lo.FromPtr(registration.Name))
	cluster, err := p.ConsoleClient.CreateCluster(*clusterAttributes)
	if err != nil {
		if strings.Contains(err.Error(), "handle has already been taken") {
			handle := lo.ToPtr(clusterAttributes.Name)
			if clusterAttributes.Handle != nil {
				handle = clusterAttributes.Handle
			}
			return p.ReinstallOperator(c, nil, handle)
		}
		return err
	}

	if cluster.CreateCluster.DeployToken == nil {
		return fmt.Errorf("could not fetch deploy token from cluster")
	}

	url := p.ConsoleClient.ExtUrl()
	if agentUrl, err := p.ConsoleClient.AgentUrl(cluster.CreateCluster.ID); err == nil {
		url = agentUrl
	}

	utils.Highlight("installing agent on %s cluster with %s URL\n", registration.Name, p.ConsoleClient.Url())
	return p.DoInstallOperator(url, *cluster.CreateCluster.DeployToken, "")
}

func (p *Plural) getClusterRegistrationAttributes(machineID, project string) (*gqlclient.ClusterRegistrationCreateAttributes, error) {
	attributes := gqlclient.ClusterRegistrationCreateAttributes{MachineID: machineID}

	if project != "" {
		proj, err := p.ConsoleClient.GetProject(project)
		if err != nil {
			return nil, err
		}
		if proj == nil {
			return nil, fmt.Errorf("cannot find %s project", project)
		}
		attributes.ProjectID = lo.ToPtr(proj.ID)
	}

	return &attributes, nil
}

func (p *Plural) getClusterAttributes(registration *gqlclient.ClusterRegistrationFragment) (*gqlclient.ClusterAttributes, error) {
	attributes := gqlclient.ClusterAttributes{
		Handle: registration.Handle,
	}

	if registration.Name != nil {
		attributes.Name = *registration.Name
	}

	if registration.Tags != nil {
		attributes.Tags = lo.Map(registration.Tags, func(tag *gqlclient.ClusterTags, index int) *gqlclient.TagAttributes {
			if tag == nil {
				return nil
			}

			return &gqlclient.TagAttributes{
				Name:  tag.Name,
				Value: tag.Value,
			}
		})
	}

	if registration.Metadata != nil {
		metadata, err := json.Marshal(registration.Metadata)
		if err != nil {
			return nil, err
		}
		attributes.Metadata = lo.ToPtr(string(metadata))
	}

	if registration.Project != nil {
		attributes.ProjectID = &registration.Project.ID
	}

	return &attributes, nil
}
