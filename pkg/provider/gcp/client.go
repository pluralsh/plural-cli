package gcp

import (
	"context"
	"errors"
	"fmt"
	"sync"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	container "cloud.google.com/go/container/apiv1"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	serviceusage "cloud.google.com/go/serviceusage/apiv1"
	"cloud.google.com/go/storage"
	"github.com/pluralsh/polly/algorithms"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/iterator"
	oauth3 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	lock   sync.RWMutex
	client *internalClient
)

func defaultRegions() []string {
	return []string{
		"asia-east1",
		"asia-east2",
		"asia-northeast1",
		"asia-northeast2",
		"asia-northeast3",
		"asia-south1",
		"asia-southeast1",
		"australia-southeast1",
		"asia-northeast1",
		"europe-central2",
		"europe-west2",
		"europe-west3",
		"europe-north1",
		"europe-southwest1",
		"us-east1",
		"us-east4",
		"us-west1",
		"us-west2",
		"us-central1",
		"northamerica-northeast1",
		"northamerica-northeast2",
		"southamerica-east1",
		"southamerica-west1",
	}
}

type internalClient struct {
	ctx                  context.Context
	storageClient        *storage.Client
	projectsClient       *resourcemanager.ProjectsClient
	regionsClient        *compute.RegionsClient
	serviceUsageClient   *serviceusage.Client
	clusterManagerClient *container.ClusterManagerClient
}

func (in *internalClient) isProjectExists(id string) (bool, error) {
	_, err := in.projectsClient.GetProject(in.ctx, &resourcemanagerpb.GetProjectRequest{
		Name: fmt.Sprintf("projects/%s", id),
	})

	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.NotFound {
				return false, nil
			}
		}

		return false, err
	}

	return true, nil
}

func (in *internalClient) projects() ([]string, error) {
	it := in.projectsClient.SearchProjects(in.ctx, &resourcemanagerpb.SearchProjectsRequest{})

	projects := make([]string, 0)
	for {
		project, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, err
		}

		projects = append(projects, project.GetProjectId())
	}

	return projects, nil
}

func (in *internalClient) regions(projectID string) ([]string, error) {
	it := in.regionsClient.List(in.ctx, &computepb.ListRegionsRequest{Project: projectID})

	regions := make([]string, 0)
	for {
		resp, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, err
		}

		regions = append(regions, resp.GetName())
	}

	return regions, nil
}

func (in *internalClient) managedZones(projectID string) ([]string, error) {
	service, err := dns.NewService(context.Background())
	if err != nil {
		return nil, err
	}

	managedZonesService := dns.NewManagedZonesService(service)
	response, err := managedZonesService.List(projectID).Do()
	if err != nil {
		return nil, err
	}

	return algorithms.Map(response.ManagedZones, func(z *dns.ManagedZone) string { return z.Name }), nil
}

func (in *internalClient) loggedInUserInfo(ctx context.Context) (email, name string, err error) {
	defaultTokenSource, err := google.DefaultTokenSource(ctx)
	if err != nil {
		return
	}

	svc, err := oauth3.NewService(ctx, option.WithTokenSource(defaultTokenSource))
	if err != nil {
		return
	}

	userInfo, err := oauth3.NewUserinfoV2MeService(svc).Get().Do()
	if err != nil {
		return
	}

	return userInfo.Email, userInfo.Name, nil
}

func (in *internalClient) initGoogleSDK() error {

	storageClient, err := storage.NewClient(in.ctx, option.WithScopes(storage.ScopeReadWrite))
	if err != nil {
		return err
	}

	projectsClient, err := resourcemanager.NewProjectsClient(in.ctx)
	if err != nil {
		return err
	}

	regionsClient, err := compute.NewRegionsRESTClient(in.ctx)
	if err != nil {
		return err
	}

	serviceUsageClient, err := serviceusage.NewClient(in.ctx)
	if err != nil {
		return err
	}

	clusterManagerClient, err := container.NewClusterManagerClient(in.ctx)
	if err != nil {
		return err
	}

	in.storageClient = storageClient
	in.projectsClient = projectsClient
	in.regionsClient = regionsClient
	in.serviceUsageClient = serviceUsageClient
	in.clusterManagerClient = clusterManagerClient

	return nil
}

func (in *internalClient) init() (*internalClient, error) {
	if err := in.initGoogleSDK(); err != nil {
		return nil, fmt.Errorf("failed to initialize google sdk. Run 'gcloud auth application-default login' to log in first: %w", err)
	}

	return in, nil
}

func newInternalClient() (*internalClient, error) {
	return (&internalClient{
		ctx: context.Background(),
	}).init()
}

func initClient() error {
	lock.RLock()
	if client != nil {
		lock.RUnlock()
		return nil
	}
	lock.RUnlock()

	tmpClient, err := newInternalClient()
	if err != nil {
		return err
	}

	lock.Lock()
	client = tmpClient
	lock.Unlock()
	return nil
}

func Projects() ([]string, error) {
	if err := initClient(); err != nil {
		return nil, err
	}

	return client.projects()
}

func Regions(projectID string) []string {
	if err := initClient(); err != nil {
		return defaultRegions()
	}

	regions, err := client.regions(projectID)
	if err != nil {
		return defaultRegions()
	}

	return regions
}

func ManagedZones(projectID string) ([]string, error) {
	if err := initClient(); err != nil {
		return nil, err
	}

	return client.managedZones(projectID)
}

func StorageClient() (*storage.Client, error) {
	if err := initClient(); err != nil {
		return nil, err
	}

	return client.storageClient, nil
}

func IsProjectExists(id string) (bool, error) {
	if err := initClient(); err != nil {
		return false, err
	}

	return client.isProjectExists(id)
}

func ServiceUsageClient() (*serviceusage.Client, error) {
	if err := initClient(); err != nil {
		return nil, err
	}

	return client.serviceUsageClient, nil
}

func ClusterManagerClient() (*container.ClusterManagerClient, error) {
	if err := initClient(); err != nil {
		return nil, err
	}

	return client.clusterManagerClient, nil
}
