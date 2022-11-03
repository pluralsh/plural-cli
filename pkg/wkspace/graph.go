package wkspace

import (
	"fmt"

	toposort "github.com/philopon/go-toposort"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils/containers"
)

type depsFetcher func(string) ([]*manifest.Dependency, error)

func SortAndFilter(installations []*api.Installation) ([]string, error) {
	names := make([]string, 0)
	for _, inst := range installations {
		if isRepo(inst.Repository.Name) {
			names = append(names, inst.Repository.Name)
		}
	}

	return TopSortNames(names)
}

func TopSort(client api.Client, installations []*api.Installation) ([]*api.Installation, error) {
	var repoMap = make(map[string]*api.Installation)
	var depsMap = make(map[string][]*manifest.Dependency)
	names := make([]string, len(installations))

	for i, installation := range installations {
		repo := installation.Repository.Name
		repoMap[repo] = installation
		names[i] = repo

		ci, tf, err := client.GetPackageInstallations(installation.Repository.Id)
		if err != nil {
			return nil, err
		}

		depsMap[repo] = buildDependencies(repo, ci, tf)
	}

	sortedNames, err := topsorter(names, func(repo string) ([]*manifest.Dependency, error) {
		if deps, ok := depsMap[repo]; ok {
			return deps, nil
		}

		return nil, fmt.Errorf("unknown repository %s", repo)
	})

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
	return topsorter(repos, findDependencies)
}

func findDependencies(repo string) ([]*manifest.Dependency, error) {
	man, err := manifest.Read(manifestPath(repo))
	if err != nil {
		return nil, err
	}

	return man.Dependencies, nil
}

func topsorter(repos []string, fn depsFetcher) ([]string, error) {
	seen := make(map[string]bool)
	graph := toposort.NewGraph(len(repos))
	isRepo := make(map[string]bool)
	for _, repo := range repos {
		isRepo[repo] = true
	}

	for _, repo := range repos {
		if _, ok := seen[repo]; !ok {
			graph.AddNode(repo)
		}
		seen[repo] = true

		deps, err := fn(repo)
		if err != nil {
			return nil, err
		}

		for _, dep := range deps {
			if _, ok := isRepo[dep.Repo]; !ok {
				continue
			}

			if _, ok := seen[dep.Repo]; !ok {
				graph.AddNode(dep.Repo)
				seen[dep.Repo] = true
			}
			graph.AddEdge(repo, dep.Repo)
		}
	}

	sorted, ok := graph.Toposort()
	if !ok {
		return nil, fmt.Errorf("cycle detected in dependency graph")
	}

	// need to reverse the order
	result := make([]string, len(sorted))
	for i := 1; i <= len(result); i++ {
		result[len(result)-i] = sorted[i-1]
	}

	return result, nil
}

func Dependencies(repo string) ([]string, error) {
	// dfs from the repo to find all dependencies
	deps, err := containers.DFS(repo, func(r string) ([]string, error) {
		res := []string{}
		ds, err := findDependencies(r)
		if err != nil {
			return res, err
		}

		for _, dep := range ds {
			if isRepo(dep.Repo) {
				res = append(res, dep.Repo)
			}
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}

	// topsort only those to find correct ordering
	return TopSortNames(deps)
}

func UntilRepo(client api.Client, repo string, installations []*api.Installation) ([]*api.Installation, error) {
	topsorted, err := TopSort(client, installations)
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
