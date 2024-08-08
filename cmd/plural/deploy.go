package plural

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/common"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/application"
	"github.com/pluralsh/plural-cli/pkg/bootstrap"
	"github.com/pluralsh/plural-cli/pkg/diff"
	"github.com/pluralsh/plural-cli/pkg/executor"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/scaffold"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/errors"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
	"github.com/pluralsh/polly/algorithms"
	"github.com/pluralsh/polly/containers"
	"github.com/urfave/cli"
)

const Bootstrap = "bootstrap"

func (p *Plural) getSortedInstallations(repo string) ([]*api.Installation, error) {
	p.InitPluralClient()
	installations, err := p.GetInstallations()
	if err != nil {
		return installations, api.GetErrorResponse(err, "GetInstallations")
	}

	if len(installations) == 0 {
		return installations, fmt.Errorf("no installations present, run `plural bundle install <repo> <bundle-name>` to install your first app")
	}

	sorted, err := wkspace.UntilRepo(p.Client, repo, installations)
	if err != nil {
		sorted = installations // we don't know all the dependencies yet
	}

	return sorted, nil
}

func (p *Plural) allSortedRepos() ([]string, error) {
	p.InitPluralClient()
	insts, err := p.GetInstallations()
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetInstallations")
	}

	return wkspace.SortAndFilter(insts)
}

func getSortedNames(filter bool) ([]string, error) {
	diffed, err := wkspace.DiffedRepos()
	if err != nil {
		return nil, err
	}

	sorted, err := wkspace.TopSortNames(diffed)
	if err != nil {
		return nil, err
	}

	if filter {
		repos := containers.ToSet(diffed)
		return algorithms.Filter(sorted, repos.Has), nil
	}

	return sorted, nil
}

func diffed(_ *cli.Context) error {
	diffed, err := wkspace.DiffedRepos()
	if err != nil {
		return err
	}

	for _, d := range diffed {
		fmt.Println(d)
	}

	return nil
}

