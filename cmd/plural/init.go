package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"github.com/mholt/archiver/v3"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/urfave/cli"
)

const (
	KUBECTL_VERSION = "1.20.5"
	HELM_VERSION = "3.5.3"
	TF_VERSION = "0.15.2"
)

func handleInit(c *cli.Context) error {
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

	email, _ := utils.ReadLine("Enter your email: ")
	pwd, _ := utils.ReadPwd("Enter password: ")
	result, err := client.Login(email, pwd)
	if err != nil {
		return err
	}

	fmt.Printf("\nlogged in as %s\n", email)
	conf.Email = email
	conf.Token = result
	client = api.FromConfig(conf)

	saEmail := c.String("service-account")
	if saEmail != "" {
		jwt, email, err := client.ImpersonateServiceAccount(email)
		if err != nil {
			return err
		}

		conf.Email = email
		conf.Token = jwt
		client = api.FromConfig(conf)
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

func initHelm(success string) error {
	err := exec.Command("helm", "init", "--client-only").Run()
	if err != nil {
		return err
	}
	utils.Success(success)
	return nil
}

func handleInstall(c *cli.Context) (err error) {
	root, found := utils.ProjectRoot()
	if !found {
		root, err = utils.RepoRoot()
		if err != nil { return }
	}

	err = os.MkdirAll(filepath.Join(root, "bin"), os.ModePerm)
	if err != nil { return }
	root = filepath.Join(root, "bin")

	goos := runtime.GOOS
	arch := runtime.GOARCH
	kubectl := fmt.Sprintf("https://dl.k8s.io/release/%s/bin/%s/%s/kubectl", KUBECTL_VERSION, goos, arch)
	err = utils.Install("kubectl", kubectl, filepath.Join(root, "kubectl"), func(dest string) (string, error) { return dest, nil })
	if err != nil {
		return
	}

	helm := fmt.Sprintf("https://get.helm.sh/helm-v%s-%s-%s.tar.gz", HELM_VERSION, goos, arch)
	err = utils.Install("helm", helm, filepath.Join(root, "helm-root.tar.gz"), func(dest string) (bin string, err error) {
		bin = filepath.Join(root, "helm")
		err = archiver.Unarchive(dest, filepath.Join(root, "helm-root"))
		if err != nil {
			return 
		}

		err = os.Rename(filepath.Join(dest, "helm"), bin)
		if err != nil {
			return
		}

		err = os.RemoveAll(dest)
		return
	})

	if err != nil { return }

	tf := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip", TF_VERSION, TF_VERSION, goos, arch)
	err = utils.Install("terraform", tf, filepath.Join(root, "terraform.zip"), func(dest string) (bin string, err error) {
		bin = filepath.Join(root, "terraform")
		err = archiver.Unarchive(dest, bin)
		if err != nil {
			return 
		}

		err = os.Remove(dest)
		return 
	})
	
	if err != nil { return }

	conf := config.Read()
	err = utils.Cmd(&conf, "helm", "plugin" , "install" , "https://github.com/chartmuseum/helm-push")
	if err != nil { return }
	err = utils.Cmd(&conf, "helm", "plugin" , "install" , "https://github.com/databus23/helm-diff")
	if err != nil { return }

	prov, err := provider.Select(true)
	if err != nil { return }
	err = prov.Install()
	return
}