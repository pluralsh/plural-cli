package up

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/console/go/polly/algorithms"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider"
	pkgup "github.com/pluralsh/plural-cli/pkg/up"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

const defaultBootstrapBranch = "main"

// FlushInput carries wizard survey values for writing workspace.yaml.
type FlushInput struct {
	ProviderID string
	Values     map[string]string
	AppDomain  string
	Cloud      bool
}

// GenerateInput is the post-Flush generate step (Build → ImportCluster → Backfill → Generate).
type GenerateInput struct {
	Cloud            bool
	CloudCluster     string // Console instance name (--cloud)
	IgnorePreflights bool
	GitRef           string
	ImportClusterID  string // optional; resolved when empty and Cloud
}

// DeployInput runs up.Context.Deploy after Generate (terraform + optional git sync).
type DeployInput struct {
	Cloud            bool
	CloudCluster     string
	ImportClusterID  string
	IgnorePreflights bool
	CommitMsg        string // empty skips git.Sync (CLI CommitMsg parity)
}

// RunInput is the Plan → Flush + Generate pipeline.
type RunInput struct {
	Flush    FlushInput
	Generate GenerateInput
}

// RunResult carries values needed for a later Deploy step.
type RunResult struct {
	ImportClusterID string
}

// Progress reports a human-readable step while Run/Deploy executes.
type ProgressFunc func(step string)

// Runner executes Flush + Generate, then optionally Deploy.
type Runner interface {
	Run(ctx context.Context, in RunInput, progress ProgressFunc) (RunResult, error)
	Deploy(ctx context.Context, in DeployInput, progress ProgressFunc) error
}

// LiveRunner writes workspace.yaml and runs up.Build / Generate / Deploy.
type LiveRunner struct{}

// DefaultRunner returns the live Flush+Generate+Deploy runner.
func DefaultRunner() Runner { return LiveRunner{} }

// Run flushes the workspace, resolves ImportCluster when cloud, then generates.
func (LiveRunner) Run(ctx context.Context, in RunInput, progress ProgressFunc) (RunResult, error) {
	report := progress
	if report == nil {
		report = func(string) {}
	}
	var result RunResult

	report("Writing workspace.yaml…")
	if err := FlushWorkspace(ctx, in.Flush); err != nil {
		return result, err
	}

	gen := in.Generate
	if gen.Cloud && gen.ImportClusterID == "" {
		report("Resolving management cluster (ImportCluster)…")
		id, err := ResolveImportCluster(ctx)
		if err != nil {
			return result, err
		}
		gen.ImportClusterID = id
	}
	result.ImportClusterID = gen.ImportClusterID

	report("Generating bootstrap / terraform…")
	return result, GenerateWorkspace(ctx, gen)
}

// Deploy runs up.Context.Deploy (CreateBucket / terraform / commit / apps).
func (LiveRunner) Deploy(ctx context.Context, in DeployInput, progress ProgressFunc) error {
	report := progress
	if report == nil {
		report = func(string) {}
	}

	provider.SetCloudFlag(in.Cloud)
	report("Building deploy context…")
	upCtx, err := pkgup.Build(in.Cloud)
	if err != nil {
		return err
	}
	upCtx.IgnorePreflights(in.IgnorePreflights)

	if in.Cloud {
		id := in.ImportClusterID
		if id == "" {
			report("Resolving management cluster (ImportCluster)…")
			id, err = ResolveImportCluster(ctx)
			if err != nil {
				return err
			}
		}
		upCtx.SetImportCluster(id)
		upCtx.CloudCluster = in.CloudCluster
	}

	report("Deploying management cluster…")
	return upCtx.Deploy(func() error {
		msg := strings.TrimSpace(in.CommitMsg)
		if msg == "" {
			report("Skipping git commit (empty message)…")
			return nil
		}
		report("Pushing git commit…")
		root, err := git.Root()
		if err != nil {
			return err
		}
		return git.Sync(root, msg, false)
	})
}

// FlushWorkspace builds ProjectManifest from survey values and writes workspace.yaml.
func FlushWorkspace(ctx context.Context, in FlushInput) error {
	if strings.TrimSpace(in.ProviderID) == "" {
		return fmt.Errorf("provider is required to write workspace.yaml")
	}
	if len(in.Values) == 0 {
		return fmt.Errorf("provider survey values are required to write workspace.yaml (complete credentials/region first)")
	}
	cluster := strings.TrimSpace(in.Values["cluster"])
	if cluster == "" {
		return fmt.Errorf("cluster name is required")
	}

	conf := config.Read()
	pm, err := projectManifestFromSurvey(ctx, in.ProviderID, in.Values, in.Cloud, conf)
	if err != nil {
		return err
	}
	if err := writeWorkspaceSilent(pm, in.Cloud, cluster); err != nil {
		return err
	}
	if d := strings.TrimSpace(in.AppDomain); d != "" {
		pm.AppDomain = d
		return pm.Write(manifest.ProjectManifestPath())
	}
	return nil
}

