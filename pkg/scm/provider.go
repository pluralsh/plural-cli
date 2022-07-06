package scm

import (
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
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
	if err := survey.AskOne(prompt, &provider, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

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

	time.Sleep(3 * time.Second)
	utils.Highlight("Cloning the repo locally (be sure you have git ssh auth set up, you can use `plural crypto ssh-keygen` to create your first ssh keys then upload the public key to your git provider)\n")
	auth, _ := git.SSHAuth("git", ctx.priv, "")
	if _, err := git.Clone(auth, ctx.url, ctx.repoName); err != nil {
		return "", err
	}

	if err := os.Chdir(ctx.repoName); err != nil {
		return "", err
	}
	if err := buildContext(&ctx); err != nil {
		return "", err
	}

	fmt.Println("")
	return ctx.repoName, nil
}
