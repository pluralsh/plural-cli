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
	imagePath := c.String("image")
	devicePath := c.String("device")

	if err := p.flashImage(imagePath, devicePath); err != nil {
		return err
	}

	utils.Success("image flashed on %s device\n", devicePath)
	return nil
}

func (p *Plural) flashImage(imagePath, devicePath string) error {
	out, err := os.OpenFile(devicePath, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open device: %v", err)
	}
	defer out.Close()

	in, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("could not open image: %v", err)
	}
	defer in.Close()

	stat, err := os.Stat(imagePath)
	if err != nil {
		return err
	}

	bar := progressbar.DefaultBytes(stat.Size(), "flashing")
	_, err = io.Copy(io.MultiWriter(out, bar), in)
	return err
}
