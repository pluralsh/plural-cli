package wkspace

import (
	"path/filepath"

	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
)

const (
	readmeUrl = "https://raw.githubusercontent.com/pluralsh/documentation/main/cli-readme/README.md"
)

func DownloadReadme() error {
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	return utils.DownloadFile(filepath.Join(repoRoot, "README.md"), readmeUrl)
}
