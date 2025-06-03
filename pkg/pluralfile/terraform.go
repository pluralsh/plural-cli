package pluralfile

import (
	"os"
	"os/exec"

	"github.com/pluralsh/plural-cli/pkg/executor"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

type Terraform struct {
	File string
}

func (a *Terraform) Type() ComponentName {
	return TERRAFORM
}

func (a *Terraform) Key() string {
	return a.File
}

func (a *Terraform) Push(repo string, sha string) (string, error) {
	newsha, err := executor.MkHash(a.File, []string{})
	if err != nil || newsha == sha {
		if err == nil {
			utils.Highlight("No change for %s\n", a.File)
		}
		return sha, err
	}

	utils.Highlight("pushing terraform %s\n", a.File)
	cmd := exec.Command("plural", "push", "terraform", a.File, repo)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return newsha, err
}
