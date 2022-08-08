package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/application"
	"github.com/pluralsh/plural/pkg/diff"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/scaffold"
	"github.com/pluralsh/plural/pkg/utils"
	pluralerr "github.com/pluralsh/plural/pkg/utils/errors"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/urfave/cli"
)

func (p *Plural) getSortedInstallations(repo string) ([]*api.Installation, error) {
	installations, err := p.GetInstallations()
	if err != nil {
		return installations, err
	}

	if len(installations) == 0 {
		return installations, fmt.Errorf("no installations present, run `plural bundle install <repo> <bundle-name>` to install your first app")
	}

	sorted, err := wkspace.Dependencies(p.Client, repo, installations)
	if err != nil {
		sorted = installations // we don't know all the dependencies yet
	}

	return sorted, nil
}

func (p *Plural) allSortedRepos() ([]string, error) {
	insts, err := p.GetInstallations()
	if err != nil {
		return nil, err
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
		result := make([]string, 0)
		isRepo := map[string]bool{}
		for _, repo := range diffed {
			isRepo[repo] = true
		}

		for _, repo := range sorted {
			if isRepo[repo] {
				result = append(result, repo)
			}
		}

		return result, nil
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
	changed, err := git.HasUpstreamChanges()
	if err != nil {
		return pluralerr.ErrorWrap(errNoGit, "Failed to get git information")
	}

	force := c.Bool("force")
	if !changed && !force {
		return pluralerr.ErrorWrap(errRemoteDiff, "Local Changes out of Sync")
	}

	if c.IsSet("only") {
		installation, err := p.GetInstallation(c.String("only"))
		if err != nil {
			return err
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
		return fmt.Errorf("You have not locally configured %s but have it registered as an installation in our api, either delete it in app.plural.sh or install it locally via a bundle in `plural bundle list %s`", repoName, repoName)
	}

	workspace, err := wkspace.New(p.Client, installation)
	if err != nil {
		return err
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

	return err
}

func (p *Plural) validate(c *cli.Context) error {
	if c.IsSet("only") {
		installation, err := p.GetInstallation(c.String("only"))
		if err != nil {
			return err
		}
		return p.doValidate(installation)
	}

	installations, err := p.getSortedInstallations("")
	if err != nil {
		return err
	}

	for _, installation := range installations {
		if err := p.doValidate(installation); err != nil {
			return err
		}
	}

	utils.Success("Workspace providers are properly configured!\n")
	return nil
}

func (p *Plural) doValidate(installation *api.Installation) error {
	utils.Highlight("Validating repository %s\n", installation.Repository.Name)
	workspace, err := wkspace.New(p.Client, installation)
	if err != nil {
		return err
	}

	return workspace.Validate()
}

func (p *Plural) deploy(c *cli.Context) error {
	verbose := c.Bool("verbose")
	repoRoot, err := git.Root()

	if err != nil {
		return err
	}

	sorted, err := getSortedNames(true)
	if err != nil {
		return err
	}

	if c.Bool("all") {
		sorted, err = p.allSortedRepos()
		if err != nil {
			return err
		}
	}

	fmt.Printf("Deploying applications [%s] in topological order\n\n", strings.Join(sorted, ", "))

	ignoreConsole := c.Bool("ignore-console")
	for _, repo := range sorted {
		if ignoreConsole && (repo == "console" || repo == "bootstrap") {
			continue
		}

		execution, err := executor.GetExecution(pathing.SanitizeFilepath(filepath.Join(repoRoot, repo)), "deploy")
		if err != nil {
			return err
		}

		if err := retry(verbose, 2, time.Second*5, execution.Execute); err != nil {
			utils.Note("It looks like your deployment failed. This may be a transient issue and rerunning the `plural deploy` command may resolve it. Or, feel free to reach out to us on discord (https://discord.gg/bEBAMXV64s) or Intercom and we should be able to help you out\n")
			return err
		}
		fmt.Printf("\n")

		installation, err := p.GetInstallation(repo)
		if err != nil {
			return err
		}

		if c.Bool("silence") {
			continue
		}

		if man, err := fetchManifest(repo); err == nil && man.Wait {
			if kubeConf, err := kubernetes.KubeConfig(); err == nil {
				fmt.Println("")
				if err := application.Wait(kubeConf, repo); err != nil {
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
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}
	repoName := c.Args().Get(0)

	if repoName != "" {
		installation, err := p.GetInstallation(repoName)
		if err != nil {
			return err
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
	repoName := installation.Repository.Name
	utils.Warn("bouncing deployments in %s\n", repoName)
	workspace, err := wkspace.New(p.Client, installation)
	if err != nil {
		return err
	}
	if err := workspace.Provider.KubeConfig(); err != nil {
		return err
	}

	if err := os.Chdir(pathing.SanitizeFilepath(filepath.Join(repoRoot, repoName))); err != nil {
		return err
	}
	return workspace.Bounce()
}

func (p *Plural) destroy(c *cli.Context) error {
	repoName := c.Args().Get(0)
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	infix := "this workspace"
	if repoName != "" {
		infix = repoName
	}

	if !confirm(fmt.Sprintf("Are you sure you want to destroy %s?", infix)) {
		return nil
	}

	if repoName != "" {
		installation, err := p.GetInstallation(repoName)
		if err != nil {
			return err
		}

		return p.doDestroy(repoRoot, installation)
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

		if err := p.doDestroy(repoRoot, installation); err != nil {
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
	}

	utils.Highlight("\n==> Commit and push your changes to record your workspace changes\n\n")

	if commit := commitMsg(c); commit != "" {
		utils.Highlight("Pushing upstream...\n")
		return git.Sync(repoRoot, commit, c.Bool("force"))
	}

	utils.Note("To remove your installations in app.plural.sh as well, you can run `plural repos reset`")
	return nil
}

func (p *Plural) doDestroy(repoRoot string, installation *api.Installation) error {
	if err := os.Chdir(repoRoot); err != nil {
		return err
	}
	utils.Error("\nDestroying application %s\n", installation.Repository.Name)
	workspace, err := wkspace.New(p.Client, installation)
	if err != nil {
		return err
	}

	return workspace.Destroy()
}

func (p *Plural) buildContext(_ *cli.Context) error {
	insts, err := p.GetInstallations()
	if err != nil {
		return err
	}

	path := manifest.ContextPath()
	return manifest.BuildContext(path, insts)
}

func fetchManifest(repo string) (*manifest.Manifest, error) {
	p, err := manifest.ManifestPath(repo)
	if err != nil {
		return nil, err
	}

	return manifest.Read(p)
}

func retry(verbose bool, attempts int, sleep time.Duration, fn func(bool) error) error {
	if err := fn(verbose); err != nil {
		var stopError *stop
		if errors.As(err, &stopError) {
			// Return the original error for later checking
			return err
		}

		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			fmt.Printf("retrying, number of attempts remaining: %d\n", attempts)
			return retry(verbose, attempts, 2*sleep, fn)
		}
		return err
	}
	return nil
}

type stop struct {
	error
}
