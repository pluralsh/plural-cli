package agents

import (
	"fmt"

	console "github.com/pluralsh/console/go/client"
)

// OverwritePrompt decides whether an existing local provider session file can be
// replaced during restore.
type OverwritePrompt func(path string) (bool, error)

func overwriteExisting(_ string) (bool, error) {
	return true, nil
}

func (opts RestoreOptions) sessionOverwritePrompt(provider console.AgentRuntimeType) OverwritePrompt {
	if opts.ConfirmOverwrite != nil {
		return opts.ConfirmOverwrite
	}
	var decided bool
	var overwrite bool
	return func(_ string) (bool, error) {
		if decided {
			return overwrite, nil
		}
		session := "session"
		if opts.Manifest.Session.ID != "" {
			session = fmt.Sprintf("session %s", opts.Manifest.Session.ID)
		}
		interaction := opts.Interaction
		if interaction == nil {
			interaction = NewSurveyInteraction()
		}
		var err error
		overwrite, err = interaction.Confirm(fmt.Sprintf("%s %s already exists. Overwrite?", provider, session), false)
		decided = true
		return overwrite, err
	}
}
