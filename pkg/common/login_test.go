package common

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/urfave/cli"
)

type loginFactory struct{ client *loginClient }

func (f loginFactory) New(context.Context, string, string) bridge.AuthClient { return f.client }

type loginClient struct{}

func (*loginClient) DeviceLogin(context.Context) (bridge.DeviceAuthorization, error) {
	return bridge.DeviceAuthorization{LoginURL: "https://example.com/device", DeviceToken: "device"}, nil
}
func (*loginClient) PollLoginToken(context.Context, string) (string, error) {
	return "device-jwt", nil
}
func (*loginClient) CurrentIdentity(context.Context) (string, error) {
	return "dev@example.com", nil
}
func (*loginClient) ImpersonateServiceAccount(context.Context, string) (string, string, error) {
	return "", "", nil
}
func (*loginClient) GrabAccessToken(context.Context) (string, error) {
	return "access-token", nil
}

func TestHandleLoginKeepsLegacyPresentationAndPersistsResult(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PLURAL_LOGIN_AFFIRM_REPORT_ERRORS", "true")

	oldService, oldOpen, oldLoggedIn := newAuthService, openLoginURL, loggedIn
	newAuthService = func() *bridge.AuthService {
		return bridge.NewAuthService(loginFactory{client: &loginClient{}}, time.Millisecond)
	}
	openLoginURL = func(string) error { return nil }
	loggedIn = false
	t.Cleanup(func() {
		newAuthService, openLoginURL, loggedIn = oldService, oldOpen, oldLoggedIn
		config.SetConfig(nil)
	})

	app := cli.NewApp()
	app.Commands = []cli.Command{{
		Name:   "login",
		Action: HandleLogin,
		Flags: []cli.Flag{
			cli.StringFlag{Name: "endpoint"},
			cli.StringFlag{Name: "service-account"},
		},
	}}

	output, err := captureLoginOutput(func() error {
		return app.Run([]string{"plural", "login", "--endpoint", "example.com"})
	})
	if err != nil {
		t.Fatalf("login error = %v", err)
	}
	want := "logging into Plural at https://example.com/device\n\nLogged in as dev@example.com!\n"
	if output != want {
		t.Fatalf("output changed\nwant: %q\n got: %q", want, output)
	}

	stored := config.Import(filepath.Join(home, ".plural", config.ConfigName))
	if stored.Email != "dev@example.com" || stored.Token != "access-token" || stored.Endpoint != "example.com" || !stored.ReportErrors {
		t.Fatalf("stored config = %#v", stored)
	}
}

func captureLoginOutput(run func() error) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	runErr := run()
	_ = w.Close()
	var output bytes.Buffer
	_, copyErr := io.Copy(&output, r)
	_ = r.Close()
	if runErr != nil {
		return "", runErr
	}
	return output.String(), copyErr
}
