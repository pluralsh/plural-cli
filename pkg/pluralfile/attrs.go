package pluralfile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/utils"
)

type RepoAttrs struct {
	File      string
	Publisher string
}

func (a *RepoAttrs) Type() ComponentName {
	return REPO_ATTRS
}

func (a *RepoAttrs) Key() string {
	return fmt.Sprintf("%s_%s", a.File, a.Publisher)
}

func (a *RepoAttrs) Push(repo string, sha string) (string, error) {
	fullPath, _ := filepath.Abs(a.File)
	contents, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	input, err := api.ConstructRepositoryInput(contents)
	if err != nil {
		return "", err
	}

	newsha, err := a.mkSha(fullPath, input)
	if err != nil || newsha == sha {
		utils.Highlight("No change for %s\n", a.File)
		return sha, err
	}

	utils.Highlight("Setting attributes for %s\n", repo)
	client := api.NewClient()
	repositoryAttributes, err := api.ConstructGqlClientRepositoryInput(contents)
	if err != nil {
		return "", err
	}
	err = client.CreateRepository(repo, a.Publisher, repositoryAttributes)
	return newsha, api.GetErrorResponse(err, "CreateRepository")
}

func (a *RepoAttrs) mkSha(fullPath string, input *api.RepositoryInput) (sha string, err error) {
	base, err := utils.Sha256(fullPath)
	if err != nil {
		return
	}

	iconSha, _ := fileSha(input.Icon)
	darkIconSha, _ := fileSha(input.DarkIcon)

	notesSha, err := fileSha(input.Notes)
	if err != nil {
		return
	}
	docssha := ""
	if input.Docs != "" {
		fpath, _ := filepath.Abs(input.Docs)
		docssha, err = executor.MkHash(fpath, []string{})
		if err != nil {
			return
		}
	}

	sha = utils.Sha([]byte(fmt.Sprintf("%s:%s:%s:%s%s", base, iconSha, darkIconSha, notesSha, docssha)))
	return
}

func fileSha(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	fpath, _ := filepath.Abs(path)
	return utils.Sha256(fpath)
}
