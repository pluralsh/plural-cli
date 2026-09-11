package edge

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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
	var err error

	outputDir := c.String("to")
	url := c.String("oci-url")

	if outputDir == "" {
		outputDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	return unpackImage(outputDir, url)
}

func unpackImage(outputDir, ociUrl string) error {
	done := make(chan struct{})
	defer close(done)

	utils.Highlight("unpacking image contents to %s   ", outputDir)
	imageDir := filepath.Join(outputDir, "build")
	if !utils.IsDir(imageDir) {
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return err
		}
	}

	ref, err := name.ParseReference(ociUrl)
	if err != nil {
		return err
	}
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return err
	}
	reader := mutate.Extract(img)
	defer func(reader io.ReadCloser) {
		err := reader.Close()
		if err != nil {
			utils.Error("%s", err.Error())
		}
	}(reader)
	go progress(done)
	return utils.Untar(outputDir, reader)
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
