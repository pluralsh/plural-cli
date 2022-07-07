package pluralfile

import (
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/utils"
)

type Helm struct {
	File string
}

func (a *Helm) Type() ComponentName {
	return HELM
}

func (a *Helm) Key() string {
	return a.File
}

func (a *Helm) Push(repo string, sha string) (string, error) {
	newsha, err := executor.MkHash(a.File, []string{})
	if err != nil || newsha == sha {
		utils.Highlight("No change for %s\n", a.File)
		return sha, nil
	}

	utils.Highlight("pushing helm %s", a.File)
	cmd, output := executor.SuppressedCommand("plural", "push", "helm", a.File, repo)

	err = executor.RunCommand(cmd, output)
	return newsha, err
}
