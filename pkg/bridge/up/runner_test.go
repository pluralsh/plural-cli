package up

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/manifest"
)

func TestFlushWorkspaceRequiresValues(t *testing.T) {
	err := FlushWorkspace(context.Background(), FlushInput{ProviderID: "aws"})
	if err == nil || !strings.Contains(err.Error(), "survey values") {
		t.Fatalf("expected survey values error, got %v", err)
	}
}

func TestFlushWorkspaceAWSWritesManifest(t *testing.T) {
	dir := t.TempDir()
	prev, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	err := FlushWorkspace(context.Background(), FlushInput{
		ProviderID: api.ProviderAWS,
		Values:     map[string]string{"cluster": "demo", "region": "us-east-2"},
		AppDomain:  "apps.example.com",
		Cloud:      true,
	})
	if err != nil {
		t.Fatalf("FlushWorkspace: %v", err)
	}
	path := filepath.Join(dir, "workspace.yaml")
	pm, err := manifest.ReadProject(path)
	if err != nil {
		t.Fatalf("ReadProject: %v", err)
	}
	if pm.Cluster != "demo" || pm.Region != "us-east-2" || pm.Provider != api.ProviderAWS {
		t.Fatalf("manifest = %#v", pm)
	}
	if pm.AppDomain != "apps.example.com" {
		t.Fatalf("appDomain = %q", pm.AppDomain)
	}
	if pm.Bucket == "" || pm.BucketPrefix != "demo" {
		t.Fatalf("bucket=%q prefix=%q", pm.Bucket, pm.BucketPrefix)
	}
	if !strings.HasPrefix(pm.Bucket, "plrl-cloud-demo-") {
		t.Fatalf("cloud bucket naming = %q", pm.Bucket)
	}
}

func TestFlushWorkspaceSelfHostedBucket(t *testing.T) {
	dir := t.TempDir()
	prev, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	if err := FlushWorkspace(context.Background(), FlushInput{
		ProviderID: api.ProviderAWS,
		Values:     map[string]string{"cluster": "acme", "region": "eu-west-1"},
		Cloud:      false,
	}); err != nil {
		t.Fatalf("FlushWorkspace: %v", err)
	}
	pm, err := manifest.ReadProject(filepath.Join(dir, "workspace.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if pm.Bucket != "acme-tf-state" {
		t.Fatalf("self-hosted bucket = %q", pm.Bucket)
	}
}

func TestFakeRunnerRecordsCloud(t *testing.T) {
	f := &fakeRunner{}
	in := RunInput{
		Flush:    FlushInput{ProviderID: "aws", Values: map[string]string{"cluster": "x", "region": "us-east-2"}, Cloud: true},
		Generate: GenerateInput{Cloud: true, CloudCluster: "demo-cloud", IgnorePreflights: true},
	}
	res, err := f.Run(context.Background(), in, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 1 || !f.calls[0].Generate.Cloud || f.calls[0].Generate.CloudCluster != "demo-cloud" {
		t.Fatalf("calls = %#v", f.calls)
	}
	if res.ImportClusterID != "mgmt-id" {
		t.Fatalf("ImportClusterID = %q", res.ImportClusterID)
	}
	if err := f.Deploy(context.Background(), DeployInput{Cloud: true, ImportClusterID: res.ImportClusterID, CommitMsg: "init"}, nil); err != nil {
		t.Fatal(err)
	}
	if len(f.deploys) != 1 || f.deploys[0].CommitMsg != "init" {
		t.Fatalf("deploys = %#v", f.deploys)
	}
}

type fakeRunner struct {
	calls    []RunInput
	deploys  []DeployInput
	progress []string
	err      error
	deployErr error
}

func (f *fakeRunner) Run(_ context.Context, in RunInput, progress ProgressFunc) (RunResult, error) {
	f.calls = append(f.calls, in)
	if progress != nil {
		progress("Writing workspace.yaml…")
		progress("Generating…")
		f.progress = append(f.progress, "Writing workspace.yaml…", "Generating…")
	}
	res := RunResult{}
	if in.Generate.Cloud {
		res.ImportClusterID = "mgmt-id"
	}
	return res, f.err
}

func (f *fakeRunner) Deploy(_ context.Context, in DeployInput, progress ProgressFunc) error {
	f.deploys = append(f.deploys, in)
	if progress != nil {
		progress("Deploying…")
	}
	return f.deployErr
}
