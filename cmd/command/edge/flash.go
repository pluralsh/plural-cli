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

	in, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("could not open image: %v", err)
	}
	defer in.Close()

	stat, err := os.Stat(imagePath)
	if err != nil {
		return err
	}

	out, err := os.OpenFile(devicePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("could not open device: %v", err)
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(stat.Size(), "flashing")
	if _, err = io.Copy(io.MultiWriter(out, bar), in); err != nil {
		return err
	}

	utils.Success("image flashed on %s device\n", devicePath)
	return nil
}
