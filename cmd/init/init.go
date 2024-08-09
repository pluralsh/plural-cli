package init

import (
	"fmt"
	cmdcrypto "github.com/pluralsh/plural-cli/cmd/crypto"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/scm"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
	"github.com/urfave/cli"
	"os"
)

const DemoingErrorMsg = "You're currently running a gcp demo cluster. Spin that down by deleting you shell at https://app.plural.sh/shell before beginning a local installation"

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:  "init",
		Usage: "initializes plural within a git repo",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "endpoint",
				Usage: "the endpoint for the plural installation you're working with",
			},
			cli.StringFlag{
				Name:  "service-account",
				Usage: "email for the service account you'd like to use for this workspace",
			},
			cli.BoolFlag{
				Name:  "ignore-preflights",
				Usage: "whether to ignore preflight check failures prior to init",
			},
		},
		Action: common.Tracked(common.LatestVersion(p.handleInit), "cli.init"),
	}
}

func (p *Plural) handleInit(c *cli.Context) error {
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

	prov, err := common.RunPreflights()
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
	if err := cmdcrypto.CryptoInit(c); err != nil {
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
