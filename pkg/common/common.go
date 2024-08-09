package common

import (
	"fmt"
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/wkspace"

	"github.com/pluralsh/plural-cli/pkg/scaffold"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

func AppReadme(name string, dryRun bool) error {
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	dir := filepath.Join(repoRoot, name, "helm", name)
	return scaffold.Readme(dir, dryRun)
}

func DoBuild(client api.Client, installation *api.Installation, force bool) error {
	repoName := installation.Repository.Name
	fmt.Printf("Building workspace for %s\n", repoName)

	if !wkspace.Configured(repoName) {
		fmt.Printf("You have not locally configured %s but have it registered as an installation in our api, ", repoName)
		fmt.Printf("either delete it with `plural apps uninstall %s` or install it locally via a bundle in `plural bundle list %s`\n", repoName, repoName)
		return nil
	}

	workspace, err := wkspace.New(client, installation)
	if err != nil {
		return err
	}

	vsn, ok := workspace.RequiredCliVsn()
	if ok && !VersionValid(vsn) {
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

	AppReadme(repoName, false) // nolint:errcheck
	return err
}
