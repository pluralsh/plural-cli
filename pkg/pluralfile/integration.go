package pluralfile

import (
	"github.com/pluralsh/plural-cli/pkg/executor"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

type Integration struct {
	File string
}

func (a *Integration) Type() ComponentName {
	return INTEGRATION
}

func (a *Integration) Key() string {
	return a.File
}

func (a *Integration) Push(repo string, sha string) (string, error) {
	newsha, err := executor.MkHash(a.File, []string{})
	if err != nil || newsha == sha {
		utils.Highlight("No change for %s\n", a.File)
		return sha, err
	}

	utils.Highlight("pushing integration %s", a.File)
	cmd, output := executor.SuppressedCommand("plural", "push", "integration", a.File, repo)

	err = executor.RunCommand(cmd, output)
	return newsha, err
}
