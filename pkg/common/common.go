package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/pkg/server"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/scaffold"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

var (
	loggedIn = false
)

func AppReadme(name string, dryRun bool) error {
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	dir := filepath.Join(repoRoot, name, "helm", name)
	return scaffold.Readme(dir, dryRun)
}

func DoBuild(client api.Client, installation *api.Installation, force bool) error {
	repoName := installation.Repository.Name
	fmt.Printf("Building workspace for %s\n", repoName)

	if !wkspace.Configured(repoName) {
		fmt.Printf("You have not locally configured %s but have it registered as an installation in our api, ", repoName)
		fmt.Printf("either delete it with `plural apps uninstall %s` or install it locally via a bundle in `plural bundle list %s`\n", repoName, repoName)
		return nil
	}

	workspace, err := wkspace.New(client, installation)
	if err != nil {
		return err
	}

	vsn, ok := workspace.RequiredCliVsn()
	if ok && !VersionValid(vsn) {
		return fmt.Errorf("Your cli version is not sufficient to complete this build, please update to at least %s", vsn)
	}

	if err := workspace.Prepare(); err != nil {
		return err
	}

	build, err := scaffold.Scaffolds(workspace)
	if err != nil {
		return err
	}

	err = build.Execute(workspace, force)
	if err == nil {
		utils.Success("Finished building %s\n\n", repoName)
	}

	workspace.PrintLinks()

	AppReadme(repoName, false) // nolint:errcheck
	return err
}

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
	client := api.FromConfig(conf)
	persist := c.Command.Name == "login"

	if config.Exists() {
		conf := config.Read()
		if Affirm(fmt.Sprintf("It looks like your current Plural user is %s, use this profile?", conf.Email), "PLURAL_LOGIN_AFFIRM_CURRENT_USER") {
			client = api.FromConfig(&conf)
			return postLogin(&conf, client, c, persist)
		}
	}

	device, err := client.DeviceLogin()
	if err != nil {
		return api.GetErrorResponse(err, "DeviceLogin")
	}

	fmt.Printf("logging into Plural at %s\n", device.LoginUrl)
	if err := browser.OpenURL(device.LoginUrl); err != nil {
		fmt.Printf("Open %s in your browser to proceed\n", device.LoginUrl)
	}

	var jwt string
	for {
		result, err := client.PollLoginToken(device.DeviceToken)
		if err == nil {
			jwt = result
			break
		}

		time.Sleep(2 * time.Second)
	}

	conf.Token = jwt
	conf.ReportErrors = Affirm("Would you be willing to report any errors to Plural to help with debugging?", "PLURAL_LOGIN_AFFIRM_REPORT_ERRORS")
	client = api.FromConfig(conf)
	return postLogin(conf, client, c, persist)
}

func postLogin(conf *config.Config, client api.Client, c *cli.Context, persist bool) error {
	me, err := client.Me()
	if err != nil {
		return api.GetErrorResponse(err, "Me")
	}

	conf.Email = me.Email
	fmt.Printf("\nlogged in as %s!\n", me.Email)

	saEmail := c.String("service-account")
	if saEmail != "" {
		jwt, email, err := client.ImpersonateServiceAccount(saEmail)
		if err != nil {
			return api.GetErrorResponse(err, "ImpersonateServiceAccount")
		}

		conf.Email = email
		conf.Token = jwt
		fmt.Printf("Assumed service account %s\n", saEmail)
		config.SetConfig(conf)
		client = api.FromConfig(conf)
		if !persist {
			return nil
		}
	}

	accessToken, err := client.GrabAccessToken()
	if err != nil {
		return api.GetErrorResponse(err, "GrabAccessToken")
	}

	conf.Token = accessToken
	return conf.Flush()
}
func Preflights(c *cli.Context) error {
	_, err := RunPreflights(c)
	return err
}

func RunPreflights(c *cli.Context) (provider.Provider, error) {
	provider.SetCloudFlag(c.Bool("cloud"))
	prov, err := provider.GetProvider()
	if err != nil {
		return prov, err
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

func DownloadReadme(c *cli.Context) error {
	return wkspace.DownloadReadme()
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

func HandleServe(c *cli.Context) error {
	return server.Run()
}

func GetIdAndName(input string) (id, name *string) {
	if strings.HasPrefix(input, "@") {
		h := strings.Trim(input, "@")
		name = &h
	} else {
		id = &input
	}
	return
}
