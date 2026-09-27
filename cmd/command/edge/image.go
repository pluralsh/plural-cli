package edge

import (
	"fmt"
	"os"
	"time"

	"github.com/pluralsh/plural-cli/pkg/console"
	pkgedge "github.com/pluralsh/plural-cli/pkg/edge"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

func (p *Plural) handleEdgeImage(c *cli.Context) error {
	options := pkgedge.ImageOptions{
		OutputDir:    c.String("output-dir"),
		Project:      c.String("project"),
		User:         c.String("user"),
		PluralConfig: c.String("plural-config"),
		CloudConfig:  c.String("cloud-config"),
		Username:     c.String("username"),
		Password:     c.String("password"),
		WifiSSID:     c.String("wifi-ssid"),
		WifiPassword: c.String("wifi-password"),
		Model:        c.String("model"),
		OCIURL:       c.String("oci-url"),
	}

	var client pkgedge.ConsoleAPI
	if options.CloudConfig == "" {
		if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
			return err
		}
		client = p.ConsoleClient
		url := consoleURL
		if url == "" {
			url = console.ReadConfig().Url
		}
		options.ConsoleURL = url
	}

	return pkgedge.NewService(client).Build(options)
}

func (p *Plural) handleEdgeDownload(c *cli.Context) error {
	outputDir := c.String("to")
	if outputDir == "" {
		var err error
		outputDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	done := make(chan struct{})
	defer close(done)
	utils.Highlight("unpacking image contents to %s   ", outputDir)
	go progress(done)
	return pkgedge.Download(pkgedge.DownloadOptions{OCIURL: c.String("oci-url"), To: outputDir}, func(string) {})
}

func progress(done <-chan struct{}) {
	frames := []rune{'|', '/', '-', '\\'}
	i := 0
	for {
		select {
		case <-done:
			return
		default:
			fmt.Printf("\b%c", frames[i%len(frames)])
			time.Sleep(500 * time.Millisecond)
			i++
		}
	}
}
