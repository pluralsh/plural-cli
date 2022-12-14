package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	serviceusage "cloud.google.com/go/serviceusage/apiv1"
	"cloud.google.com/go/storage"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	gcompute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	serviceusagepb "google.golang.org/genproto/googleapis/api/serviceusage/v1"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	resourcemanagerpb "google.golang.org/genproto/googleapis/cloud/resourcemanager/v3"
	v1 "k8s.io/api/core/v1"
)

type GCPProvider struct {
	Clust         string `survey:"cluster"`
	Proj          string `survey:"project"`
	bucket        string
	Reg           string `survey:"region"`
	storageClient *storage.Client
	ctx           map[string]interface{}
	writer        manifest.Writer
}

type BucketLocation string

const (
	BucketLocationUS   BucketLocation = "US"
	BucketLocationEU   BucketLocation = "EU"
	BucketLocationASIA BucketLocation = "ASIA"
)

var (
	gcpRegions = []string{
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
)

func getGCPSurvey() []*survey.Question {
	prompt, validate := setProject()
	return []*survey.Question{
		{
			Name:     "cluster",
			Prompt:   &survey.Input{Message: "Enter the name of your cluster"},
			Validate: validCluster,
		},
		{
			Name:     "project",
			Prompt:   prompt,
			Validate: validate,
		},
		{
			Name:     "region",
			Prompt:   &survey.Select{Message: "What region will you deploy to?", Default: "us-east1", Options: gcpRegions},
			Validate: survey.Required,
		},
	}
}

func setProject() (survey.Prompt, survey.Validator) {
	projects, err := getGcpProjects()
	if err != nil {
		return &survey.Input{Message: "Enter the name of its gcp project"}, utils.ValidateAlphaNumeric
	}

	return &survey.Select{Message: "Select the name of gcp project?", Options: projects}, survey.Required
}

func getGcpProjects() ([]string, error) {
	client, err := google.DefaultClient(context.Background(),
		gcompute.ComputeScope)
	if err != nil {
		return nil, err
	}
	svc, err := cloudresourcemanager.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	projectList := []string{}
	resp, err := svc.Projects.List().Do()
	if err != nil {
		return nil, err
	}
	for _, project := range resp.Projects {
		projectList = append(projectList, project.Name)
	}
	return projectList, nil
}

func mkGCP(conf config.Config) (provider *GCPProvider, err error) {
	provider = &GCPProvider{}
	if err = survey.Ask(getGCPSurvey(), provider); err != nil {
		return
	}

	client, err := storageClient()
	if err != nil {
		return
	}

	provider.storageClient = client
	provider.ctx = map[string]interface{}{
		"BucketLocation": getBucketLocation(provider.Region()),
		// Location might conflict with the region set by users. However, this is only a temporary solution that should be removed
		"Location": provider.Reg,
	}

	projectManifest := manifest.ProjectManifest{
		Cluster:  provider.Cluster(),
		Project:  provider.Project(),
		Provider: GCP,
		Region:   provider.Region(),
		Context:  provider.Context(),
		Owner:    &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint},
	}

	provider.writer = projectManifest.Configure()
	provider.bucket = projectManifest.Bucket
	return
}

func getBucketLocation(region string) BucketLocation {
	reg := strings.ToLower(region)
	//nolint:gocritic
	if strings.Contains(reg, "us") ||
		strings.Contains(reg, "northamerica") ||
		strings.Contains(reg, "southamerica") {
		return BucketLocationUS
	} else if strings.Contains(reg, "europe") {
		return BucketLocationEU
	} else if strings.Contains(reg, "asia") ||
		strings.Contains(reg, "australia") {
		return BucketLocationASIA
	} else {
		return BucketLocationUS
	}
}

func storageClient() (*storage.Client, error) {
	client, err := storage.NewClient(context.Background())
	return client, err
}

func gcpFromManifest(man *manifest.ProjectManifest) (*GCPProvider, error) {
	client, err := storageClient()
	if err != nil {
		return nil, err
	}

	// Needed to update legacy deployments
	if man.Region == "" {
		man.Region = "us-east1"
		if err := man.Write(manifest.ProjectManifestPath()); err != nil {
			return nil, err
		}
	} else if location := strings.Split(man.Region, "-"); len(location) >= 3 {
		man.Context["Location"] = man.Region
		man.Region = fmt.Sprintf("%s-%s", location[0], location[1])
		man.Context["BucketLocation"] = getBucketLocation(man.Region)
		if err := man.Write(manifest.ProjectManifestPath()); err != nil {
			return nil, err
		}
	}
	// Needed to update legacy deployments
	if _, ok := man.Context["BucketLocation"]; !ok {
		man.Context["BucketLocation"] = "US"
		if err := man.Write(manifest.ProjectManifestPath()); err != nil {
			return nil, err
		}
	}
	// Needed to update legacy deployments
	if _, ok := man.Context["Location"]; !ok {
		man.Context["Location"] = man.Region
		if err := man.Write(manifest.ProjectManifestPath()); err != nil {
			return nil, err
		}
	}

	return &GCPProvider{man.Cluster, man.Project, man.Bucket, man.Region, client, man.Context, nil}, nil
}

