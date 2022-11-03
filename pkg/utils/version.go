package utils

import (
	"context"

	semverlib "github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v45/github"
)

func CheckLatestVersion(currentVersion string) {
	client := github.NewClient(nil)
	latestRelease, _, err := client.Repositories.GetLatestRelease(context.Background(), "pluralsh", "plural-cli")
	if err == nil {
		latestVersion := *latestRelease.TagName
		cv, err := semverlib.NewVersion(currentVersion)
		if err != nil {
			return
		}
		lv, err := semverlib.NewVersion(latestVersion)
		if err != nil {
			return
		}
		if cv.Major() == lv.Major() && cv.Minor() == lv.Minor() && cv.Patch() == lv.Patch() {
			return
		}
		Warn("** There is a new version of the Plural CLI, please upgrade it with your package manager **\n")
	}
}
