package edge

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"
)

type fakeBootstrapAPI struct {
	createErr       error
	alreadyTaken    bool
	polls           int
	completeAfter   int
	registration    *gqlclient.ClusterRegistrationFragment
	cluster         *gqlclient.CreateCluster
	createCluster   error
	url, ext, agent string
	machineID       string
	attrs           gqlclient.ClusterAttributes
}

func (f *fakeBootstrapAPI) CreateClusterRegistration(attrs gqlclient.ClusterRegistrationCreateAttributes) (*gqlclient.ClusterRegistrationFragment, error) {
	f.machineID = attrs.MachineID
	if f.alreadyTaken {
		return nil, errors.New("machine_id has already been taken")
	}
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.registration, nil
}

func (f *fakeBootstrapAPI) IsClusterRegistrationComplete(string) (bool, *gqlclient.ClusterRegistrationFragment) {
	f.polls++
	if f.completeAfter > 0 && f.polls < f.completeAfter {
		return false, nil
	}
	return true, f.registration
}

func (f *fakeBootstrapAPI) CreateCluster(attrs gqlclient.ClusterAttributes) (*gqlclient.CreateCluster, error) {
	f.attrs = attrs
	if f.createCluster != nil {
		return nil, f.createCluster
	}
	return f.cluster, nil
}

func (f *fakeBootstrapAPI) Url() string    { return f.url }
func (f *fakeBootstrapAPI) ExtUrl() string { return f.ext }
func (f *fakeBootstrapAPI) AgentUrl(string) (string, error) {
	if f.agent == "" {
		return "", errors.New("no agent url")
	}
	return f.agent, nil
}

func testRegistration() *gqlclient.ClusterRegistrationFragment {
	return &gqlclient.ClusterRegistrationFragment{
		Name:    lo.ToPtr("edge-pi"),
		Handle:  lo.ToPtr("edge-pi"),
		Project: &gqlclient.TinyProjectFragment{ID: "proj-1"},
		Tags:    []*gqlclient.ClusterTags{{Name: "env", Value: "edge"}},
	}
}

func testCluster() *gqlclient.CreateCluster {
	return &gqlclient.CreateCluster{
		CreateCluster: &gqlclient.CreateCluster_CreateCluster{
			ID:          "cluster-1",
			DeployToken: lo.ToPtr("deploy-token"),
		},
	}
}

func TestBootstrapRegistersCreatesAndInstalls(t *testing.T) {
	api := &fakeBootstrapAPI{
		registration: testRegistration(),
		cluster:      testCluster(),
		url:          "https://console.example",
		ext:          "https://console.example/ext",
		agent:        "https://agent.example",
	}
	var got struct{ url, token, chart, id string }
	err := Bootstrap(t.Context(), api, func(url, token, chartLoc, clusterID string) error {
		got.url, got.token, got.chart, got.id = url, token, chartLoc, clusterID
		return nil
	}, BootstrapOptions{MachineID: "pi-1", ChartLoc: "oci://chart", PollInterval: time.Millisecond}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if api.machineID != "pi-1" || api.attrs.Name != "edge-pi" || lo.FromPtr(api.attrs.ProjectID) != "proj-1" {
		t.Fatalf("registration/cluster attrs = %#v %#v", api.machineID, api.attrs)
	}
	if got.url != "https://agent.example" || got.token != "deploy-token" || got.chart != "oci://chart" || got.id != "cluster-1" {
		t.Fatalf("install = %#v", got)
	}
}

func TestBootstrapIgnoresAlreadyTakenAndPolls(t *testing.T) {
	api := &fakeBootstrapAPI{
		alreadyTaken:  true,
		completeAfter: 3,
		registration:  testRegistration(),
		cluster:       testCluster(),
		ext:           "https://console.example/ext",
	}
	if err := Bootstrap(t.Context(), api, func(string, string, string, string) error { return nil }, BootstrapOptions{MachineID: "pi-1", PollInterval: time.Millisecond}, nil); err != nil {
		t.Fatal(err)
	}
	if api.polls < 3 {
		t.Fatalf("polls = %d", api.polls)
	}
}

func TestBootstrapCancelStopsPoll(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	api := &fakeBootstrapAPI{completeAfter: 100, registration: testRegistration(), cluster: testCluster()}
	err := Bootstrap(ctx, api, func(string, string, string, string) error { return nil }, BootstrapOptions{MachineID: "pi-1", PollInterval: time.Hour}, nil)
	if err == nil {
		t.Fatal("expected cancel error")
	}
}

func TestBootstrapRequiresMachineID(t *testing.T) {
	err := Bootstrap(t.Context(), &fakeBootstrapAPI{}, func(string, string, string, string) error { return nil }, BootstrapOptions{}, nil)
	if err == nil || !strings.Contains(err.Error(), "machine-id") {
		t.Fatalf("error = %v", err)
	}
}

func TestBootstrapRequiresDeployToken(t *testing.T) {
	api := &fakeBootstrapAPI{
		registration: testRegistration(),
		cluster:      &gqlclient.CreateCluster{CreateCluster: &gqlclient.CreateCluster_CreateCluster{ID: "cluster-1"}},
	}
	err := Bootstrap(t.Context(), api, func(string, string, string, string) error { return nil }, BootstrapOptions{MachineID: "pi-1", PollInterval: time.Millisecond}, nil)
	if err == nil || !strings.Contains(err.Error(), "deploy token") {
		t.Fatalf("error = %v", err)
	}
}
