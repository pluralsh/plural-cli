package up

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/scm"
)

// SCMProvider is one option from scm.Setup's first survey.
type SCMProvider struct {
	ID    string
	Title string
	Blurb string
}

// SCMProviders returns the SCM choices offered by pkg/scm.Setup.
func SCMProviders() []SCMProvider {
	return []SCMProvider{
		{ID: "github", Title: "GitHub", Blurb: "authenticate · create repo · clone"},
		{ID: "gitlab", Title: "GitLab", Blurb: "authenticate · create repo · clone"},
		{ID: "bitbucket", Title: "Bitbucket", Blurb: "authenticate · create repo · clone"},
	}
}

// SetupSCM runs device login + create repo + clone for providerID (github/gitlab/bitbucket).
// Intended to run under tea.Exec so the TUI releases the terminal for oauth/surveys.
func SetupSCM(providerID string) (repoName string, err error) {
	if providerID == "" {
		return "", fmt.Errorf("scm provider is required")
	}
	return scm.SetupProvider(providerID)
}

// DomainNoneOption matches cmd/command/up noneOption for skipping app domain.
const DomainNoneOption = "None"
