package pluralfile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
)

type Artifact struct {
	File     string
	Platform string
	Arch     string
}

func (a *Artifact) Type() ComponentName {
	return ARTIFACT
}

func (a *Artifact) Key() string {
	return fmt.Sprintf("%s_%s_%s", a.File, a.Platform, a.Arch)
}

func (a *Artifact) Push(repo string, sha string) (string, error) {
	newsha, err := mkSha(a.File)
	if err != nil || newsha == sha {
		utils.Highlight("No change for %s\n", a.File)
		return sha, err
	}

	utils.Highlight("pushing artifact %s\n [plat=%s,arch=%s]", a.File, a.Platform, a.Arch)
	cmd := exec.Command("plural", "push", "artifact", a.File, repo, "--platform", a.Platform, "--arch", a.Arch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return newsha, err
}

func mkSha(file string) (sha string, err error) {
	fullPath, _ := filepath.Abs(file)
	base, err := utils.Sha256(fullPath)
	if err != nil {
		return
	}

	contents, err := os.ReadFile(fullPath)
	if err != nil {
		return
	}

	input, err := api.ConstructArtifactAttributes(contents)
	if err != nil {
		return
	}

	readme, err := fileSha(input.Readme)
	if err != nil {
		return
	}

	blob, err := fileSha(input.Blob)
	if err != nil {
		return
	}

	sha = utils.Sha([]byte(fmt.Sprintf("%s:%s:%s", base, readme, blob)))
	return
}
