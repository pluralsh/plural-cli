package destroy

import (
	"fmt"
	"os"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
	"github.com/urfave/cli"
)

type Plural struct {
	Plural client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:      "destroy",
		Aliases:   []string{"d"},
		Usage:     "iterates through all installations in reverse topological order, deleting helm installations and terraform",
		ArgsUsage: "APP",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from",
				Usage: "where to start your deploy command (useful when restarting interrupted destroys)",
			},
			cli.StringFlag{
				Name:  "commit",
				Usage: "commits your changes with this message",
			},
			cli.BoolFlag{
				Name:  "force",
				Usage: "use force push when pushing to git",
			},
			cli.BoolFlag{
				Name:  "all",
				Usage: "tear down the entire cluster gracefully in one go",
			},
		},
		Action: common.Tracked(common.LatestVersion(common.Owned(common.UpstreamSynced(p.destroy))), "cli.destroy"),
	}
}

func (p *Plural) destroy(c *cli.Context) error {
	p.Plural.InitPluralClient()
	repoName := c.Args().Get(0)
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}
	force := c.Bool("force")
	all := c.Bool("all")

	project, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	infix := "this workspace"
	if repoName != "" {
		infix = repoName
	} else if !all {
		return fmt.Errorf("you must either specify an individual application or `--all` to destroy the entire workspace")
	}

	if !force && !common.Confirm(fmt.Sprintf("Are you sure you want to destroy %s?", infix), "PLURAL_DESTROY_CONFIRM") {
		return nil
	}

	delete := force || common.Affirm("Do you want to uninstall your applications from the plural api as well?", "PLURAL_DESTROY_AFFIRM_UNINSTALL_APPS")

	if repoName != "" {
		installation, err := p.Plural.GetInstallation(repoName)
		if err != nil {
			return api.GetErrorResponse(err, "GetInstallation")
		}

		if installation == nil {
			return fmt.Errorf("No installation for app %s to destroy, if the app is still in your repo, you can always run cd %s/terraform && terraform destroy", repoName, repoName)
		}

		return p.doDestroy(repoRoot, installation, delete, project.ClusterAPI)
	}

	installations, err := client.GetSortedInstallations(p.Plural, repoName)
	if err != nil {
		return err
	}

	from := c.String("from")
	started := from == ""
	for i := len(installations) - 1; i >= 0; i-- {
		installation := installations[i]
		if installation.Repository.Name == from {
			started = true
		}

		if !started {
			continue
		}

		if err := p.doDestroy(repoRoot, installation, delete, project.ClusterAPI); err != nil {
			return err
		}
	}

	man, _ := manifest.FetchProject()
	if err := p.Plural.DeleteEabCredential(man.Cluster, man.Provider); err != nil {
		fmt.Printf("no eab key to delete %s\n", err)
	}

	if repoName == "" {
		utils.Success("Finished destroying workspace\n")
		utils.Note("if you want to recreate this workspace, be sure to rename the cluster to ensure a clean redeploy")
		man, err := manifest.FetchProject()
		if err != nil {
			return err
		}
		if err := p.Plural.DestroyCluster(man.Network.Subdomain, man.Cluster, man.Provider); err != nil {
			return api.GetErrorResponse(err, "DestroyCluster")
		}
	}

	utils.Highlight("\n==> Commit and push your changes to record your workspace changes\n\n")

	if commit := common.CommitMsg(c); commit != "" {
		utils.Highlight("Pushing upstream...\n")
		return git.Sync(repoRoot, commit, force)
	}

	return nil
}

func (p *Plural) doDestroy(repoRoot string, installation *api.Installation, delete, clusterAPI bool) error {
	p.Plural.InitPluralClient()
	if err := os.Chdir(repoRoot); err != nil {
		return err
	}
	repo := installation.Repository.Name
	if ctx, err := manifest.FetchContext(); err == nil && ctx.Protected(repo) {
		return fmt.Errorf("This app is protected, you cannot plural destroy without updating context.yaml")
	}

	utils.Error("\nDestroying application %s\n", repo)
	workspace, err := wkspace.New(p.Plural.Client, installation)
	if err != nil {
		return err
	}

	// TODO fix for clusterAPI
	// if repo == Bootstrap && clusterAPI {
	//	if err = bootstrap.DestroyCluster(workspace.Destroy, plural.RunPlural); err != nil {
	//		return err
	//	}
	//
	// } else {
	//	if err := workspace.Destroy(); err != nil {
	//		return err
	//	}
	// }

	if err := workspace.Destroy(); err != nil {
		return err
	}

	if delete {
		utils.Highlight("Uninstalling %s from the plural api as well...\n", repo)
		return p.Plural.Client.DeleteInstallation(installation.Id)
	}

	return nil
}
