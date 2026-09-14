package edge

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"

	pkgedge "github.com/pluralsh/plural-cli/pkg/edge"
)

type fakeResolver struct {
	url, token string
	err        error
}

func (f fakeResolver) ActiveConsole(context.Context) (string, string, error) {
	return f.url, f.token, f.err
}

type fakeAPI struct {
	token string
}

func (f fakeAPI) GetUser(string) (*gqlclient.UserFragment, error) {
	return &gqlclient.UserFragment{ID: "user-1"}, nil
}
func (f fakeAPI) GetProject(string) (*gqlclient.ProjectFragment, error) {
	return &gqlclient.ProjectFragment{ID: "proj-1"}, nil
}
func (f fakeAPI) CreateBootstrapToken(gqlclient.BootstrapTokenAttributes) (string, error) {
	return f.token, nil
}
func (f fakeAPI) CreateClusterRegistration(gqlclient.ClusterRegistrationCreateAttributes) (*gqlclient.ClusterRegistrationFragment, error) {
	return nil, errors.New("not implemented")
}
func (f fakeAPI) IsClusterRegistrationComplete(string) (bool, *gqlclient.ClusterRegistrationFragment) {
	return false, nil
}
func (f fakeAPI) CreateCluster(gqlclient.ClusterAttributes) (*gqlclient.CreateCluster, error) {
	return nil, errors.New("not implemented")
}
func (f fakeAPI) Url() string                     { return "https://console.example" }
func (f fakeAPI) ExtUrl() string                  { return "https://console.example/ext" }
func (f fakeAPI) AgentUrl(string) (string, error) { return "", errors.New("not implemented") }

type bootstrapFake struct {
	fakeAPI
	reg     *gqlclient.ClusterRegistrationFragment
	cluster *gqlclient.CreateCluster
}

func (f bootstrapFake) CreateClusterRegistration(gqlclient.ClusterRegistrationCreateAttributes) (*gqlclient.ClusterRegistrationFragment, error) {
	return f.reg, nil
}
func (f bootstrapFake) IsClusterRegistrationComplete(string) (bool, *gqlclient.ClusterRegistrationFragment) {
	return true, f.reg
}
func (f bootstrapFake) CreateCluster(gqlclient.ClusterAttributes) (*gqlclient.CreateCluster, error) {
	return f.cluster, nil
}
func (f bootstrapFake) AgentUrl(string) (string, error) { return "https://agent.example", nil }

func TestBuildImageRequiresConsoleWithoutCloudConfig(t *testing.T) {
	service := NewService(fakeResolver{err: errors.New("no console")})
	err := service.BuildImage(t.Context(), pkgedge.ImageOptions{Password: "x"}, nil)
	if err == nil || !strings.Contains(err.Error(), "no console") {
		t.Fatalf("error = %v", err)
	}
}

func TestFlashDelegates(t *testing.T) {
	var got pkgedge.FlashOptions
	service := NewService(fakeResolver{})
	service.flash = func(options pkgedge.FlashOptions) error {
		got = options
		return nil
	}
	if err := service.Flash(t.Context(), pkgedge.FlashOptions{Image: "kairos.img", Device: "/dev/sda"}, nil); err != nil {
		t.Fatal(err)
	}
	if got.Image != "kairos.img" || got.Device != "/dev/sda" {
		t.Fatalf("flash options = %#v", got)
	}
}

func TestBootstrapRequiresConsole(t *testing.T) {
	service := NewService(fakeResolver{err: errors.New("no console")})
	err := service.Bootstrap(t.Context(), pkgedge.BootstrapOptions{MachineID: "pi-1"}, nil)
	if err == nil || !strings.Contains(err.Error(), "no console") {
		t.Fatalf("error = %v", err)
	}
}

func TestBootstrapDelegates(t *testing.T) {
	var installed struct{ url, token, chart, id string }
	service := NewService(fakeResolver{url: "https://console.example", token: "t"})
	service.newClient = func(string, string) (API, error) {
		return bootstrapFake{
			reg: &gqlclient.ClusterRegistrationFragment{Name: lo.ToPtr("edge-pi")},
			cluster: &gqlclient.CreateCluster{CreateCluster: &gqlclient.CreateCluster_CreateCluster{
				ID:          "cluster-1",
				DeployToken: lo.ToPtr("deploy-token"),
			}},
		}, nil
	}
	service.install = func(url, token, chartLoc, clusterID string) error {
		installed.url, installed.token, installed.chart, installed.id = url, token, chartLoc, clusterID
		return nil
	}
	if err := service.Bootstrap(t.Context(), pkgedge.BootstrapOptions{MachineID: "pi-1", ChartLoc: "chart.tgz", PollInterval: time.Millisecond}, nil); err != nil {
		t.Fatal(err)
	}
	if installed.url != "https://agent.example" || installed.token != "deploy-token" || installed.chart != "chart.tgz" || installed.id != "cluster-1" {
		t.Fatalf("install = %#v", installed)
	}
}

func TestDownloadDelegates(t *testing.T) {
	var got pkgedge.DownloadOptions
	service := NewService(fakeResolver{})
	service.download = func(options pkgedge.DownloadOptions, _ func(string)) error {
		got = options
		return nil
	}
	if err := service.Download(t.Context(), pkgedge.DownloadOptions{OCIURL: "ghcr.io/img", To: "/tmp"}, nil); err != nil {
		t.Fatal(err)
	}
	if got.OCIURL != "ghcr.io/img" || got.To != "/tmp" {
		t.Fatalf("download options = %#v", got)
	}
}
