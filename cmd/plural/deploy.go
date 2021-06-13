package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/diff"
	"github.com/pluralsh/plural/pkg/scaffold"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/urfave/cli"
)

func getSortedInstallations(repo string, client *api.Client) ([]*api.Installation, error) {
	installations, err := client.GetInstallations()
	if err != nil {
		return installations, err
	}

	sorted, err := wkspace.Dependencies(repo, installations)
	if err != nil {
		sorted = installations // we don't know all the dependencies yet
	}

	return sorted, nil
}

func build(c *cli.Context) error {
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
	client := api.NewClient()
	repoRoot, err := utils.RepoRoot()
	repoName := c.Args().Get(0)
	if err != nil {
		return err
	}

	if repoName != "" {
		installation, err := client.GetInstallation(repoName)
		if err != nil {
			return err
		}

		return doDeploy(repoRoot, installation)
	}

	installations, err := getSortedInstallations(repoName, client)
	if err != nil {
		return err
	}

	for _, installation := range installations {
		if err := doDeploy(repoRoot, installation); err != nil {
			return err
		}
		fmt.Printf("\n")
	}
	return nil
}

func doDeploy(repoRoot string, installation *api.Installation) error {
	name := installation.Repository.Name
	execution, err := executor.GetExecution(filepath.Join(repoRoot, name), "deploy")
	if err != nil {
		return err
	}

	return execution.Execute()
}

func handleDiff(c *cli.Context) error {
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

		return doDiff(repoRoot, installation)
	}

	installations, err := getSortedInstallations("", client)
	if err != nil {
		return err
	}
	
	for _, installation := range installations {
		if err := doDiff(repoRoot, installation); err != nil {
			return err
		}
		fmt.Printf("\n")
	}
	return nil
}

func doDiff(repoRoot string, installation *api.Installation) error {
	name := installation.Repository.Name

	d, err := diff.GetDiff(filepath.Join(repoRoot, name), "diff")
	if err != nil {
		return err
	}

	return d.Execute()
}

func bounce(c *cli.Context) error {
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
	client := api.NewClient()
	repoName := c.Args().Get(0)

	if repoName != "" {
		installation, err := client.GetInstallation(repoName)
		if err != nil {
			return err
		}

		return doDestroy(client, installation)
	}
	
	installations, err := getSortedInstallations(repoName, client)
	if err != nil {
		return err
	}

	for i := len(installations) - 1; i >= 0; i-- {
		installation := installations[i]
		if err := doDestroy(client, installation); err != nil {
			return err
		}
	}

	utils.Success("Finished destroying workspace")
	return nil
}

func doDestroy(client *api.Client, installation *api.Installation) error {
	dir, _ := os.Getwd()
	os.Chdir(dir)
	utils.Warn("Destroying workspace %s\n", installation.Repository.Name)
	workspace, err := wkspace.New(client, installation)
	if err != nil {
		return err
	}

	if err := workspace.DestroyHelm(); err != nil {
		return err
	}

	return workspace.DestroyTerraform()
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