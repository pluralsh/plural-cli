package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pkg/browser"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/pluralsh/plural/pkg/server"
	"github.com/urfave/cli"
)

const (
	KUBECTL_VERSION = "1.20.5"
	HELM_VERSION    = "3.5.3"
	TF_VERSION      = "0.15.2"
)

func handleInit(c *cli.Context) error {
	if err := wkspace.Preflight(); err != nil {
		return err
	}

	if err := handleLogin(c); err != nil {
		return err
	}

	if err := cryptoInit(c); err != nil {
		return err
	}

	manifestPath, _ := filepath.Abs("manifest.yaml")
	if _, err := provider.Bootstrap(manifestPath, false); err != nil {
		return err
	}

	utils.Success("Workspace is properly configured!\n")
	return nil
}

func handleLogin(c *cli.Context) error {
	conf := &config.Config{}
	conf.Token = ""
	conf.Endpoint = c.String("endpoint")
	client := api.FromConfig(conf)

	device, err := client.DeviceLogin()
	if err != nil {
		return err
	}

	fmt.Printf("logging in at %s\n", device.LoginUrl)
	if err := browser.OpenURL(device.LoginUrl); err != nil {
		fmt.Println("Open %s in your browser to proceed")
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
	client = api.FromConfig(conf)
	me, err := client.Me()

	fmt.Printf("\nlogged in as %s!\n", me.Email)
	conf.Email = me.Email
	client = api.FromConfig(conf)

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

	conf := config.Import(filepath.Join(dir, "config.yml"))
	if err := conf.Flush(); err != nil {
		return err
	}

	if err := cryptoInit(c); err != nil {
		return err
	}

	data, err := ioutil.ReadFile(filepath.Join(dir, "key"))
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

func initHelm(success string) error {
	err := exec.Command("helm", "init", "--client-only").Run()
	if err != nil {
		return err
	}
	utils.Success(success)
	return nil
}
