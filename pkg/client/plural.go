package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/pluralsh/polly/algorithms"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/scm"
	"github.com/pluralsh/plural-cli/pkg/wkspace"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"sigs.k8s.io/yaml"
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

func (p *Plural) HandleInit(c *cli.Context) error {
	gitCreated := false
	repo := ""
	p.InitPluralClient()

	if utils.Exists("./workspace.yaml") {
		utils.Highlight("Found workspace.yaml, skipping init as this repo has already been initialized\n")
		utils.Highlight("Checking domain...\n")
		proj, err := manifest.FetchProject()
		if err != nil {
			return err
		}
		if proj.Network != nil && proj.Network.PluralDns {
			if err := p.Client.CreateDomain(proj.Network.Subdomain); err != nil {
				return err
			}
		}
		utils.Highlight("Domain OK \n")
		branch, err := git.CurrentBranch()
		if err != nil {
			return err
		}
		proj.Context["Branch"] = branch
		if err := proj.Flush(); err != nil {
			return err
		}
		if err := common.CryptoInit(c); err != nil {
			return err
		}
		_ = wkspace.DownloadReadme()
		return nil
	}

	git, err := wkspace.Preflight()
	if err != nil && git {
		return err
	}

	if err := common.HandleLogin(c); err != nil {
		return err
	}

	me, err := p.Me()
	if err != nil {
		return api.GetErrorResponse(err, "Me")
	}
	if me.Demoing {
		return fmt.Errorf("You're currently running a gcp demo cluster. Spin that down by deleting you shell at https://app.plural.sh/shell before beginning a local installation")
	}

	if _, err := os.Stat(manifest.ProjectManifestPath()); err == nil && git && !common.Affirm("This repository's workspace.yaml already exists. Would you like to use it?", "PLURAL_INIT_AFFIRM_CURRENT_REPO") {
		fmt.Println("Run `plural init` from empty repository or outside any in order to start from scratch.")
		return nil
	}

	prov, err := common.RunPreflights(c)
	if err != nil && !c.Bool("ignore-preflights") {
		return err
	}

	if !git && common.Affirm("You're attempting to setup plural outside a git repository. Would you like us to set one up for you here?", "PLURAL_INIT_AFFIRM_SETUP_REPO") {
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

func (p *Plural) DoInstallOperator(url, token, values, chart_loc string) error {
	err := p.InitKube()
	if err != nil {
		return err
	}
	alreadyExists, err := console.IsAlreadyAgentInstalled(p.Kube.GetClient())
	if err != nil {
		return err
	}
	if alreadyExists && !common.Confirm("the deployment operator is already installed. Do you want to replace it", "PLURAL_INSTALL_AGENT_CONFIRM_IF_EXISTS") {
		utils.Success("deployment operator is already installed, skip installation\n")
		return nil
	}

	err = p.Kube.CreateNamespace(console.OperatorNamespace, false)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	vals := map[string]interface{}{}
	globalVals := map[string]interface{}{}
	version := ""

	if p.ConsoleClient != nil {
		settings, err := p.ConsoleClient.GetGlobalSettings()
		if err == nil && settings != nil {
			version = strings.Trim(settings.AgentVsn, "v")
			if settings.AgentHelmValues != nil {
				if err := yaml.Unmarshal([]byte(*settings.AgentHelmValues), &globalVals); err != nil {
					return err
				}
			}
		}
	}

	if values != "" {
		if err := utils.YamlFile(values, &vals); err != nil {
			return err
		}
	}
	vals = algorithms.Merge(vals, globalVals)
	err = console.InstallAgent(url, token, console.OperatorNamespace, version, chart_loc, vals)
	if err == nil {
		utils.Success("deployment operator installed successfully\n")
	}
	return err
}

func (p *Plural) ReinstallOperator(c *cli.Context, id, handle *string, chart_loc string) error {
	deployToken, err := p.ConsoleClient.GetDeployToken(id, handle)
	if err != nil {
		return err
	}

	url := p.ConsoleClient.ExtUrl()
	if cluster, err := p.ConsoleClient.GetCluster(id, handle); err == nil {
		if agentUrl, err := p.ConsoleClient.AgentUrl(cluster.ID); err == nil {
			url = agentUrl
		}
	}

	return p.DoInstallOperator(url, deployToken, c.String("values"), chart_loc)
}
