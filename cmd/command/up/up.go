package up

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	cdpkg "github.com/pluralsh/plural-cli/cmd/command/cd"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/pkg/up"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/samber/lo"
	"github.com/urfave/cli"
)

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:  "up",
		Usage: "sets up your repository and an initial management cluster",
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
			cli.BoolFlag{
				Name:  "cloud",
				Usage: "Whether you're provisioning against a cloud-hosted Plural Console",
			},
			cli.StringFlag{
				Name:  "commit",
				Usage: "commits your changes with this message",
			},
		},
		Action: common.LatestVersion(p.handleUp),
	}
}

func (p *Plural) handleUp(c *cli.Context) error {
	// provider.IgnoreProviders([]string{"GENERIC", "KIND"})
	if err := common.HandleLogin(c); err != nil {
		return err
	}
	p.InitPluralClient()

	cd := &cdpkg.Plural{Plural: p.Plural}

	var name, url string
	var err error

	if c.Bool("cloud") {
		name, url, err = p.choseCluster()
		if err != nil {
			return err
		}

		cdpkg.SetConsoleURL(url)
		provider.SetClusterFlag(name)
		if err := cd.HandleCdLogin(c); err != nil {
			return err
		}

		if err := p.backfillEncryption(); err != nil {
			return err
		}
	}

	if err := p.HandleInit(c); err != nil {
		return err
	}

	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	ctx, err := up.Build(c.Bool("cloud"))
	if err != nil {
		return err
	}

	if c.Bool("cloud") {
		id, err := getCluster(cd)
		if err != nil {
			return err
		}

		ctx.ImportCluster = lo.ToPtr(id)
		ctx.CloudCluster = name
	}

	if err := ctx.Backfill(); err != nil {
		return err
	}

	dir, err := ctx.Generate()
	defer func() { os.RemoveAll(dir) }()
	if err != nil {
		return err
	}

	if !common.Affirm(common.AffirmUp, "PLURAL_UP_AFFIRM_DEPLOY") {
		return fmt.Errorf("cancelled deploy")
	}

	if err := ctx.Deploy(func() error {
		utils.Highlight("\n==> Commit and push your configuration\n\n")
		if commit := common.CommitMsg(c); commit != "" {
			utils.Highlight("Pushing upstream...\n")
			return git.Sync(repoRoot, commit, c.Bool("force"))
		}
		return nil
	}); err != nil {
		return err
	}

	utils.Success("Finished setting up your management cluster!\n")
	utils.Highlight("Feel free to use terraform as you normally would, and leverage the gitops setup we've generated in the apps/ subfolder\n")
	return nil
}

func (p *Plural) choseCluster() (name, url string, err error) {
	instances, err := p.GetConsoleInstances()
	if err != nil {
		return
	}

	clusterNames := []string{}
	clusterMap := map[string]string{}

	for _, cluster := range instances {
		clusterNames = append(clusterNames, cluster.Name)
		clusterMap[cluster.Name] = cluster.URL
	}

	prompt := &survey.Select{
		Message: "Select one of the following clusters:",
		Options: clusterNames,
	}
	if err = survey.AskOne(prompt, &name, survey.WithValidator(survey.Required)); err != nil {
		return
	}
	url = clusterMap[name]
	return
}

func getCluster(cd *cdpkg.Plural) (id string, err error) {
	if cd == nil {
		err = fmt.Errorf("your CLI is not logged into Plural, try running `plural login` to generate local credentials")
		return
	}

	clusters, err := cd.ListClusters()
	if err != nil {
		return
	}

	for _, cluster := range clusters {
		if lo.FromPtr(cluster.Node.Handle) == "mgmt" {
			return cluster.Node.ID, nil
		}
	}

	err = fmt.Errorf("could not find the management cluster in your Plural cloud instance, contact support for assistance")
	return
}
