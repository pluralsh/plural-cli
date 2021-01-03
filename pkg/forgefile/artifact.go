package forgefile

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/michaeljguarino/forge/pkg/api"
	"github.com/michaeljguarino/forge/pkg/utils"
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
	cmd := exec.Command("forge", "push", "artifact", a.File, repo, "--platform", a.Platform, "--arch", a.Arch)
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

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return
	}

	input, err := api.ConstructArtifactAttributes(contents)
	if err != nil {
		return
	}

	readmePath, _ := filepath.Abs(input.Readme)
	readme, err := utils.Sha256(readmePath)
	if err != nil {
		return
	}

	blobPath, _ := filepath.Abs(input.Blob)
	blob, err := utils.Sha256(blobPath)
	if err != nil {
		return
	}

	sha = utils.Sha([]byte(fmt.Sprintf("%s:%s:%s", base, readme, blob)))
	return
}
