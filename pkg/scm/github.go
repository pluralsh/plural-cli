package scm

import (
	"context"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/go-github/v44/github"
	"github.com/pluralsh/oauth"
	"golang.org/x/oauth2"

	"github.com/pluralsh/plural/pkg/utils"
)

type Github struct {
	Client *github.Client
}

func (gh *Github) Init() error {
	flow := &oauth.Flow{
		Host:     oauth.GitHubHost("https://github.com"),
		ClientID: "049b15d353583be950b7",
		Scopes:   []string{"user", "user:email", "repo", "read:org"},
	}
	accessToken, err := flow.DetectFlow()
	if err != nil {
		return err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	gh.Client = github.NewClient(tc)
	return nil
}

func (gh *Github) Setup() (con Context, err error) {
	ctx := context.Background()
	user, _, err := gh.Client.Users.Get(ctx, "")
	if err != nil {
		return
	}

	emails, _, err := gh.Client.Users.ListEmails(ctx, &github.ListOptions{PerPage: 10})
	if err != nil {
		return
	}

	for _, email := range emails {
		if *email.Primary {
			con.email = *email.Email
		}
	}

	orgs, _, err := gh.Client.Organizations.List(ctx, "", &github.ListOptions{PerPage: 10})
	if err != nil {
		return
	}

	orgNames := make([]string, len(orgs))
	for i, o := range orgs {
		orgNames[i] = *o.Login
	}
	orgNames = append(orgNames, *user.Login)

	org := ""
	prompt := &survey.Select{
		Message: "Select the org for your repo:",
		Options: orgNames,
	}
	if err := survey.AskOne(prompt, &org, survey.WithValidator(survey.Required)); err != nil {
		return Context{}, err
	}

	pub, priv, err := GenerateKeys(false)
	if err != nil {
		return
	}

	owner := org
	if owner == *user.Login {
		owner = ""
	}
	repoName, err := repoName()
	if err != nil {
		return
	}
	utils.Highlight("\ncreating github repository %s/%s...\n", org, repoName)
	repo, _, err := gh.Client.Repositories.Create(ctx, owner, &github.Repository{
		Name:     github.String(repoName),
		Private:  github.Bool(true),
		AutoInit: github.Bool(true),
	})
	if err != nil {
		return
	}

	utils.Highlight("Setting up a read-write deploy key for this repo...\n")
	_, _, err = gh.Client.Repositories.CreateKey(ctx, org, *repo.Name, &github.Key{
		Key:      github.String(pub),
		Title:    github.String("plural key"),
		ReadOnly: github.Bool(false),
	})
	if err != nil {
		return
	}

	con.pub = pub
	con.priv = priv
	con.username = *user.Login
	con.url = *repo.SSHURL
	con.repoName = repoName
	return
}
