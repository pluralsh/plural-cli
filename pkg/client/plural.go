package client

import (
	"fmt"
	"os"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

type Plural struct {
	api.Client
	ConsoleClient console.ConsoleClient
	kubernetes.Kube
}

func (p *Plural) InitKube() error {
	if p.Kube == nil {
		kube, err := kubernetes.Kubernetes()
		if err != nil {
			return err
		}
		p.Kube = kube
	}
	return nil
}

func (p *Plural) InitConsoleClient(token, url string) error {
	if p.ConsoleClient == nil {
		if token == "" {
			conf := console.ReadConfig()
			if conf.Token == "" {
				return fmt.Errorf("you have not set up a console login, you can run `plural cd login` to save your credentials")
			}

			token = conf.Token
			url = conf.Url
		}
		consoleClient, err := console.NewConsoleClient(token, url)
		if err != nil {
			return err
		}
		p.ConsoleClient = consoleClient
	}
	return nil
}

func (p *Plural) InitPluralClient() {
	if p.Client == nil {
		if project, err := manifest.FetchProject(); err == nil && config.Exists() {
			conf := config.Read()
			if owner := project.Owner; owner != nil && conf.Email != owner.Email {
				utils.LogInfo().Printf("Trying to impersonate service account: %s \n", owner.Email)
				if err := p.AssumeServiceAccount(conf, project); err != nil {
					os.Exit(1)
				}
				return
			}
		}

		p.Client = api.NewClient()
	}
}

func (p *Plural) AssumeServiceAccount(conf config.Config, man *manifest.ProjectManifest) error {
	owner := man.Owner
	jwt, email, err := api.FromConfig(&conf).ImpersonateServiceAccount(owner.Email)
	if err != nil {
		utils.Error("You (%s) are not the owner of this repo %s, %v \n", conf.Email, owner.Email, api.GetErrorResponse(err, "ImpersonateServiceAccount"))
		return err
	}
	conf.Email = email
	conf.Token = jwt
	p.Client = api.FromConfig(&conf)
	accessToken, err := p.GrabAccessToken()
	if err != nil {
		utils.Error("failed to create access token, bailing")
		return api.GetErrorResponse(err, "GrabAccessToken")
	}
	conf.Token = accessToken
	config.SetConfig(&conf)
	return nil
}
