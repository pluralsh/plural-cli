package wkspace

import (
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/polly/algorithms"
	"github.com/pluralsh/polly/containers"
)

type depsFetcher func(string) ([]*manifest.Dependency, error)

func SortAndFilter(installations []*api.Installation) ([]string, error) {
	names := algorithms.Map(installations, func(i *api.Installation) string { return i.Repository.Name })
	names = algorithms.Filter(names, isRepo)
	return TopSortNames(names)
}

func TopSort(client api.Client, installations []*api.Installation) ([]*api.Installation, error) {
	var repoMap = make(map[string]*api.Installation)
	g := containers.NewGraph[string]()
	for _, installation := range installations {
		repo := installation.Repository.Name
		repoMap[repo] = installation
		ci, tf, err := client.GetPackageInstallations(installation.Repository.Id)
		if err != nil {
			return nil, err
		}

		deps := algorithms.Map(buildDependencies(repo, ci, tf), func(d *manifest.Dependency) string { return d.Repo })
		for _, d := range algorithms.Filter(deps, isRepo) {
			g.AddEdge(d, repo)
		}
	}

	names, err := algorithms.TopsortGraph(g)
	if err != nil {
		return nil, err
	}

	return algorithms.Map(names, func(r string) *api.Installation { return repoMap[r] }), nil
}

func TopSortNames(repos []string) ([]string, error) {
	g := containers.NewGraph[string]()
	for _, r := range repos {
		deps, err := findDependencies(r)
		if err != nil {
			return nil, err
		}

		for _, dep := range deps {
			g.AddEdge(dep, r)
		}
	}

	return algorithms.TopsortGraph(g)
}

func findDependencies(repo string) ([]string, error) {
	man, err := manifest.Read(manifestPath(repo))
	if err != nil {
		return nil, err
	}

	deps := algorithms.Map(man.Dependencies, func(d *manifest.Dependency) string { return d.Repo })
	return algorithms.Filter(deps, func(r string) bool { return isRepo(r) }), nil
}

func Dependencies(repo string) ([]string, error) {
	// dfs from the repo to find all dependencies
	deps := []string{}
	visit := func(r string) error {
		deps = append(deps, r)
		return nil
	}
	if err := algorithms.DFS(repo, findDependencies, visit); err != nil {
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
