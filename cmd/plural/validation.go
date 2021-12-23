package main

import (
	"fmt"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/AlecAivazis/survey/v2"
)

func validateOwner() error {
	path := manifest.ProjectManifestPath()
	project, err := manifest.ReadProject(path)
	conf := config.Read()
	if err != nil {
		return fmt.Errorf("Your workspace hasn't been configured, try running `plural init`")
	}

	if owner := project.Owner; owner != nil {
		if owner.Email != conf.Email || owner.Endpoint != conf.Endpoint {
			return fmt.Errorf(
				"The owner of this project is actually %s; plural environemnt = %s",
				owner.Email,
				config.PluralUrl(owner.Endpoint),
			)
		}
	}

	return nil
}

func confirm(msg string) bool {
	res := true
	prompt := &survey.Confirm{Message: msg}
	survey.AskOne(prompt, &res, survey.WithValidator(survey.Required))
	return res
}