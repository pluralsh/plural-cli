package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/go-homedir"

	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
)

func toConfig(setup *SetupRequest) *config.Config {
	return &config.Config{
		Email:        setup.User.Email,
		Token:        setup.User.AccessToken,
		ReportErrors: true,
	}
}

func toManifest(setup *SetupRequest) *manifest.ProjectManifest {
	wk := setup.Workspace
	return &manifest.ProjectManifest{
		Cluster:      wk.Cluster,
		Bucket:       wk.Bucket,
		Project:      wk.Project,
		Provider:     toProvider(setup.Provider),
		Region:       wk.Region,
		BucketPrefix: wk.BucketPrefix,
		Owner:        &manifest.Owner{Email: setup.User.Email},
		Network: &manifest.NetworkConfig{
			PluralDns: true,
			Subdomain: wk.Subdomain,
		},
		Context: setup.Context,
	}
}

func toContext(setup *SetupRequest) *manifest.Context {
	ctx := manifest.NewContext()
	consoleConf := map[string]interface{}{
		"private_key": setup.SshPrivateKey,
		"public_key":  setup.SshPublicKey,
		"passphrase":  "",
		"repo_url":    setup.GitUrl,
		"console_dns": fmt.Sprintf("console.%s", setup.Workspace.Subdomain),
		"is_demo":     setup.IsDemo,
	}

	if setup.GitInfo != nil {
		consoleConf["git_email"] = setup.GitInfo.Email
		consoleConf["git_user"] = setup.GitInfo.Username
	}

	if setup.User.Name != "" {
		consoleConf["admin_name"] = setup.User.Name
	}

	if setup.User.Email != "" {
		consoleConf["admin_email"] = setup.User.Email
	}

	ctx.Configuration = map[string]map[string]interface{}{
		"console": consoleConf,
	}
	return ctx
}

func setupCli(c *gin.Context) error {
	fmt.Println("Beginning to setup workspace")
	var setup SetupRequest
	if err := c.ShouldBindJSON(&setup); err != nil {
		return err
	}

	p, err := homedir.Expand("~/.plural")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(p, 0755); err != nil {
		return err
	}

	if err := crypto.Setup(setup.AesKey); err != nil {
		return err
	}

	conf := toConfig(&setup)
	if err := conf.Flush(); err != nil {
		return err
	}

	if err := setupProvider(&setup); err != nil {
		return fmt.Errorf("error setting up provider: %w", err)
	}
	marker("cloud")

	exists, err := gitExists()
	if err != nil {
		return err
	}

	if exists {
		c.JSON(http.StatusOK, gin.H{"success": true})
		return nil
	}

	if err := setupGit(&setup); err != nil {
		return fmt.Errorf("error setting up git: %w", err)
	}

	man := toManifest(&setup)
	path := manifest.ProjectManifestPath()
	if !utils.Exists(path) {
		if err := man.Write(path); err != nil {
			return fmt.Errorf("error writing manifest: %w", err)
		}
	} else if setup.Provider == "azure" {
		current, err := manifest.FetchProject()
		if err != nil {
			return err
		}

		current.Project = man.Project
		if err := current.Write(path); err != nil {
			return fmt.Errorf("error writing manifest: %w", err)
		}
	}

	ctx := toContext(&setup)
	path = manifest.ContextPath()
	if !utils.Exists(path) {
		if err := ctx.Write(path); err != nil {
			return fmt.Errorf("error writing context: %w", err)
		}
	}

	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}
	missing, err := prov.Permissions()
	if err != nil {
		return err
	}

	// try to initialize kubeconfig if we can, but don't stress if it fails
	_ = execCmd("plural", "wkspace", "kube-init")
	c.JSON(http.StatusOK, gin.H{"success": true, "missing": missing})
	return nil
}
