package plural

import (
	"io"
	"os"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/urfave/cli"
)

func (p *Plural) handleUpgrade(c *cli.Context) (err error) {
	p.InitPluralClient()
	queue, repo := c.Args().Get(0), c.Args().Get(1)
	f := os.Stdin
	fname := c.String("f")
	if fname != "-" && fname != "" {
		f, err = os.Open(fname)
		if err != nil {
			return
		}
		defer f.Close()
	}

	contents, err := io.ReadAll(f)
	if err != nil {
		return
	}

	attrs, err := api.ConstructUpgradeAttributes(contents)
	if err != nil {
		return
	}

	err = p.CreateUpgrade(queue, repo, attrs)
	return api.GetErrorResponse(err, "CreateUpgrade")
}
