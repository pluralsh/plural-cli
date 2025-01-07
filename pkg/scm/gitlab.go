package scm

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/oauth"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"gitlab.com/gitlab-org/api/client-go"
)

var (
	GitlabClientSecret string
)

type Gitlab struct {
	Client *gitlab.Client
}

func (gl *Gitlab) StarPluralGitHubRep() error {
	return nil
}

func (gl *Gitlab) Init() error {
	flow := &oauth.Flow{
		Host: &oauth.Host{
			AuthorizeURL: "https://gitlab.com/oauth/authorize",
			TokenURL:     "https://gitlab.com/oauth/token",
		},
		ClientID:     "96dc439ce4bfab647a07b96878210015ab83f173b7f5162218954a95b8c10ebe",
		ClientSecret: GitlabClientSecret,
		CallbackURI:  "http://127.0.0.1:1337/callback",
		Scopes:       []string{"api", "openid", "profile", "email"},
		ResponseType: "code",
	}

	accessToken, err := flow.WebAppFlow()
	if err != nil {
		return err
	}

	git, err := gitlab.NewOAuthClient(accessToken.Token)
	gl.Client = git
	return err
}

func (gl *Gitlab) Setup() (con Context, err error) {
	user, _, err := gl.Client.Users.CurrentUser()
	if err != nil {
		return
	}

	emails, _, err := gl.Client.Users.ListEmails()
	if err != nil {
		return
	}

	if len(emails) > 0 {
		con.email = emails[0].Email
	}

	groups, _, err := gl.Client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		MinAccessLevel: gitlab.Ptr(gitlab.DeveloperPermissions),
	})
	if err != nil {
		return
	}

	orgNames := make([]string, len(groups))
	namespaces := make(map[string]int)
	for i, g := range groups {
		orgNames[i] = g.Path
		namespaces[g.Path] = g.ID
	}
	orgNames = append(orgNames, user.Username)

	org := ""
	prompt := &survey.Select{
		Message: "Select the group or path for your repo:",
		Options: orgNames,
	}
	if err := survey.AskOne(prompt, &org, survey.WithValidator(survey.Required)); err != nil {
		return Context{}, err
	}

	pub, priv, err := GenerateKeys(false)
	if err != nil {
		return
	}

	repoName, err := repoName()
	if err != nil {
		return
	}
	opts := &gitlab.CreateProjectOptions{
		Name:                 gitlab.Ptr(repoName),
		Visibility:           gitlab.Ptr(gitlab.PrivateVisibility),
		Description:          gitlab.Ptr("my plural installation repository"),
		InitializeWithReadme: gitlab.Ptr(true),
	}

	if org != user.Username {
		opts.NamespaceID = gitlab.Ptr(namespaces[org])
	}

	utils.Highlight("\ncreating gitlab repository %s/%s...\n", org, repoName)
	repo, _, err := gl.Client.Projects.CreateProject(opts)
	if err != nil {
		return
	}

	utils.Highlight("Setting up a read-write deploy key for this repo...\n")
	_, _, err = gl.Client.DeployKeys.AddDeployKey(repo.ID, &gitlab.AddDeployKeyOptions{
		Title:   gitlab.Ptr("Plural Deploy Key"),
		Key:     gitlab.Ptr(pub),
		CanPush: gitlab.Ptr(true),
	})
	if err != nil {
		return
	}

	con.pub = pub
	con.priv = priv
	con.username = user.Username
	con.url = repo.SSHURLToRepo
	con.repoName = repoName
	return
}
