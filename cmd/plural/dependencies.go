package main

import (
	"fmt"

	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/urfave/cli"
)

func (p *Plural) topsort(c *cli.Context) error {
	p.InitPluralClient()
	installations, _ := p.GetInstallations()
	repoName := c.Args().Get(0)
	sorted, err := wkspace.Dependencies(p.Client, repoName, installations)
	if err != nil {
		return err
	}

	for _, inst := range sorted {
		fmt.Println(inst.Repository.Name)
	}
	return nil
}
