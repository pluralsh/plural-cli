package edge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

const defaultRegistrationPollInterval = 30 * time.Second

const alreadyRegistered = "machine_id has already been taken"

// BootstrapOptions are the CLI flags for plural edge bootstrap.
type BootstrapOptions struct {
	MachineID    string
	ChartLoc     string
	PollInterval time.Duration
}

// BootstrapAPI is the Console subset used to register an edge cluster.
type BootstrapAPI interface {
	CreateClusterRegistration(gqlclient.ClusterRegistrationCreateAttributes) (*gqlclient.ClusterRegistrationFragment, error)
	IsClusterRegistrationComplete(machineID string) (bool, *gqlclient.ClusterRegistrationFragment)
	CreateCluster(attributes gqlclient.ClusterAttributes) (*gqlclient.CreateCluster, error)
	Url() string
	ExtUrl() string
	AgentUrl(id string) (string, error)
}

// InstallOperatorFunc installs the deployment operator into the current kubeconfig.
type InstallOperatorFunc func(url, token, chartLoc, clusterID string) error

// Bootstrap registers an edge machine in Console, creates the cluster, and installs the agent.
func Bootstrap(ctx context.Context, client BootstrapAPI, install InstallOperatorFunc, opts BootstrapOptions, log func(string)) error {
	if ctx == nil {
		ctx = context.Background()
	}
	machineID := strings.TrimSpace(opts.MachineID)
	if machineID == "" {
		return fmt.Errorf("machine-id is required")
	}
	if client == nil {
		return fmt.Errorf("console client is not configured")
	}
	if install == nil {
		return fmt.Errorf("agent installer is not configured")
	}

	logf := bootstrapLog(log)
	logf("registering new cluster on %s machine", machineID)
	if _, err := client.CreateClusterRegistration(gqlclient.ClusterRegistrationCreateAttributes{MachineID: machineID}); err != nil {
		if !strings.Contains(err.Error(), alreadyRegistered) {
			return err
		}
		logf("cluster registration already exists")
	}

	interval := opts.PollInterval
	if interval <= 0 {
		interval = defaultRegistrationPollInterval
	}

	logf("waiting for registration to be completed")
	var complete bool
	var registration *gqlclient.ClusterRegistrationFragment
	if err := wait.PollUntilContextCancel(ctx, interval, true, func(context.Context) (bool, error) {
		complete, registration = client.IsClusterRegistrationComplete(machineID)
		return complete, nil
	}); err != nil {
		return err
	}
	if !complete || registration == nil || registration.Name == nil {
		return fmt.Errorf("cluster registration was not completed")
	}

	logf("creating %s cluster", lo.FromPtr(registration.Name))
	attributes, err := clusterAttributes(registration)
	if err != nil {
		return err
	}
	cluster, err := client.CreateCluster(*attributes)
	if err != nil {
		return err
	}
	if cluster == nil || cluster.CreateCluster == nil {
		return fmt.Errorf("could not create cluster")
	}
	if cluster.CreateCluster.DeployToken == nil {
		return fmt.Errorf("could not fetch deploy token from cluster")
	}

	url := client.ExtUrl()
	if agentURL, err := client.AgentUrl(cluster.CreateCluster.ID); err == nil {
		url = agentURL
	}
	logf("installing agent on %s cluster with %s URL", lo.FromPtr(registration.Name), client.Url())
	return install(url, *cluster.CreateCluster.DeployToken, strings.TrimSpace(opts.ChartLoc), cluster.CreateCluster.ID)
}

func bootstrapLog(log func(string)) func(string, ...any) {
	return func(msg string, args ...any) {
		line := fmt.Sprintf(msg, args...)
		if log != nil {
			log(line)
			return
		}
		utils.Highlight("%s\n", line)
	}
}

func clusterAttributes(registration *gqlclient.ClusterRegistrationFragment) (*gqlclient.ClusterAttributes, error) {
	attributes := gqlclient.ClusterAttributes{
		Name:   lo.FromPtr(registration.Name),
		Handle: registration.Handle,
	}
	if registration.Tags != nil {
		attributes.Tags = lo.Map(registration.Tags, func(tag *gqlclient.ClusterTags, _ int) *gqlclient.TagAttributes {
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
