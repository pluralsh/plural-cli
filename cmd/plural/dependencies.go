package main

import (
	"fmt"

	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/urfave/cli/v2"
)

func (p *Plural) topsort(c *cli.Context) error {
	p.InitPluralClient()
	installations, _ := p.GetInstallations()
	repoName := c.Args().Get(0)
	sorted, err := wkspace.UntilRepo(p.Client, repoName, installations)
	if err != nil {
		return err
	}

	for _, inst := range sorted {
		fmt.Println(inst.Repository.Name)
	}
	return nil
}

func (p *Plural) dependencies(c *cli.Context) error {
	repo := c.Args().Get(0)
	deps, err := wkspace.Dependencies(repo)
	if err != nil {
		return err
	}

	for _, dep := range deps {
		fmt.Println(dep)
	}

	return nil
}
