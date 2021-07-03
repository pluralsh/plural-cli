package wkspace

import (
	"fmt"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	toposort "github.com/philopon/go-toposort"
)

func TopSort(installations []*api.Installation) ([]*api.Installation, error) {
	var repoMap = make(map[string]*api.Installation)
	names := make([]string, len(installations))
	for i, installation := range installations {
		repo := installation.Repository.Name
		repoMap[repo] = installation
		names[i] = repo
	}

	sortedNames, err := TopSortNames(names)
	if err != nil {
		return nil, err
	}

	sorted := make([]*api.Installation, len(installations))
	for i, name := range sortedNames {
		sorted[i] = repoMap[name]
	}
	return sorted, nil
}

func TopSortNames(repos []string) ([]string, error) {
	seen := make(map[string]bool)
	graph := toposort.NewGraph(len(repos))
	for _, repo := range repos {
		if _, ok := seen[repo]; !ok {
			graph.AddNode(repo)
		}
		seen[repo] = true

		man, err := manifest.Read(manifestPath(repo))
		if err != nil {
			return nil, err
		}

		for _, dep := range man.Dependencies {
			if _, ok := seen[dep.Repo]; !ok {
				graph.AddNode(dep.Repo)
				seen[dep.Repo] = true
			}
			graph.AddEdge(repo, dep.Repo)
		}
	}

	sorted, ok := graph.Toposort()
	if !ok {
		return nil, fmt.Errorf("Cycle detected in dependency graph")
	}

	// need to reverse the order
	result := make([]string, len(sorted))
	for i := 1; i <= len(result); i++ {
		result[len(result) - i] = sorted[i - 1]
	}

	return result, nil
}

func Dependencies(repo string, installations []*api.Installation) ([]*api.Installation, error) {
	topsorted, err := TopSort(installations)
	if err != nil {
		return topsorted, err
	}

	ind := 0
	for i := 0; i < len(topsorted); i++ {
		ind = i
		if topsorted[i].Repository.Name == repo {
			break
		}
	}

	return topsorted[:(ind + 1)], err
}
