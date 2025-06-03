package pluralfile

import (
	"github.com/pluralsh/plural-cli/pkg/executor"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

type ResourceDefinition struct {
	File string
}

func (a *ResourceDefinition) Type() ComponentName {
	return IRD
}

func (a *ResourceDefinition) Key() string {
	return a.File
}

func (a *ResourceDefinition) Push(repo string, sha string) (string, error) {
	newsha, err := executor.MkHash(a.File, []string{})
	if err != nil || newsha == sha {
		utils.Highlight("No change for %s\n", a.File)
		return sha, err
	}

	utils.Highlight("pushing integration definition %s", a.File)
	cmd, output := executor.SuppressedCommand("plural", "push", "resourcedefinition", a.File, repo)

	err = executor.RunCommand(cmd, output)
	return newsha, err
}
