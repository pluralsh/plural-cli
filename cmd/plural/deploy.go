package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/diff"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/scaffold"
	"github.com/pluralsh/plural/pkg/utils"
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

func diffed(c *cli.Context) error {
	diffed, err := wkspace.DiffedRepos()
	if err != nil {
		return err
	}

	for _, diff := range diffed {
		fmt.Println(diff)
	}

	return nil
}

func build(c *cli.Context) error {
	if err := validateOwner(); err != nil {
		return err
	}

	client := api.NewClient()
	if c.IsSet("only") {
		installation, err := client.GetInstallation(c.String("only"))
		if err != nil {
			return err
		}

		return doBuild(client, installation)
	}

	installations, err := getSortedInstallations("", client)
	if err != nil {
		return err
	}

	for _, installation := range installations {
		if err := doBuild(client, installation); err != nil {
			return err
		}
	}
	return nil
}

func doBuild(client *api.Client, installation *api.Installation) error {
	repoName := installation.Repository.Name
	fmt.Printf("Building workspace for %s\n", repoName)
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

	err = build.Execute(workspace)
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
	if err := validateOwner(); err != nil {
		return err
	}

	client := api.NewClient()
	repoRoot, err := utils.RepoRoot()
	if err != nil {
		return err
	}

	sorted, err := getSortedNames(true)
	if err != nil {
		return err
	}

	fmt.Printf("Deploying applications [%s] in topological order\n\n", strings.Join(sorted, ", "))

	ignoreConsole := c.Bool("ignore-console")
	for _, repo := range sorted {
		if ignoreConsole && (repo == "console" || repo == "bootstrap") {
			continue
		}

		execution, err := executor.GetExecution(filepath.Join(repoRoot, repo), "deploy")
		if err != nil {
			return err
		}

		if err := execution.Execute(); err != nil {
			return err
		}
		fmt.Printf("\n")

		installation, err := client.GetInstallation(repo)
		if err != nil {
			return err
		}
		workspace, err := wkspace.New(client, installation)
		if err != nil {
			return err
		}

		if c.Bool("silence") {
			continue
		}

		if err := scaffold.Notes(workspace); err != nil {
			return err
		}
	}

	utils.Highlight("\n==> Commit and push your changes to record your deployment\n")

	if commit := c.String("commit"); commit != "" {
		utils.Highlight("Pushing upstream...\n")
		return utils.Sync(commit, c.Bool("force"))
	}
	return nil
}

func handleDiff(c *cli.Context) error {
	repoRoot, err := utils.RepoRoot()
	if err != nil {
		return err
	}

	sorted, err := getSortedNames(true)
	if err != nil {
		return err
	}

	fmt.Printf("Diffing applications [%s] in topological order\n\n", strings.Join(sorted, ", "))

	for _, repo := range sorted {
		d, err := diff.GetDiff(filepath.Join(repoRoot, repo), "diff")
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
	if err := validateOwner(); err != nil {
		return err
	}

	client := api.NewClient()
	repoRoot, err := utils.RepoRoot()
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
	workspace.Provider.KubeConfig()

	os.Chdir(filepath.Join(repoRoot, repoName))
	return workspace.Bounce()
}

func destroy(c *cli.Context) error {
	if err := validateOwner(); err != nil {
		return err
	}

	if ok := confirm("Are you sure you want to destroy this workspace?"); !ok {
		return nil
	}

	client := api.NewClient()
	repoName := c.Args().Get(0)
	repoRoot, err := utils.RepoRoot()
	if err != nil {
		return err
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

	utils.Success("Finished destroying workspace\n")
	return nil
}

func doDestroy(repoRoot string, client *api.Client, installation *api.Installation) error {
	os.Chdir(repoRoot)
	utils.Error("\nDestroying workspace %s\n", installation.Repository.Name)
	workspace, err := wkspace.New(client, installation)
	if err != nil {
		return err
	}

	return workspace.Destroy()
}

func buildContext(c *cli.Context) error {
	client := api.NewClient()
	insts, err := client.GetInstallations()
	if err != nil {
		return err
	}

	path := manifest.ContextPath()
	return manifest.BuildContext(path, insts)
}
