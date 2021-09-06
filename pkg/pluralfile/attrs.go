package pluralfile

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/api"
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
	contents, err := ioutil.ReadFile(fullPath)
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
	client := api.NewUploadClient()
	return newsha, client.CreateRepository(repo, a.Publisher, input)
}

func (a *RepoAttrs) mkSha(fullPath string, input *api.RepositoryInput) (sha string, err error) {
	base, err := utils.Sha256(fullPath)
	if err != nil {
		return
	}

	iconSha, err := fileSha(input.Icon)
	if err != nil {
		return
	}
	
	darkIconSha, err := fileSha(input.DarkIcon)
	if err != nil {
		return
	}

	sha = utils.Sha([]byte(fmt.Sprintf("%s:%s:%s", base, iconSha, darkIconSha)))
	return
}

func fileSha(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	fpath, _ := filepath.Abs(path)
	return utils.Sha256(fpath) 
}
