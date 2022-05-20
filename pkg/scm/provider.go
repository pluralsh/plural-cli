package scm

import (
	"os"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils"
)

var providers = []string{"github", "gitlab"}

type Provider interface {
	Init() error
	Setup() (Context, error)
}

func Setup() (string, error) {
	provider := ""
	prompt := &survey.Select{
		Message: "Select the SCM provider to use for your repository:",
		Options: providers,
	}
	survey.AskOne(prompt, &provider, survey.WithValidator(survey.Required))

	var prov Provider
	switch provider {
	case "github":
		prov = &Github{}
	case "gitlab":
		prov = &Gitlab{}
	default:
		return "", nil
	}

	if err := prov.Init(); err != nil {
		return "", err
	}

	ctx, err := prov.Setup()
	if err != nil {
		return "", err
	}

	utils.Highlight("Cloning the repo locally (be sure you have git ssh auth set up)\n")
	auth, _ := git.SSHAuth("git", ctx.priv, "")
	if _, err := git.Clone(auth, ctx.url, ctx.repoName); err != nil {
		return "", err
	}

	os.Chdir(ctx.repoName)
	if err := buildContext(&ctx); err != nil {
		return "", err
	}

	fmt.Println("")
	return ctx.repoName, nil
}