package up

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

// DomainNoneOption matches cmd/command/up noneOption for skipping app domain.
const DomainNoneOption = "None"
