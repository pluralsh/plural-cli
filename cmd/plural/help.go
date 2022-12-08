package main

import (
	"fmt"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
)

func (p *Plural) aiHelp(c *cli.Context) error {
	p.InitPluralClient()
	prompt, _ := utils.ReadLine("What do you need help with?\n")
	res, err := p.Client.GetHelp(prompt)
	fmt.Println("")
	fmt.Println(res)
	return api.GetErrorResponse(err, "GetHelp")
}
