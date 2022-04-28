package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	v1 "k8s.io/api/core/v1"

	"cloud.google.com/go/storage"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/errors"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type GCPProvider struct {
	Clust         string `survey:"cluster"`
	Proj          string `survey:"project"`
	bucket        string
	Reg           string `survey:"region"`
	storageClient *storage.Client
	ctx           map[string]interface{}
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
  		"us-east1",
  		"us-west1",
  		"us-west2",
	}
)

var gcpSurvey = []*survey.Question{
	{
		Name:     "cluster",
		Prompt:   &survey.Input{Message: "Enter the name of your cluster"},
		Validate: validCluster,
	},
	{
		Name:     "project",
		Prompt:   &survey.Input{Message: "Enter the name of its gcp project"},
		Validate: utils.ValidateAlphaNumeric,
	},
	{
		Name:     "region",
		Prompt:   &survey.Select{Message: "What region will you deploy to?", Default: "us-east1", Options: gcpRegions},
		Validate: survey.Required,
	},
}

func mkGCP(conf config.Config) (*GCPProvider, error) {
	provider := &GCPProvider{}
	if err := survey.Ask(gcpSurvey, provider); err != nil {
		return nil, err
	}

	client, err := storageClient()
	if err != nil {
		return nil, err
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

	if err := projectManifest.Configure(); err != nil {
		return nil, err
	}

	provider.bucket = projectManifest.Bucket
	return provider, nil
}

func getBucketLocation(region string) BucketLocation {
	if strings.Contains(strings.ToLower(region), "us") ||
		strings.Contains(strings.ToLower(region), "northamerica") ||
		strings.Contains(strings.ToLower(region), "southamerica") {
		return BucketLocationUS
	} else if strings.Contains(strings.ToLower(region), "europe") {
		return BucketLocationEU
	} else if strings.Contains(strings.ToLower(region), "asia") ||
		strings.Contains(strings.ToLower(region), "australia") {
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
		man.Write(manifest.ProjectManifestPath())
	} else if location := strings.Split(man.Region, "-"); len(location) >= 3 {
		man.Context["Location"] = man.Region
		man.Region = fmt.Sprintf("%s-%s", location[0], location[1])
		man.Context["BucketLocation"] = getBucketLocation(man.Region)
		man.Write(manifest.ProjectManifestPath())
	}
	// Needed to update legacy deployments
	if _, ok := man.Context["BucketLocation"]; !ok {
		man.Context["BucketLocation"] = "US"
		man.Write(manifest.ProjectManifestPath())
	}
	// Needed to update legacy deployments
	if _, ok := man.Context["Location"]; !ok {
		man.Context["Location"] = man.Region
		man.Write(manifest.ProjectManifestPath())
	}

	return &GCPProvider{man.Cluster, man.Project, man.Bucket, man.Region, client, man.Context}, nil
}

func (gcp *GCPProvider) KubeConfig() error {
	if utils.InKubernetes() {
		return nil
	}

	cmd := exec.Command(
		"gcloud", "container", "clusters", "get-credentials", gcp.Clust,
		"--region", gcp.Region(), "--project", gcp.Proj)
	return utils.Execute(cmd)
}

func (gcp *GCPProvider) CreateBackend(prefix string, ctx map[string]interface{}) (string, error) {
	if err := gcp.mkBucket(gcp.bucket); err != nil {
		return "", errors.ErrorWrap(err, fmt.Sprintf("Failed to create terraform state bucket %s", gcp.Bucket))
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
	scaffold, err := GetProviderScaffold("GCP")
	if err != nil {
		return "", err
	}
	return template.RenderString(scaffold, ctx)
}

func (gcp *GCPProvider) mkBucket(name string) error {
	bkt := gcp.storageClient.Bucket(name)
	if _, err := bkt.Attrs(context.Background()); err != nil {
		return bkt.Create(context.Background(), gcp.Project(), &storage.BucketAttrs{
			Location: fmt.Sprintf("%s", getBucketLocation(gcp.Reg)),
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
	defer c.Close()

	_, err = c.Delete(ctx, &computepb.DeleteInstanceRequest{
		Instance: node.Name,
		Project:  gcp.Project(),
		Zone:     gcp.Region(),
	})

	return errors.ErrorWrap(err, "failed to delete instance")
}
