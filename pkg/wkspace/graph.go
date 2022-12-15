package wkspace

import (
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/polly/algorithms"
	"github.com/pluralsh/polly/containers"
	"github.com/samber/lo"
)

func SortAndFilter(installations []*api.Installation) ([]string, error) {
	names := lo.FilterMap(installations, func(i *api.Installation, ind int) (string, bool) {
		name := i.Repository.Name
		return name, isRepo(name)
	})
	return TopSortNames(names)
}

func TopSort(client api.Client, installations []*api.Installation) ([]*api.Installation, error) {
	repoMap := map[string]*api.Installation{}
	g := containers.NewGraph[string]()
	for _, installation := range installations {
		repo := installation.Repository.Name
		repoMap[repo] = installation
		g.AddNode(repo)

		ci, tf, err := client.GetPackageInstallations(installation.Repository.Id)
		if err != nil {
			return nil, api.GetErrorResponse(err, "GetPackageInstallations")
		}
		deps := algorithms.Map(buildDependencies(repo, ci, tf), func(d *manifest.Dependency) string { return d.Repo })
		for _, d := range deps {
			g.AddEdge(d, repo)
		}
	}

	names, err := algorithms.TopsortGraph(g)
	if err != nil {
		return nil, err
	}

	insts := lo.FilterMap(names, func(r string, ind int) (i *api.Installation, ok bool) {
		i, ok = repoMap[r]
		return
	})
	return insts, nil
}

func TopSortNames(repos []string) ([]string, error) {
	g := containers.NewGraph[string]()
	for _, r := range repos {
		g.AddNode(r)
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

	return lo.FilterMap(man.Dependencies, func(d *manifest.Dependency, ind int) (string, bool) { return d.Repo, isRepo(d.Repo) }), nil
}

func AllDependencies(repos []string) ([]string, error) {
	deps := []string{}
	visit := func(r string) error {
		deps = append(deps, r)
		return nil
	}
	for _, repo := range repos {
		if err := algorithms.DFS(repo, findDependencies, visit); err != nil {
			return deps, err
		}
	}

	return TopSortNames(lo.Uniq(deps))
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
	if err != nil || repo == "" {
		return topsorted, err
	}

	ind := algorithms.Index(topsorted, func(i *api.Installation) bool { return i.Repository.Name == repo })
	return topsorted[:(ind + 1)], err
}
