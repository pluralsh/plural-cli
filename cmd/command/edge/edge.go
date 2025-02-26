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
					Name:  "output-dir",
					Usage: "output directory where the image will be stored",
					Value: "image",
				},
				cli.StringFlag{
					Name:  "project",
					Usage: "name of the project to use",
					Value: "default",
				},
				cli.StringFlag{
					Name:  "user",
					Usage: "email of the user to be the user identity for bootstrap token in audit logs",
				},
				cli.StringFlag{
					Name:  "plural-config",
					Usage: "optional path to plural configuration file",
				},
				cli.StringFlag{
					Name:  "cloud-config",
					Usage: "optional path to cloud configuration file, if provided templating will be skipped",
				},
				cli.StringFlag{
					Name:  "username",
					Usage: "name for the initial user account, used during cloud config templating",
					Value: "plural",
				},
				cli.StringFlag{
					Name:  "password",
					Usage: "password for the initial user account, used during cloud config templating",
				},
				cli.StringFlag{
					Name:  "wifi-ssid",
					Usage: "ssid of the wifi network, needs to be used with wifi-password, used during cloud config templating",
				},
				cli.StringFlag{
					Name:  "wifi-password",
					Usage: "password for the wifi network, needs to be used with wifi-ssid, used during cloud config templating",
				},
				cli.StringFlag{
					Name:  "model",
					Usage: "the board model",
					Value: "rpi5",
				},
			},
		},
		{
			Name:   "flash",
			Action: p.handleEdgeFlash,
			Usage:  "flashes image onto storage device",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "image",
					Usage:    "image file path",
					Required: true,
				},
				cli.StringFlag{
					Name:     "device",
					Usage:    "storage device path",
					Required: true,
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
					Name:     "chart-loc",
					Usage:    "URL or filepath of helm chart tar file. Use if not wanting to install helm chart from default plural repository.",
					Required: false,
				},
			},
		},
	}
}
