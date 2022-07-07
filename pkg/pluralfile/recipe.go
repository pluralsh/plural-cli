package pluralfile

import (
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/utils"
)

type Recipe struct {
	File string
}

func (a *Recipe) Type() ComponentName {
	return RECIPE
}

func (a *Recipe) Key() string {
	return a.File
}

func (a *Recipe) Push(repo string, sha string) (string, error) {
	newsha, err := executor.MkHash(a.File, []string{})
	if err != nil || newsha == sha {
		utils.Highlight("No change for %s\n", a.File)
		return sha, err
	}

	utils.Highlight("pushing recipe %s", a.File)
	cmd, output := executor.SuppressedCommand("plural", "push", "recipe", a.File, repo)

	err = executor.RunCommand(cmd, output)
	return newsha, err
}
