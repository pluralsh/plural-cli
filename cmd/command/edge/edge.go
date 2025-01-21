package edge

import (
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/urfave/cli"
	"helm.sh/helm/v3/pkg/action"
)

var consoleToken string
var consoleURL string

type Plural struct {
	client.Plural
	HelmConfiguration *action.Configuration
}

func init() {
	consoleToken = ""
	consoleURL = ""
}

func Command(clients client.Plural, helmConfiguration *action.Configuration) cli.Command {
	return cli.Command{
		Name:        "edge",
		Usage:       "manage edge clusters",
		Subcommands: Commands(clients, helmConfiguration),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "token",
				Usage:       "console token",
				EnvVar:      "PLURAL_CONSOLE_TOKEN",
				Destination: &consoleToken,
			},
			cli.StringFlag{
				Name:        "url",
				Usage:       "console url address",
				EnvVar:      "PLURAL_CONSOLE_URL",
				Destination: &consoleURL,
			},
		},
		Category: "Edge",
	}
}

func Commands(clients client.Plural, helmConfiguration *action.Configuration) []cli.Command {
	p := Plural{
		HelmConfiguration: helmConfiguration,
		Plural:            clients,
	}
	return []cli.Command{
		{
			Name:   "image",
			Action: p.handleEdgeImage,
			Usage:  "prepares image ready to be used on Raspberry Pi 4",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "username",
					Usage:    "name for the initial user account",
					Value:    "plural",
					Required: false,
				},
				cli.StringFlag{
					Name:     "password",
					Usage:    "password for the initial user account",
					Required: true,
				},
				cli.StringFlag{
					Name:     "wifi-ssid",
					Usage:    "ssid of the wifi network",
					Required: false,
				},
				cli.StringFlag{
					Name:     "wifi-password",
					Usage:    "password for the wifi network",
					Required: false,
				},
				cli.StringFlag{
					Name:     "output-dir",
					Usage:    "output directory where the image will be stored",
					Value:    "image",
					Required: false,
				},
				cli.StringFlag{
					Name:     "cloud-config",
					Usage:    "path to cloud configuration file, if provided templating will be skipped",
					Required: false,
				},
				cli.StringFlag{
					Name:     "plural-config",
					Usage:    "path to plural configuration file",
					Required: false,
				},
			},
		},
		{
			Name:   "bootstrap",
			Action: p.handleEdgeBootstrap,
			Usage:  "registers edge cluster and installs agent onto it using the current kubeconfig",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "machine-id",
					Usage:    "the unique id of the edge device on which this cluster runs",
					Required: true,
				},
				cli.StringFlag{
					Name:     "project",
					Usage:    "the project this cluster will belong to, if bootstrap token is used then it will inferred from it",
					Required: false,
				},
			},
		},
	}
}
