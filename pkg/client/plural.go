package client

import (
	"fmt"
	"os"

	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/scm"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

const DemoingErrorMsg = "You're currently running a gcp demo cluster. Spin that down by deleting you shell at https://app.plural.sh/shell before beginning a local installation"

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

func (p *Plural) HandleInit(c *cli.Context) error {
	gitCreated := false
	repo := ""

	if utils.Exists("./workspace.yaml") {
		utils.Highlight("Found workspace.yaml, skipping init as this repo has already been initialized...\n")
		return nil
	}

	git, err := wkspace.Preflight()
	if err != nil && git {
		return err
	}

	if err := common.HandleLogin(c); err != nil {
		return err
	}
	p.InitPluralClient()

	me, err := p.Me()
	if err != nil {
		return api.GetErrorResponse(err, "Me")
	}
	if me.Demoing {
		return fmt.Errorf(DemoingErrorMsg)
	}

	if _, err := os.Stat(manifest.ProjectManifestPath()); err == nil && git && !common.Affirm("This repository's workspace.yaml already exists. Would you like to use it?", "PLURAL_INIT_AFFIRM_CURRENT_REPO") {
		fmt.Println("Run `plural init` from empty repository or outside any in order to start from scratch.")
		return nil
	}

	prov, err := common.RunPreflights(c)
	if err != nil && !c.Bool("ignore-preflights") {
		return err
	}

	if !git && common.Affirm("you're attempting to setup plural outside a git repository. would you like us to set one up for you here?", "PLURAL_INIT_AFFIRM_SETUP_REPO") {
		repo, err = scm.Setup()
		if err != nil {
			return err
		}
		gitCreated = true
	}
	if !git && !gitCreated {
		return fmt.Errorf("you're not in a git repository, either clone one directly or let us set it up for you by rerunning `plural init`")
	}

	// create workspace.yaml when git repository is ready
	if err := prov.Flush(); err != nil {
		return err
	}
	if err := common.CryptoInit(c); err != nil {
		return err
	}
	_ = wkspace.DownloadReadme()

	if common.Affirm(common.BackupMsg, "PLURAL_INIT_AFFIRM_BACKUP_KEY") {
		if err := crypto.BackupKey(p.Client); err != nil {
			return api.GetErrorResponse(err, "BackupKey")
		}
	}

	if err := crypto.CreateKeyFingerprintFile(); err != nil {
		return err
	}

	utils.Success("Workspace is properly configured!\n")
	if gitCreated {
		utils.Highlight("Be sure to `cd %s` to use your configured git repo\n", repo)
	}
	return nil
}