func projectManifestFromSurvey(ctx context.Context, providerID string, values map[string]string, cloud bool, conf config.Config) (*manifest.ProjectManifest, error) {
	owner := &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint}
	cluster := strings.TrimSpace(values["cluster"])

	switch providerID {
	case api.ProviderAWS:
		region := strings.TrimSpace(values["region"])
		if region == "" {
			return nil, fmt.Errorf("region is required for aws")
		}
		project := ""
		ctxMap := map[string]interface{}{}
		if sess, identity, err := provider.GetAWSCallerIdentity(ctx); err == nil {
			project = lo.FromPtr(identity.Account)
			ctxMap["IAMSession"] = sess
		}
		return &manifest.ProjectManifest{
			Cluster:  cluster,
			Project:  project,
			Provider: api.ProviderAWS,
			Region:   region,
			Context:  ctxMap,
			Owner:    owner,
		}, nil

	case api.ProviderAzure:
		location := strings.TrimSpace(values["location"])
		if location == "" {
			return nil, fmt.Errorf("location is required for azure")
		}
		rg := strings.TrimSpace(values["resourceGroup"])
		storage := strings.TrimSpace(values["storageAccount"])
		ctxMap := map[string]interface{}{}
		if subID, tenID, _, err := provider.GetAzureAccount(); err == nil {
			ctxMap["SubscriptionId"] = subID
			ctxMap["TenantId"] = tenID
		}
		if storage != "" {
			ctxMap["StorageAccount"] = storage
		}
		return &manifest.ProjectManifest{
			Cluster:  cluster,
			Project:  rg,
			Provider: api.ProviderAzure,
			Region:   location,
			Context:  ctxMap,
			Owner:    owner,
		}, nil

	case api.ProviderGCP:
		project := strings.TrimSpace(values["project"])
		region := strings.TrimSpace(values["region"])
		if region == "" {
			region = strings.TrimSpace(values["location"])
		}
		if project == "" || region == "" {
			return nil, fmt.Errorf("project and region are required for gcp")
		}
		ctxMap := map[string]interface{}{
			"BucketLocation": strings.ToUpper(strings.Split(region, "-")[0]),
			"Location":       region,
		}
		return &manifest.ProjectManifest{
			Cluster:  cluster,
			Project:  project,
			Provider: api.ProviderGCP,
			Region:   region,
			Context:  ctxMap,
			Owner:    owner,
		}, nil

	case api.BYOK:
		ctxMap := map[string]interface{}{}
		kubePath := strings.TrimSpace(values["kubeconfig"])
		if kubePath == "" {
			kubePath = "~/.kube/config"
		}
		expanded, err := expandPath(kubePath)
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(expanded)
		if err != nil {
			return nil, fmt.Errorf("kubeconfig: %w", err)
		}
		ctxMap["kubeconfig"] = base64.StdEncoding.EncodeToString(data)
		pm := &manifest.ProjectManifest{
			Cluster:  cluster,
			Provider: api.BYOK,
			Owner:    owner,
			Context:  ctxMap,
		}
		if !cloud {
			if db := strings.TrimSpace(values["database"]); db != "" {
				ctxMap["DbUrl"] = db
			}
			if domain := strings.TrimSpace(values["domain"]); domain != "" {
				pm.Network = &manifest.NetworkConfig{Subdomain: domain, PluralDns: false}
			}
		}
		return pm, nil

	default:
		return nil, fmt.Errorf("unsupported provider %q", providerID)
	}
}

// writeWorkspaceSilent mirrors Configure without interactive bucket/network surveys.
func writeWorkspaceSilent(pm *manifest.ProjectManifest, cloud bool, cluster string) error {
	pm.BucketPrefix = cluster
	if cloud {
		pm.Bucket = fmt.Sprintf("plrl-cloud-%s-%s", cluster, algorithms.String(4))
	} else {
		pm.Bucket = fmt.Sprintf("%s-tf-state", cluster)
	}
	return pm.Write(manifest.ProjectManifestPath())
}

func expandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, p[2:]), nil
	}
	if filepath.IsAbs(p) {
		return p, nil
	}
	return filepath.Abs(p)
}

// ResolveImportCluster finds the Console cluster with handle "mgmt".
func ResolveImportCluster(ctx context.Context) (string, error) {
	_ = ctx
	conf := console.ReadConfig()
	if conf.Token == "" || conf.Url == "" {
		return "", fmt.Errorf("you have not set up a console login, you can run `plural cd login` to save your credentials")
	}
	client, err := console.NewConsoleClient(conf.Token, conf.Url)
	if err != nil {
		return "", err
	}
	clusters, err := client.ListClusters()
	if err != nil {
		return "", err
	}
	if clusters == nil || clusters.Clusters == nil {
		return "", fmt.Errorf("could not find the management cluster in your Plural cloud instance, contact support for assistance")
	}
	for _, edge := range clusters.Clusters.Edges {
		if edge == nil || edge.Node == nil {
			continue
		}
		if lo.FromPtr(edge.Node.Handle) == "mgmt" {
			return edge.Node.ID, nil
		}
	}
	return "", fmt.Errorf("could not find the management cluster in your Plural cloud instance, contact support for assistance")
}

// GenerateWorkspace runs up.Build → optional ImportCluster → Backfill → Generate.
func GenerateWorkspace(ctx context.Context, in GenerateInput) error {
	_ = ctx
	provider.SetCloudFlag(in.Cloud)

	upCtx, err := pkgup.Build(in.Cloud)
	if err != nil {
		return err
	}
	upCtx.IgnorePreflights(in.IgnorePreflights)

	if in.Cloud {
		if in.ImportClusterID == "" {
			return fmt.Errorf("ImportCluster id is required for cloud generate")
		}
		upCtx.SetImportCluster(in.ImportClusterID)
		upCtx.CloudCluster = in.CloudCluster
	}

	if err := upCtx.Backfill(); err != nil {
		return err
	}

	gitRef := in.GitRef
	if gitRef == "" {
		gitRef = defaultBootstrapBranch
	}
	dir, err := upCtx.Generate(gitRef)
	if dir != "" {
		defer func() { _ = os.RemoveAll(dir) }()
	}
	return err
}
