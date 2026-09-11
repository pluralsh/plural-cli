package edge

import (
	"context"
	"errors"
	"strings"
	"testing"

	gqlclient "github.com/pluralsh/console/go/client"

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
