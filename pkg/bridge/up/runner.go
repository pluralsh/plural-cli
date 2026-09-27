package up

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
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
	ProviderID   string
	Values       map[string]string
	AppDomain    string
	Cloud        bool
	BucketPrefix string // self-hosted Configure bucket naming
	PluralDNS    string // self-hosted subdomain.onplural.sh (full domain)
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
// Git commit runs inside Deploy at the "commit" checkpoint (after mgmt terraform,
// before apps) — matching plural up. Set PromptCommit to survey during that step.
type DeployInput struct {
	Cloud            bool
	CloudCluster     string
	ImportClusterID  string
	IgnorePreflights bool
	CommitMsg        string        // if set, used at commit checkpoint; empty + !PromptCommit skips
	PromptCommit     bool          // survey (or CommitPrompt) at commit checkpoint after mgmt terraform
	CommitPrompt     func() string // optional; when set with PromptCommit, called instead of survey
	CommittedMsg     *string       // optional out: final commit message used (may be empty if skipped)
	Output           io.Writer     // optional; terraform + highlight output (TUI log capture)
}

// DestroyInput tears down the management cluster (plural down).
type DestroyInput struct {
	Cloud  bool
	Output io.Writer // optional; terraform output (TUI log capture)
}

// RunInput is the Plan → Flush + Generate pipeline.
type RunInput struct {
	Flush     FlushInput
	Generate  GenerateInput
	SkipFlush bool      // true when workspace.yaml already exists (CLI ensureWorkspace path)
	Output    io.Writer // optional; generation-related stdout (TUI log capture)
}

// RunResult carries values needed for a later Deploy step.
type RunResult struct {
	ImportClusterID string
}

// Progress reports a human-readable step while Run/Deploy/Destroy executes.
type ProgressFunc func(step string)

// Runner executes Flush + Generate, Deploy, and Destroy.
type Runner interface {
	Run(ctx context.Context, in RunInput, progress ProgressFunc) (RunResult, error)
	Deploy(ctx context.Context, in DeployInput, progress ProgressFunc) error
	Destroy(ctx context.Context, in DestroyInput, progress ProgressFunc) error
}

// LiveRunner writes workspace.yaml and runs up.Build / Generate / Deploy.
type LiveRunner struct{}

// DefaultRunner returns the live Flush+Generate+Deploy+Destroy runner.
func DefaultRunner() Runner { return LiveRunner{} }

// Run flushes the workspace, resolves ImportCluster when cloud, then generates.
func (LiveRunner) Run(ctx context.Context, in RunInput, progress ProgressFunc) (RunResult, error) {
	return withCommandOutput(in.Output, func() (RunResult, error) {
		report := progress
		if report == nil {
			report = func(string) {}
		}
		var result RunResult

		if in.SkipFlush {
			report("Skipping workspace.yaml write (already initialized)…")
		} else {
			report("Writing workspace.yaml…")
			if err := FlushWorkspace(ctx, in.Flush); err != nil {
				return result, err
			}
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
	})
}

// Deploy runs up.Context.Deploy (CreateBucket / terraform / commit / apps).
func (LiveRunner) Deploy(ctx context.Context, in DeployInput, progress ProgressFunc) error {
	return withCommandOutputErr(in.Output, func() error {
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
			if msg == "" && in.PromptCommit {
				report("Commit checkpoint — enter a commit message…")
				if in.CommitPrompt != nil {
					msg = strings.TrimSpace(in.CommitPrompt())
				} else {
					msg = promptCommitMessage()
				}
			}
			if in.CommittedMsg != nil {
				*in.CommittedMsg = msg
			}
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
	})
}

// Destroy tears down the management cluster (plural down).
func (LiveRunner) Destroy(ctx context.Context, in DestroyInput, progress ProgressFunc) error {
	_ = ctx
	return withCommandOutputErr(in.Output, func() error {
		report := func(step string) {
			if progress != nil {
				progress(step)
			}
		}
		report("Building destroy context…")
		upCtx, err := pkgup.Build(in.Cloud)
		if err != nil {
			return err
		}
		report("Destroying management cluster terraform…")
		return upCtx.Destroy()
	})
}

func withCommandOutput[T any](w io.Writer, fn func() (T, error)) (T, error) {
	if w == nil {
		return fn()
	}
	prevOut, prevErr := color.Output, color.Error
	pkgup.SetCommandOutput(w, w)
	color.Output = w
	color.Error = w
	defer func() {
		pkgup.SetCommandOutput(nil, nil)
		color.Output = prevOut
		color.Error = prevErr
	}()
	return fn()
}

func withCommandOutputErr(w io.Writer, fn func() error) error {
	_, err := withCommandOutput(w, func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
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
	pm.AppDomain = strings.TrimSpace(in.AppDomain)
	pm.AppDomainConfigured = true
	return writeWorkspaceSilent(pm, in.Cloud, cluster, in.BucketPrefix, in.PluralDNS)
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

// writeWorkspaceSilent mirrors Configure without interactive surveys.
// Self-hosted requires BucketPrefix (+ optional PluralDNS from the TUI prompts).
func writeWorkspaceSilent(pm *manifest.ProjectManifest, cloud bool, cluster, bucketPrefix, pluralDNS string) error {
	if cloud {
		pm.BucketPrefix = cluster
		pm.Bucket = fmt.Sprintf("plrl-cloud-%s-%s", cluster, algorithms.String(4))
	} else {
		prefix := strings.TrimSpace(bucketPrefix)
		if prefix == "" {
			return fmt.Errorf("bucket naming prefix is required for self-hosted up")
		}
		if err := ValidateBucketPrefix(prefix); err != nil {
			return err
		}
		pm.BucketPrefix = prefix
		pm.Bucket = fmt.Sprintf("%s-tf-state", prefix)
		if d := strings.TrimSpace(pluralDNS); d != "" {
			pm.Network = &manifest.NetworkConfig{Subdomain: d, PluralDns: true}
		}
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
