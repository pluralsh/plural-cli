package ai

import (
	"fmt"
	"time"

	"github.com/pluralsh/plural-cli/pkg/client"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

const intro = "What can we do to help you with Plural, using open source, or kubernetes?"

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
	chat := []*api.ChatMessage{{Role: "system", Content: intro}}
	utils.Success("Plural AI:\n")
	fmt.Printf("%s\n\n", intro)

	for {
		prompt, err := utils.ReadLine(color.New(color.FgYellow).Sprintf("You:\n"))
		if err != nil {
			return err
		}
		chat = append(chat, &api.ChatMessage{Role: "user", Content: prompt})
		fmt.Print("\n")

		utils.Success("Plural AI:\n")
		s := spinner.New(spinner.CharSets[32], 100*time.Millisecond)
		s.Prefix = "Thinking "
		s.Start()

		msg, err := p.Client.Chat(chat)
		if err != nil {
			return err
		}
		s.Stop()

		fmt.Printf("%s\n\n", msg.Content)
		chat = append(chat, msg)
	}
}
