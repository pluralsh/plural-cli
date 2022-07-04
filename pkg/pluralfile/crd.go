package pluralfile

import (
	"fmt"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
)

type Crd struct {
	File  string
	Chart string
}

func (a *Crd) Type() ComponentName {
	return CRD
}

func (a *Crd) Key() string {
	return a.File
}

func (c *Crd) Push(repo string, sha string) (string, error) {
	crdSha, err := executor.MkHash(c.File, []string{})
	if err != nil {
		return sha, err
	}

	chartSha, err := executor.MkHash(c.Chart, []string{})
	if err != nil {
		return sha, err
	}

	newsha := fmt.Sprintf("%s:%s", crdSha, chartSha)
	if newsha == sha {
		utils.Highlight("No change for %s\n", c.File)
		return sha, nil
	}

	chart := filepath.Base(c.Chart)
	utils.Highlight("pushing crd %s for %s\n", c.File, chart)
	cmd := exec.Command("plural", "push", "crd", c.File, repo, chart)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return newsha, cmd.Run()
}
