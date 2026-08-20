package common

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/browser"
	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/provider"
	providerapi "github.com/pluralsh/plural-cli/pkg/provider/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"

	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

var (
	loggedIn       = false
	newAuthService = func() *bridge.AuthService { return bridge.NewAuthService(bridge.PluralAuthFactory{}, 0) }
	openLoginURL   = browser.OpenURL
)

func HandleLogin(c *cli.Context) error {
	if loggedIn {
		return nil
	}
	defer func() {
		loggedIn = true
	}()

	conf := &config.Config{}
	conf.Token = ""
	conf.Endpoint = c.String("endpoint")
	auth := newAuthService()
	ctx := context.Background()
	persist := c.Command.Name == "login"

	if config.Exists() {
		conf := config.Read()
		if Affirm(fmt.Sprintf("It looks like your current Plural user is %s, use this profile?", conf.Email), "PLURAL_LOGIN_AFFIRM_CURRENT_USER") {
			return establishLogin(ctx, auth, &conf, c.String("service-account"), persist)
		}
	}

	device, err := auth.BeginDeviceLogin(ctx, conf.Endpoint)
	if err != nil {
		return authError(err)
	}

	fmt.Printf("logging into Plural at %s\n", device.LoginURL)
	if err := openLoginURL(device.LoginURL); err != nil {
		fmt.Printf("Open %s in your browser to proceed\n", device.LoginURL)
	}

	jwt, err := auth.AwaitDeviceLogin(ctx, conf.Endpoint, device.DeviceToken)
	if err != nil {
		return authError(err)
	}

	conf.Token = jwt
	conf.ReportErrors = Affirm("Would you be willing to report any errors to Plural to help with debugging?", "PLURAL_LOGIN_AFFIRM_REPORT_ERRORS")
	return establishLogin(ctx, auth, conf, c.String("service-account"), persist)
}

func establishLogin(ctx context.Context, auth *bridge.AuthService, conf *config.Config, serviceAccount string, persist bool) error {
	profiles := bridge.LegacyProfileStore{}
	session, err := auth.EstablishSession(ctx, conf.Endpoint, conf.Token, serviceAccount, persist, func(event bridge.AuthEvent) {
		switch event.Kind {
		case bridge.AuthEventIdentified:
			conf.Email = event.Email
			fmt.Printf("\nLogged in as %s!\n", event.Email)
		case bridge.AuthEventImpersonated:
			conf.Email = event.Email
			conf.Token = event.Credential
			fmt.Printf("Assumed service account %s\n", serviceAccount)
			_ = profiles.Activate(ctx, conf)
		}
	})
	if err != nil {
		return authError(err)
	}

	conf.Email = session.EffectiveEmail
	conf.Token = session.Credential
	if session.Impersonated && !persist {
		return profiles.Activate(ctx, conf)
	}
	return profiles.Persist(ctx, conf)
}

func authError(err error) error {
	if appErr, ok := errors.AsType[*bridge.Error](err); ok {
		return api.GetErrorResponse(appErr.Err, string(appErr.Operation))
	}
	return err
}
func Preflights(c *cli.Context) error {
	_, err := RunPreflights(c)
	return err
}

func RunPreflights(c *cli.Context) (providerapi.Provider, error) {
	provider.SetCloudFlag(c.Bool("cloud"))
	provider.SetDryrunFlag(c.Bool("dry-run"))

	prov, err := provider.GetProvider()
	if err != nil {
		return prov, err
	}

	if c.Bool("dry-run") {
		return prov, nil
	}

	for _, pre := range prov.Preflights() {
		if err := pre.Validate(); err != nil {
			return prov, err
		}
	}

	return prov, nil
}

func HandleClone(c *cli.Context) error {
	url := c.Args().Get(0)
	cmd := exec.Command("git", "clone", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	repo := git.RepoName(url)
	_ = os.Chdir(repo)
	if err := CryptoInit(c); err != nil {
		return err
	}

	if err := HandleUnlock(c); err != nil {
		return err
	}

	utils.Success("Your repo has been cloned and decrypted, cd %s to start working\n", repo)
	return nil
}

func HandleImport(c *cli.Context) error {
	dir, err := filepath.Abs(c.Args().Get(0))
	if err != nil {
		return err
	}

	conf := config.Import(pathing.SanitizeFilepath(filepath.Join(dir, "config.yml")))
	if err := conf.Flush(); err != nil {
		return err
	}

	if err := CryptoInit(c); err != nil {
		return err
	}

	data, err := os.ReadFile(pathing.SanitizeFilepath(filepath.Join(dir, "key")))
	if err != nil {
		return err
	}

	key, err := crypto.Import(data)
	if err != nil {
		return err
	}
	if err := key.Flush(); err != nil {
		return err
	}

	utils.Success("Workspace properly imported\n")
	return nil
}

func IsUUIDv4(input string) bool {
	_, err := uuid.Parse(input)
	return err == nil
}

func GetIdAndName(input string) (id, name *string) {
	switch {
	case strings.HasPrefix(input, "@"):
		name = new(strings.Trim(input, "@"))
	case IsUUIDv4(input):
		id = &input
	default:
		name = &input
	}
	return
}

func GetHostnameFromURL(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}
	if parsed.Scheme == "" && parsed.Host == "" {
		if parsed, err = url.Parse("//" + u); err != nil {
			return ""
		}
	}
	hostname := parsed.Hostname()
	return hostname
}
