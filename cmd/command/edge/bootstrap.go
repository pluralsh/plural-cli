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

	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	utils.Highlight("registering new cluster on %s machine\n", machineID)
	if _, err := p.ConsoleClient.CreateClusterRegistration(gqlclient.ClusterRegistrationCreateAttributes{MachineID: machineID}); err != nil {
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

	utils.Highlight("creating %s cluster\n", lo.FromPtr(registration.Name))
	clusterAttributes, err := p.getClusterAttributes(registration)
	if err != nil {
		return err
	}

	cluster, err := p.ConsoleClient.CreateCluster(*clusterAttributes)
	if err != nil {
		return err
	}
	if cluster.CreateCluster.DeployToken == nil {
		return fmt.Errorf("could not fetch deploy token from cluster")
	}

	utils.Highlight("installing agent on %s cluster with %s URL\n", lo.FromPtr(registration.Name), p.ConsoleClient.Url())
	url := p.ConsoleClient.ExtUrl()
	if agentUrl, err := p.ConsoleClient.AgentUrl(cluster.CreateCluster.ID); err == nil {
		url = agentUrl
	}
	return p.DoInstallOperator(url, *cluster.CreateCluster.DeployToken, "")
}

func (p *Plural) getClusterAttributes(registration *gqlclient.ClusterRegistrationFragment) (*gqlclient.ClusterAttributes, error) {
	attributes := gqlclient.ClusterAttributes{
		Name:   lo.FromPtr(registration.Name),
		Handle: registration.Handle,
	}

	if registration.Tags != nil {
		attributes.Tags = lo.Map(registration.Tags, func(tag *gqlclient.ClusterTags, index int) *gqlclient.TagAttributes {
			if tag == nil {
				return nil
			}

			return &gqlclient.TagAttributes{Name: tag.Name, Value: tag.Value}
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