func (gcp *GCPProvider) KubeConfig() error {
	if kubernetes.InKubernetes() {
		return nil
	}

	cmd := exec.Command(
		"gcloud", "container", "clusters", "get-credentials", gcp.Clust,
		"--region", gcp.Region(), "--project", gcp.Proj)
	return utils.Execute(cmd)
}

func (gcp *GCPProvider) Flush() error {
	if gcp.writer == nil {
		return nil
	}
	return gcp.writer()
}

func (gcp *GCPProvider) CreateBackend(prefix string, version string, ctx map[string]interface{}) (string, error) {
	if err := gcp.mkBucket(gcp.bucket); err != nil {
		return "", errors.ErrorWrap(err, fmt.Sprintf("Failed to create terraform state bucket %s", gcp.Bucket()))
	}

	ctx["Project"] = gcp.Project()
	// Location is here for backwards compatibility
	ctx["Location"] = gcp.Context()["Location"]
	ctx["Region"] = gcp.Region()
	ctx["Bucket"] = gcp.Bucket()
	ctx["Prefix"] = prefix
	ctx["ClusterCreated"] = false
	ctx["__CLUSTER__"] = gcp.Cluster()
	if cluster, ok := ctx["cluster"]; ok {
		ctx["Cluster"] = cluster
		ctx["ClusterCreated"] = true
	} else {
		ctx["Cluster"] = fmt.Sprintf(`"%s"`, gcp.Cluster())
	}
	scaffold, err := GetProviderScaffold("GCP", version)
	if err != nil {
		return "", err
	}
	return template.RenderString(scaffold, ctx)
}

func (gcp *GCPProvider) mkBucket(name string) error {
	bkt := gcp.storageClient.Bucket(name)
	if _, err := bkt.Attrs(context.Background()); err != nil {
		return bkt.Create(context.Background(), gcp.Project(), &storage.BucketAttrs{
			Location: string(getBucketLocation(gcp.Reg)),
		})
	}
	return nil
}

func (gcp *GCPProvider) Name() string {
	return GCP
}

func (gcp *GCPProvider) Cluster() string {
	return gcp.Clust
}

func (gcp *GCPProvider) Project() string {
	return gcp.Proj
}

func (gcp *GCPProvider) Bucket() string {
	return gcp.bucket
}

func (gcp *GCPProvider) Region() string {
	return gcp.Reg
}

func (gcp *GCPProvider) Context() map[string]interface{} {
	return gcp.ctx
}

func (gcp *GCPProvider) Decommision(node *v1.Node) error {
	ctx := context.Background()
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return errors.ErrorWrap(err, "failed to initialize compute client")
	}
	defer func(c *compute.InstancesClient) {
		_ = c.Close()
	}(c)

	_, err = c.Delete(ctx, &computepb.DeleteInstanceRequest{
		Instance: node.Name,
		Project:  gcp.Project(),
		Zone:     gcp.Region(),
	})

	return errors.ErrorWrap(err, "failed to delete instance")
}

func (gcp *GCPProvider) Preflights() []*Preflight {
	return []*Preflight{
		{Name: "Enabled Services", Callback: gcp.validateEnabled},
	}
}

func (gcp *GCPProvider) validateEnabled() error {
	ctx := context.Background()
	c, err := serviceusage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("Could not set up gcp client. Are your credentials valid?")
	}
	defer func(c *serviceusage.Client) {
		_ = c.Close()
	}(c)

	errEnabled := fmt.Errorf("You don't have necessary services enabled. Please run: `gcloud services enable serviceusage.googleapis.com cloudresourcemanager.googleapis.com container.googleapis.com` with an owner of the project to enable or enable them in the GCP console.")
	proj, err := gcp.getProject()
	if err != nil {
		return errEnabled
	}

	wrapped := func(name string) string {
		return fmt.Sprintf("projects/%s/services/%s", proj.ProjectId, name)
	}
	req := &serviceusagepb.BatchGetServicesRequest{
		Parent: fmt.Sprintf("projects/%s", proj.ProjectId),
		Names: []string{
			wrapped("serviceusage.googleapis.com"),
			wrapped("cloudresourcemanager.googleapis.com"),
			wrapped("container.googleapis.com"),
		},
	}
	resp, err := c.BatchGetServices(ctx, req)
	if err != nil {
		return errEnabled
	}

	for _, svc := range resp.Services {
		if svc.State != serviceusagepb.State_ENABLED {
			return errEnabled
		}
	}
	return nil
}

func (gcp *GCPProvider) getProject() (*resourcemanagerpb.Project, error) {
	ctx := context.Background()
	c, err := resourcemanager.NewProjectsClient(ctx)
	if err != nil {
		return nil, err
	}
	defer func(c *resourcemanager.ProjectsClient) {
		_ = c.Close()
	}(c)
	return c.GetProject(ctx, &resourcemanagerpb.GetProjectRequest{Name: fmt.Sprintf("projects/%s", gcp.Project())})
}
