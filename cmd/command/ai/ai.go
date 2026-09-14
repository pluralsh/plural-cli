package ai

import (
	"fmt"
	"time"

	"github.com/pluralsh/plural-cli/pkg/client"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/pluralsh/plural-cli/pkg/api"
	aibridge "github.com/pluralsh/plural-cli/pkg/bridge/ai"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:     "ai",
		Usage:    "utilize openai to get help with your setup",
		Action:   p.aiHelp,
		Category: "Debugging",
	}
}

func (p *Plural) aiHelp(c *cli.Context) error {
	p.InitPluralClient()
	chat := []*api.ChatMessage{{Role: aibridge.RoleSystem, Content: aibridge.Intro}}
	utils.Success("Plural AI:\n")
	fmt.Printf("%s\n\n", aibridge.Intro)

	for {
		prompt, err := utils.ReadLine(color.New(color.FgYellow).Sprintf("You:\n"))
		if err != nil {
			return err
		}
		chat = append(chat, &api.ChatMessage{Role: aibridge.RoleUser, Content: prompt})
		fmt.Print("\n")

		utils.Success("Plural AI:\n")
		s := spinner.New(spinner.CharSets[32], 100*time.Millisecond)
		s.Prefix = "Thinking "
		s.Start()

		msg, err := p.Chat(chat)
		if err != nil {
			return err
		}
		s.Stop()

		fmt.Printf("%s\n\n", msg.Content)
		chat = append(chat, msg)
	}
}
