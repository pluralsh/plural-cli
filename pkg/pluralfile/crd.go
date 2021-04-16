package pluralfile

import (
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/utils"
	"os"
	"os/exec"
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

func (a *Crd) Push(repo string, sha string) (string, error) {
	newsha, _ := executor.MkHash(a.File, []string{})
	// if err != nil || newsha == sha {
	// 	utils.Highlight("No change for %s\n", a.File)
	// 	return sha, nil
	// }

	utils.Highlight("pushing crd %s for %s\n", a.File, a.Chart)
	cmd := exec.Command("plural", "push", "crd", a.File, repo, a.Chart)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return newsha, cmd.Run()
}