func (p *Plural) build(c *cli.Context) error {
	p.InitPluralClient()
	force := c.Bool("force")
	if err := CheckGitCrypt(c); err != nil {
		return errors.ErrorWrap(errNoGit, "Failed to scan your repo for secrets to encrypt them")
	}

	if c.IsSet("only") {
		installation, err := p.GetInstallation(c.String("only"))
		if err != nil {
			return api.GetErrorResponse(err, "GetInstallation")
		} else if installation == nil {
			return utils.HighlightError(fmt.Errorf("%s is not installed. Please install it with `plural bundle install`", c.String("only")))
		}

		return p.doBuild(installation, force)
	}

	installations, err := p.getSortedInstallations("")
	if err != nil {
		return err
	}

	for _, installation := range installations {
		if err := p.doBuild(installation, force); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plural) doBuild(installation *api.Installation, force bool) error {
	repoName := installation.Repository.Name
	fmt.Printf("Building workspace for %s\n", repoName)

	if !wkspace.Configured(repoName) {
		fmt.Printf("You have not locally configured %s but have it registered as an installation in our api, ", repoName)
		fmt.Printf("either delete it with `plural apps uninstall %s` or install it locally via a bundle in `plural bundle list %s`\n", repoName, repoName)
		return nil
	}

	workspace, err := wkspace.New(p.Client, installation)
	if err != nil {
		return err
	}

	vsn, ok := workspace.RequiredCliVsn()
	if ok && !common.VersionValid(vsn) {
		return fmt.Errorf("Your cli version is not sufficient to complete this build, please update to at least %s", vsn)
	}

	if err := workspace.Prepare(); err != nil {
		return err
	}

	build, err := scaffold.Scaffolds(workspace)
	if err != nil {
		return err
	}

	err = build.Execute(workspace, force)
	if err == nil {
		utils.Success("Finished building %s\n\n", repoName)
	}

	workspace.PrintLinks()

	appReadme(repoName, false) // nolint:errcheck
	return err
}

func (p *Plural) info(c *cli.Context) error {
	p.InitPluralClient()
	repo := c.Args().Get(0)
	installation, err := p.GetInstallation(repo)
	if err != nil {
		return api.GetErrorResponse(err, "GetInstallation")
	}
	if installation == nil {
		return fmt.Errorf("You have not installed %s", repo)
	}

	return scaffold.Notes(installation)
}

func (p *Plural) deploy(c *cli.Context) error {
	p.InitPluralClient()
	verbose := c.Bool("verbose")
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	project, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	var sorted []string
	switch {
	case len(c.StringSlice("from")) > 0:
		sorted, err = wkspace.AllDependencies(c.StringSlice("from"))
	case c.Bool("all"):
		sorted, err = p.allSortedRepos()
	default:
		sorted, err = getSortedNames(true)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Deploying applications [%s] in topological order\n\n", strings.Join(sorted, ", "))

	ignoreConsole := c.Bool("ignore-console")
	for _, repo := range sorted {
		if ignoreConsole && (repo == "console" || repo == Bootstrap) {
			continue
		}

		if repo == Bootstrap && project.ClusterAPI {
			ready, err := bootstrap.CheckClusterReadiness(project.Cluster, Bootstrap)

			// Stop if cluster exists, but it is not ready yet.
			if err != nil && err.Error() == bootstrap.ClusterNotReadyError {
				return err
			}

			// If cluster does not exist bootstrap needs to be done first.
			if !ready {
				err := bootstrap.BootstrapCluster(RunPlural)
				if err != nil {
					return err
				}
			}
		}

		execution, err := executor.GetExecution(pathing.SanitizeFilepath(filepath.Join(repoRoot, repo)), "deploy")
		if err != nil {
			return err
		}

		if err := execution.Execute("deploying", verbose); err != nil {
			utils.Note("It looks like your deployment failed. This may be a transient issue and rerunning the `plural deploy` command may resolve it. Or, feel free to reach out to us on discord (https://discord.gg/bEBAMXV64s) or Intercom and we should be able to help you out\n")
			return err
		}

		fmt.Printf("\n")

		installation, err := p.GetInstallation(repo)
		if err != nil {
			return api.GetErrorResponse(err, "GetInstallation")
		}
		if installation == nil {
			return fmt.Errorf("The %s was unistalled, run `plural bundle install %s <bundle-name>` ", repo, repo)
		}

		if err := p.Client.MarkSynced(repo); err != nil {
			utils.Warn("failed to mark %s as synced, this is not a critical error but might drift state in our api, you can run `plural repos synced %s` to mark it manually", repo, repo)
		}

		if c.Bool("silence") {
			continue
		}

		if man, err := fetchManifest(repo); err == nil && man.Wait {
			if kubeConf, err := kubernetes.KubeConfig(); err == nil {
				fmt.Printf("Waiting for %s to become ready...\n", repo)
				if err := application.SilentWait(kubeConf, repo); err != nil {
					return err
				}
				fmt.Println("")
			}
		}

		if err := scaffold.Notes(installation); err != nil {
			return err
		}
	}

	utils.Highlight("\n==> Commit and push your changes to record your deployment\n\n")

	if commit := commitMsg(c); commit != "" {
		utils.Highlight("Pushing upstream...\n")
		return git.Sync(repoRoot, commit, c.Bool("force"))
	}

	return nil
}

func commitMsg(c *cli.Context) string {
	if commit := c.String("commit"); commit != "" {
		return commit
	}

	if !c.Bool("silence") {
		var commit string
		if err := survey.AskOne(&survey.Input{Message: "Enter a commit message (empty to not commit right now)"}, &commit); err != nil {
			return ""
		}
		return commit
	}

	return ""
}

func handleDiff(_ *cli.Context) error {
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	sorted, err := getSortedNames(true)
	if err != nil {
		return err
	}

	fmt.Printf("Diffing applications [%s] in topological order\n\n", strings.Join(sorted, ", "))

	for _, repo := range sorted {
		d, err := diff.GetDiff(pathing.SanitizeFilepath(filepath.Join(repoRoot, repo)), "diff")
		if err != nil {
			return err
		}

		if err := d.Execute(); err != nil {
			return err
		}

		fmt.Printf("\n")
	}
	return nil
}

func (p *Plural) bounce(c *cli.Context) error {
	p.InitPluralClient()
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}
	repoName := c.Args().Get(0)

	if repoName != "" {
		installation, err := p.GetInstallation(repoName)
		if err != nil {
			return api.GetErrorResponse(err, "GetInstallation")
		}
		return p.doBounce(repoRoot, installation)
	}

	installations, err := p.getSortedInstallations(repoName)
	if err != nil {
		return err
	}

	for _, installation := range installations {
		if err := p.doBounce(repoRoot, installation); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plural) doBounce(repoRoot string, installation *api.Installation) error {
	p.InitPluralClient()
	repoName := installation.Repository.Name
	utils.Warn("bouncing deployments in %s\n", repoName)
	workspace, err := wkspace.New(p.Client, installation)
	if err != nil {
		return err
	}

	if err := os.Chdir(pathing.SanitizeFilepath(filepath.Join(repoRoot, repoName))); err != nil {
		return err
	}
	return workspace.Bounce()
}

func (p *Plural) destroy(c *cli.Context) error {
	p.InitPluralClient()
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

	if !force && !confirm(fmt.Sprintf("Are you sure you want to destroy %s?", infix), "PLURAL_DESTROY_CONFIRM") {
		return nil
	}

	delete := force || affirm("Do you want to uninstall your applications from the plural api as well?", "PLURAL_DESTROY_AFFIRM_UNINSTALL_APPS")

	if repoName != "" {
		installation, err := p.GetInstallation(repoName)
		if err != nil {
			return api.GetErrorResponse(err, "GetInstallation")
		}

		if installation == nil {
			return fmt.Errorf("No installation for app %s to destroy, if the app is still in your repo, you can always run cd %s/terraform && terraform destroy", repoName, repoName)
		}

		return p.doDestroy(repoRoot, installation, delete, project.ClusterAPI)
	}

	installations, err := p.getSortedInstallations(repoName)
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
	if err := p.DeleteEabCredential(man.Cluster, man.Provider); err != nil {
		fmt.Printf("no eab key to delete %s\n", err)
	}

	if repoName == "" {
		utils.Success("Finished destroying workspace\n")
		utils.Note("if you want to recreate this workspace, be sure to rename the cluster to ensure a clean redeploy")
		man, err := manifest.FetchProject()
		if err != nil {
			return err
		}
		if err := p.DestroyCluster(man.Network.Subdomain, man.Cluster, man.Provider); err != nil {
			return api.GetErrorResponse(err, "DestroyCluster")
		}
	}

	utils.Highlight("\n==> Commit and push your changes to record your workspace changes\n\n")

	if commit := commitMsg(c); commit != "" {
		utils.Highlight("Pushing upstream...\n")
		return git.Sync(repoRoot, commit, force)
	}

	return nil
}

func (p *Plural) doDestroy(repoRoot string, installation *api.Installation, delete, clusterAPI bool) error {
	p.InitPluralClient()
	if err := os.Chdir(repoRoot); err != nil {
		return err
	}
	repo := installation.Repository.Name
	if ctx, err := manifest.FetchContext(); err == nil && ctx.Protected(repo) {
		return fmt.Errorf("This app is protected, you cannot plural destroy without updating context.yaml")
	}

	utils.Error("\nDestroying application %s\n", repo)
	workspace, err := wkspace.New(p.Client, installation)
	if err != nil {
		return err
	}

	if repo == Bootstrap && clusterAPI {
		if err = bootstrap.DestroyCluster(workspace.Destroy, RunPlural); err != nil {
			return err
		}

	} else {
		if err := workspace.Destroy(); err != nil {
			return err
		}
	}

	if delete {
		utils.Highlight("Uninstalling %s from the plural api as well...\n", repo)
		return p.Client.DeleteInstallation(installation.Id)
	}

	return nil
}

func fetchManifest(repo string) (*manifest.Manifest, error) {
	p, err := manifest.ManifestPath(repo)
	if err != nil {
		return nil, err
	}

	return manifest.Read(p)
}
