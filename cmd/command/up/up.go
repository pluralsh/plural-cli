package up

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural-cli/cmd/command/cd"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
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
	if err := p.HandleInit(c); err != nil {
		return err
	}
	p.InitPluralClient()

	cd := &cd.Plural{Plural: p.Plural}

	if c.Bool("cloud") {
		if err := cd.HandleCdLogin(c); err != nil {
			return err
		}

		if err := p.backfillEncryption(); err != nil {
			return err
		}
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
		id, name, err := getCluster(cd)
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

func getCluster(cd *cd.Plural) (id string, name string, err error) {
	if cd == nil {
		return "", "", fmt.Errorf("please provide a plural client")
	}
	clusters, err := cd.ListClusters()
	if err != nil {
		return "", "", err
	}

	clusterNames := []string{}
	clusterMap := map[string]string{}

	for _, cluster := range clusters {
		clusterNames = append(clusterNames, cluster.Node.Name)
		clusterMap[cluster.Node.Name] = cluster.Node.ID
	}

	prompt := &survey.Select{
		Message: "Select one of the following clusters:",
		Options: clusterNames,
	}
	if err = survey.AskOne(prompt, &name, survey.WithValidator(survey.Required)); err != nil {
		return
	}
	id = clusterMap[name]
	return
}
