package main

import (
	"fmt"
	"os"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli"
)

func requireArgs(fn func(*cli.Context) error, args []string) func(*cli.Context) error {
	return func(c *cli.Context) error {
		nargs := c.NArg()
		if nargs > len(args) {
			return fmt.Errorf("Too many args passed to %s.  Try running --help to see usage", c.Command.FullName())
		}

		if nargs < len(args) {
			return fmt.Errorf("Not enough arguments provided, needs %s, try running --help to see usage", args[nargs])
		}

		return fn(c)
	}
}

func rooted(fn func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := repoRoot(); err != nil {
			return err
		}

		return fn(c)
	}
}

func owned(fn func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := validateOwner(); err != nil {
			return err
		}

		return fn(c)
	}
}

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

func repoRoot() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	root, err := git.Root()
	if err != nil {
		return err
	}

	if root != dir {
		return fmt.Errorf("You must run this command at the root of your git repository")
	}

	return nil
}