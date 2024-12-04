package client

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/application"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/executor"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/scaffold"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"github.com/pluralsh/polly/algorithms"
	"github.com/pluralsh/polly/containers"
	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
)

const Bootstrap = "bootstrap"

func (p *Plural) Deploy(c *cli.Context) error {
	p.InitPluralClient()
	verbose := c.Bool("verbose")
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	// project, err := manifest.FetchProject()
	// if err != nil {
	//	return err
	// }

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

	if commit := common.CommitMsg(c); commit != "" {
		utils.Highlight("Pushing upstream...\n")
		return git.Sync(repoRoot, commit, c.Bool("force"))
	}

	return nil
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

func fetchManifest(repo string) (*manifest.Manifest, error) {
	p, err := manifest.ManifestPath(repo)
	if err != nil {
		return nil, err
	}

	return manifest.Read(p)
}
