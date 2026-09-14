package edge

import (
	"os"

	pkgedge "github.com/pluralsh/plural-cli/pkg/edge"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli"
)

func (p *Plural) handleEdgeFlash(c *cli.Context) error {
	image := c.String("image")
	device := c.String("device")

	stat, err := os.Stat(image)
	if err != nil {
		return err
	}
	bar := progressbar.DefaultBytes(stat.Size(), "flashing")
	if err := pkgedge.Flash(pkgedge.FlashOptions{Image: image, Device: device, Progress: bar}); err != nil {
		return err
	}

	utils.Success("image flashed on %s device\n", device)
	return nil
}
