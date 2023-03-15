package diff

import (
	"path/filepath"

	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func defaultDiff(path string) []*executor.Step {
	return []*executor.Step{
		{
			Name:    "terraform-init",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Command: "terraform",
			Args:    []string{"init", "-upgrade"},
			Sha:     "",
		},
		{
			Name:    "terraform",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Command: "plural",
			Args:    []string{"wkspace", "terraform-diff", path},
			Sha:     "",
		},
		{
			Name:    "helm",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "helm")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "helm")),
			Command: "plural",
			Args:    []string{"wkspace", "helm-diff", path},
			Sha:     "",
		},
	}
}
