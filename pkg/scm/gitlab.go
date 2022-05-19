package scm

import (
	"fmt"
	"github.com/cli/oauth"
	"github.com/xanzy/go-gitlab"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/utils"
)

var (
	GitlabClientSecret string
)

type Gitlab struct {
	Client *gitlab.Client
}

func (gl *Gitlab) Init() error {
	flow := &oauth.Flow{
		Host: &oauth.Host{
			AuthorizeURL: "https://gitlab.com/oauth/authorize",
			TokenURL: "https://gitlab.com/oauth/token",
		},
		ClientID: "96dc439ce4bfab647a07b96878210015ab83f173b7f5162218954a95b8c10ebe",
		ClientSecret: GitlabClientSecret,
		CallbackURI: "http://127.0.0.1:1337/callback",
		Scopes:   []string{"read_api", "write_repository", "openid", "profile", "email"},
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
		MinAccessLevel: gitlab.AccessLevel(gitlab.DeveloperPermissions),
	})
	if err != nil {
		return
	}

	orgNames := make([]string, len(groups))
	for i, g := range groups {
		orgNames[i] = g.Path
	}
	orgNames = append(orgNames, user.Username)

	org := ""
	prompt := &survey.Select{
		Message: "Select the group or path for your repo:",
		Options: orgNames,
	}
	survey.AskOne(prompt, &org, survey.WithValidator(survey.Required))

	pub, priv, err := generateKeys()
	if err != nil {
		return
	}
	
	owner := org
	if owner == user.Username {
		owner = ""
	}

	repoName := repoName()
	utils.Highlight("\ncreating gitlab repository %s/%s...\n", org, repoName)
	repo, _, err := gl.Client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Path:                 gitlab.String(fmt.Sprintf("%s/%s", owner, repoName)),
		Visibility:           gitlab.Visibility(gitlab.PrivateVisibility),
		InitializeWithReadme: gitlab.Bool(true),
	})
	if err != nil {
		return
	}

	utils.Highlight("Setting up a read-write deploy key for this repo...\n")
	_, _, err = gl.Client.DeployKeys.AddDeployKey(repo.ID, &gitlab.AddDeployKeyOptions{
		Title:   gitlab.String("Plural Key"),
		Key:     gitlab.String(pub),
		CanPush: gitlab.Bool(true),
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