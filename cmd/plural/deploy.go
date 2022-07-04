package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/application"
	"github.com/pluralsh/plural/pkg/diff"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/scaffold"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/errors"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/urfave/cli"
)

func getSortedInstallations(repo string, client *api.Client) ([]*api.Installation, error) {
	installations, err := client.GetInstallations()
	if err != nil {
		return installations, err
	}

	if len(installations) == 0 {
		return installations, fmt.Errorf("no installations present, run `plural bundle install <repo> <bundle-name>` to install your first app")
	}

	sorted, err := wkspace.Dependencies(repo, installations)
	if err != nil {
		sorted = installations // we don't know all the dependencies yet
	}

	return sorted, nil
}

func allSortedRepos(client *api.Client) ([]string, error) {
	insts, err := client.GetInstallations()
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

func diffed(*cli.Context) error {
	diffed, err := wkspace.DiffedRepos()
	if err != nil {
		return err
	}

	for _, d := range diffed {
		fmt.Println(d)
	}

	return nil
}

func build(c *cli.Context) error {
	changed, err := git.HasUpstreamChanges()
	if err != nil {
		return errors.ErrorWrap(noGit, "Failed to get git information")
	}

	force := c.Bool("force")
	if !changed && !force {
		return errors.ErrorWrap(remoteDiff, "Local Changes out of Sync")
	}

	client := api.NewClient()
	if c.IsSet("only") {
		installation, err := client.GetInstallation(c.String("only"))
		if err != nil {
			return err
		} else if installation == nil {
			return utils.HighlightError(fmt.Errorf("%s is not installed. Please install it with `plural bundle install`", c.String("only")))
		}

		return doBuild(client, installation, force)
	}

	installations, err := getSortedInstallations("", client)
	if err != nil {
		return err
	}

	for _, installation := range installations {
		if err := doBuild(client, installation, force); err != nil {
			return err
		}
	}
	return nil
}

func doBuild(client *api.Client, installation *api.Installation, force bool) error {
	repoName := installation.Repository.Name
	fmt.Printf("Building workspace for %s\n", repoName)

	if !wkspace.Configured(repoName) {
		return fmt.Errorf("You have not locally configured %s but have it registered as an installation in our api, either delete it in app.plural.sh or install it locally via a bundle in `plural bundle list %s`", repoName, repoName)
	}

	workspace, err := wkspace.New(client, installation)
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

func validate(c *cli.Context) error {
	client := api.NewClient()
	if c.IsSet("only") {
		installation, err := client.GetInstallation(c.String("only"))
		if err != nil {
			return err
		}
		return doValidate(client, installation)
	}

	installations, err := getSortedInstallations("", client)
	if err != nil {
		return err
	}

	for _, installation := range installations {
		if err := doValidate(client, installation); err != nil {
			return err
		}
	}

	utils.Success("Workspace providers are properly configured!\n")
	return nil
}

func doValidate(client *api.Client, installation *api.Installation) error {
	utils.Highlight("Validating repository %s\n", installation.Repository.Name)
	workspace, err := wkspace.New(client, installation)
	if err != nil {
		return err
	}

	return workspace.Validate()
}

func deploy(c *cli.Context) error {
	verbose := c.Bool("verbose")
	client := api.NewClient()
	repoRoot, err := git.Root()

	if err != nil {
		return err
	}

	sorted, err := getSortedNames(true)
	if err != nil {
		return err
	}

	if c.Bool("all") {
		sorted, err = allSortedRepos(client)
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

		if err := execution.Execute(verbose); err != nil {
			utils.Note("It looks like your deployment failed, feel free to reach out to us on discord (https://discord.gg/bEBAMXV64s) or Intercom and we should be able to help you out\n")
			return err
		}
		fmt.Printf("\n")

		installation, err := client.GetInstallation(repo)
		if err != nil {
			return err
		}

		if c.Bool("silence") {
			continue
		}

		if man, err := fetchManifest(repo); err == nil && man.Wait {
			if kubeConf, err := utils.KubeConfig(); err == nil {
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
		err := survey.AskOne(&survey.Input{Message: "Enter a commit message (empty to not commit right now)"}, &commit)
		if err != nil {
			return ""
		}
		return commit
	}

	return ""
}

func handleDiff(*cli.Context) error {
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

func bounce(c *cli.Context) error {
	client := api.NewClient()
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}
	repoName := c.Args().Get(0)

	if repoName != "" {
		installation, err := client.GetInstallation(repoName)
		if err != nil {
			return err
		}
		return doBounce(repoRoot, client, installation)
	}

	installations, err := getSortedInstallations(repoName, client)
	if err != nil {
		return err
	}

	for _, installation := range installations {
		if err := doBounce(repoRoot, client, installation); err != nil {
			return err
		}
	}
	return nil
}

func doBounce(repoRoot string, client *api.Client, installation *api.Installation) error {
	repoName := installation.Repository.Name
	utils.Warn("bouncing deployments in %s\n", repoName)
	workspace, err := wkspace.New(client, installation)
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

func destroy(c *cli.Context) error {
	client := api.NewClient()
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
		installation, err := client.GetInstallation(repoName)
		if err != nil {
			return err
		}

		return doDestroy(repoRoot, client, installation)
	}

	installations, err := getSortedInstallations(repoName, client)
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

		if err := doDestroy(repoRoot, client, installation); err != nil {
			return err
		}
	}

	man, _ := manifest.FetchProject()
	if err := client.DeleteEabCredential(man.Cluster, man.Provider); err != nil {
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

func doDestroy(repoRoot string, client *api.Client, installation *api.Installation) error {
	err := os.Chdir(repoRoot)
	if err != nil {
		return err
	}
	utils.Error("\nDestroying application %s\n", installation.Repository.Name)
	workspace, err := wkspace.New(client, installation)
	if err != nil {
		return err
	}

	return workspace.Destroy()
}

func buildContext(*cli.Context) error {
	client := api.NewClient()
	insts, err := client.GetInstallations()
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
