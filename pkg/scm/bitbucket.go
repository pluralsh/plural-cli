package scm

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/pluralsh/oauth"
	"github.com/pluralsh/plural/pkg/utils"
)

var (
	BitbucketClientSecret string
)

const emailError = "Can't find the user email address"

type Link struct {
	Name string `json:"name"`
	Href string `json:"href"`
}

type Values struct {
	Email     string `json:"email"`
	IsPrimary bool   `json:"is_primary"`
}

type Bitbucket struct {
	Client *bitbucket.Client
}

func (b *Bitbucket) Init() error {
	flow := &oauth.Flow{
		Host: &oauth.Host{
			AuthorizeURL: "https://bitbucket.org/site/oauth2/authorize",
			TokenURL:     "https://bitbucket.org/site/oauth2/access_token",
		},
		ClientID:     "GVEHgz5FMA24BdcKfA",
		ClientSecret: BitbucketClientSecret,
		CallbackURI:  "http://127.0.0.1/callback",
		ResponseType: "code",
	}

	accessToken, err := flow.WebAppFlow()
	if err != nil {
		return err
	}
	b.Client = bitbucket.NewOAuthbearerToken(accessToken.Token)

	return nil
}

func (b *Bitbucket) Setup() (Context, error) {
	user, err := b.Client.User.Profile()
	if err != nil {
		return Context{}, err
	}
	emails, err := b.Client.User.Emails()
	if err != nil {
		return Context{}, err
	}
	emailAddress := ""
	emailValues, ok := emails.(map[string]interface{})
	if !ok {
		return Context{}, fmt.Errorf(emailError)
	}
	emailAddress, err = getEmailAddress(emailValues)
	if err != nil {
		return Context{}, err
	}

	wl, err := b.Client.Workspaces.List()
	if err != nil {
		return Context{}, err
	}

	workspaces := make([]string, 0)
	for _, w := range wl.Workspaces {
		workspaces = append(workspaces, w.Slug)
	}

	if len(workspaces) == 0 {
		return Context{}, fmt.Errorf("You don't have any Bitbucket workspace created. Please create one first \n")
	}

	workspace := workspaces[0]
	if len(workspaces) > 1 {
		prompt := &survey.Select{
			Message: "Select the workspace:",
			Options: workspaces,
		}
		if err := survey.AskOne(prompt, &workspace, survey.WithValidator(survey.Required)); err != nil {
			return Context{}, err
		}
	}

	pr, err := b.Client.Workspaces.Projects(workspace)
	if err != nil {
		return Context{}, err
	}
	projectKeys := make(map[string]string, 0)
	projects := make([]string, 0)
	for _, p := range pr.Items {
		projectKeys[p.Name] = p.Key
		projects = append(projects, p.Name)
	}

	project := ""
	if len(projects) == 0 {
		return Context{}, fmt.Errorf("You don't have any Bitbucket project created. Please create one first \n")
	}

	prompt := &survey.Select{
		Message: "Select the project for your repo:",
		Options: projects,
	}
	if err := survey.AskOne(prompt, &project, survey.WithValidator(survey.Required)); err != nil {
		return Context{}, err
	}

	pub, priv, err := GenerateKeys(false)
	if err != nil {
		return Context{}, err
	}

	repoName, err := repoName()
	if err != nil {
		return Context{}, err
	}

	opt := &bitbucket.RepositoryOptions{
		Uuid:      "",
		Owner:     user.Username,
		RepoSlug:  repoName,
		Scm:       "git",
		IsPrivate: "true",
		Project:   projectKeys[project],
	}

	utils.Highlight("\ncreating bitbucket repository %s...\n", repoName)

	res, err := b.Client.Repositories.Repository.Create(opt)
	if err != nil {
		return Context{}, err
	}
	sshAddress, err := getSSHAddress(res.Links)
	if err != nil {
		return Context{}, err
	}

	utils.Highlight("Setting up a read-write deploy key for this repo...\n")

	if _, err := b.Client.Repositories.DeployKeys.Create(&bitbucket.DeployKeyOptions{
		Owner:    user.Username,
		RepoSlug: repoName,
		Label:    "Plural Deploy Key",
		Key:      pub,
	}); err != nil {
		return Context{}, err
	}
	// The bitbucket creates empty repository.
	// Initialize bitbucket repo with .gitignore file.
	dir, err := os.MkdirTemp("", "bitbucket")
	if err != nil {
		return Context{}, err
	}
	defer os.RemoveAll(dir)
	gitIgnore := path.Join(dir, ".gitignore")
	if err := os.WriteFile(gitIgnore, []byte(""), 0644); err != nil {
		return Context{}, err
	}

	if err := b.Client.Repositories.Repository.WriteFileBlob(&bitbucket.RepositoryBlobWriteOptions{
		Owner:    user.Username,
		RepoSlug: repoName,
		FilePath: gitIgnore,
		FileName: ".gitignore",
		Author:   fmt.Sprintf("%s <%s>", user.DisplayName, emailAddress),
		Message:  "init",
		Branch:   "master",
	}); err != nil {
		return Context{}, err
	}

	con := Context{}
	con.pub = pub
	con.priv = priv
	con.username = user.Username
	con.url = sshAddress
	con.repoName = repoName
	return con, nil
}

func (b *Bitbucket) StarPluralGitHubRep() error {
	return nil
}

func getSSHAddress(links map[string]interface{}) (string, error) {
	sshAddress := ""
	var linkList []Link
	if links != nil {
		clone := links["clone"]
		jsonStr, err := json.Marshal(clone)
		if err != nil {
			return "", err
		}
		if err := json.Unmarshal(jsonStr, &linkList); err != nil {
			return "", err
		}
	}
	for _, link := range linkList {
		if link.Name == "ssh" {
			sshAddress = link.Href
			break
		}
	}
	if sshAddress == "" {
		return "", fmt.Errorf("Can't find the repository SSH address")
	}
	return sshAddress, nil
}

func getEmailAddress(values map[string]interface{}) (string, error) {
	emailAddress := ""
	var valuesList []Values
	if values != nil {
		clone := values["values"]
		jsonStr, err := json.Marshal(clone)
		if err != nil {
			return "", err
		}
		if err := json.Unmarshal(jsonStr, &valuesList); err != nil {
			return "", err
		}
	}
	for _, value := range valuesList {
		if value.IsPrimary {
			emailAddress = value.Email
			break
		}
	}
	if emailAddress == "" {
		return "", fmt.Errorf(emailError)
	}
	return emailAddress, nil
}
