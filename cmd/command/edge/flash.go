package edge

import (
	"fmt"
	"io"
	"os"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli"
)

func (p *Plural) handleEdgeFlash(c *cli.Context) error {
	image := c.String("image")
	device := c.String("device")

	if err := p.flashImage(image, device); err != nil {
		return err
	}

	utils.Success("image flashed on %s device\n", device)
	return nil
}

func (p *Plural) flashImage(image, device string) error {
	out, err := os.OpenFile(device, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open device: %v", err)
	}
	defer out.Close()

	in, err := os.Open(image)
	if err != nil {
		return fmt.Errorf("could not open image: %v", err)
	}
	defer in.Close()

	stat, err := os.Stat(image)
	if err != nil {
		return err
	}

	bar := progressbar.DefaultBytes(stat.Size(), "flashing")
	_, err = io.Copy(io.MultiWriter(out, bar), in)
	return err
}
