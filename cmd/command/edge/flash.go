package edge

import (
	"fmt"
	"io"
	"os"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

func (p *Plural) handleEdgeFlash(c *cli.Context) error {
	imagePath := c.String("image")
	devicePath := c.String("device")
	blockSize := c.Int("block-size")

	in, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("could not open image: %v", err)
	}
	defer in.Close()

	out, err := os.OpenFile(devicePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("could not open device: %v", err)
	}
	defer out.Close()

	buf := make([]byte, blockSize)
	for {
		n, err := in.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("could not read image: %v", err)
		}
		_, err = out.Write(buf[:n])
		if err != nil {
			return fmt.Errorf("could not write image: %v", err)
		}
	}

	utils.Success("image flashed on %s device\n", devicePath)
	return nil
}
