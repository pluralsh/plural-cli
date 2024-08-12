package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/diff"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
	"github.com/pluralsh/polly/algorithms"
	"github.com/pluralsh/polly/containers"
	"github.com/urfave/cli"
)

func Diffed(_ *cli.Context) error {
	diffed, err := wkspace.DiffedRepos()
	if err != nil {
		return err
	}

	for _, d := range diffed {
		fmt.Println(d)
	}

	return nil
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

func HandleDiff(_ *cli.Context) error {
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
