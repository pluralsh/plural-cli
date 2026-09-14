package up

import (
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

// promptCommitMessage mirrors common.CommitMsg's interactive survey (no cli.Context).
func promptCommitMessage() string {
	utils.Highlight("\n==> Enter a commit message to push your configuration\n\n")
	var commit string
	if err := survey.AskOne(&survey.Input{
		Message: "Enter a commit message (empty to not commit right now)",
	}, &commit); err != nil {
		return ""
	}
	return strings.TrimSpace(commit)
}
