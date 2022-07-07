package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/pkg/browser"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/scm"
	"github.com/pluralsh/plural/pkg/server"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/urfave/cli"
)

func handleInit(c *cli.Context) error {
	git, err := wkspace.Preflight()
	gitCreated := false
	repo := ""
	if err != nil && git {
		return err
	}

	if err := handleLogin(c); err != nil {
		return err
	}

	prov, err := runPreflights()
	if err != nil {
		return err
	}
	defer func(prov provider.Provider) {
		_ = prov.Flush()

	}(prov)

	if !git && affirm("you're attempting to setup plural outside a git repository. would you like us to set one up for you here?") {
		repo, err = scm.Setup()
		if err != nil {
			return err
		}
		gitCreated = true
	} else if !git && err != nil {
		return err
	}

	if err := cryptoInit(c); err != nil {
		return err
	}

	utils.Success("Workspace is properly configured!\n")
	if gitCreated {
		utils.Highlight("Be sure to `cd %s` to use your configured git repo\n", repo)
	}
	return nil
}

func preflights(c *cli.Context) error {
	_, err := runPreflights()
	return err
}

func runPreflights() (provider.Provider, error) {
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

func handleLogin(c *cli.Context) error {
	conf := &config.Config{}
	conf.Token = ""
	conf.Endpoint = c.String("endpoint")
	client := api.FromConfig(conf)

	if config.Exists() {
		conf := config.Read()
		if !confirm(fmt.Sprintf("It looks like your current Plural user is %s, use a different profile?", conf.Email)) {
			client = api.FromConfig(&conf)
			return postLogin(&conf, client, c)
		}
	}

	device, err := client.DeviceLogin()
	if err != nil {
		return err
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
	conf.ReportErrors = affirm("Would you be willing to report any errors to Plural to help with debugging?")
	client = api.FromConfig(conf)
	return postLogin(conf, client, c)
}

func postLogin(conf *config.Config, client *api.Client, c *cli.Context) error {
	me, err := client.Me()
	if err != nil {
		return err
	}

	conf.Email = me.Email
	fmt.Printf("\nlogged in as %s!\n", me.Email)

	saEmail := c.String("service-account")
	if saEmail != "" {
		jwt, email, err := client.ImpersonateServiceAccount(saEmail)
		if err != nil {
			return err
		}

		conf.Email = email
		conf.Token = jwt
		client = api.FromConfig(conf)
		fmt.Printf("Assumed service account %s\n", saEmail)
	}

	accessToken, err := client.GrabAccessToken()
	if err != nil {
		return err
	}

	conf.Token = accessToken
	return conf.Flush()
}

func handleImport(c *cli.Context) error {
	dir, err := filepath.Abs(c.Args().Get(0))
	if err != nil {
		return err
	}

	conf := config.Import(pathing.SanitizeFilepath(filepath.Join(dir, "config.yml")))
	if err := conf.Flush(); err != nil {
		return err
	}

	if err := cryptoInit(c); err != nil {
		return err
	}

	data, err := ioutil.ReadFile(pathing.SanitizeFilepath(filepath.Join(dir, "key")))
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

func handleServe(c *cli.Context) error {
	return server.Run()
}
