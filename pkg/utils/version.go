package utils

import (
	"context"
	"sort"

	semverlib "github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v45/github"
)

type Versions []*semverlib.Version

func (s Versions) Len() int {
	return len(s)
}

func (s Versions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Versions) Less(i, j int) bool {
	return s[i].LessThan(s[j])
}

func CheckLatestVersion(currentVersion string) {
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(context.Background(), "pluralsh", "plural-cli", &github.ListOptions{
		Page:    0,
		PerPage: 10,
	})
	if err != nil {
		return
	}
	var versions Versions
	for _, r := range releases {
		v, err := semverlib.NewVersion(*r.TagName)
		if err != nil {
			return
		}
		versions = append(versions, v)
	}
	sort.Sort(sort.Reverse(versions))
	if len(versions) > 0 {
		lv := versions[0]
		cv, err := semverlib.NewVersion(currentVersion)
		if err != nil {
			return
		}

		if cv.Major() == lv.Major() && cv.Minor() == lv.Minor() && cv.Patch() >= lv.Patch() {
			return
		}

		Warn("** There is a new version of the Plural CLI, please upgrade it with your package manager **\n")
	}
}
