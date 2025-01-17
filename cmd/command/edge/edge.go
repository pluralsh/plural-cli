package edge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/console/errors"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/samber/lo"
	"github.com/urfave/cli"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	rpi4Image = "quay.io/kairos/alpine:3.19-standard-arm64-rpi4-v3.2.4-k3sv1.31.3-k3s1"
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
					Name:     "output-dir",
					Usage:    "output directory where the image will be stored",
					Value:    "image",
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
					Usage:    "the project this cluster will belong to",
					Required: false,
				},
			},
		},
	}
}

func (p *Plural) handleEdgeImage(c *cli.Context) error {
	username := c.String("username")
	password := c.String("password")
	outputDir := c.String("output-dir")

	currentDir, err := os.Getwd()
	outputDirPath := filepath.Join(currentDir, outputDir)
	if err = os.MkdirAll(outputDirPath, os.ModePerm); err != nil {
		return err
	}

	configPath := filepath.Join(outputDirPath, "cloud-config.yaml")
	if err = p.writeCloudConfig(username, password, configPath); err != nil {
		return err
	}

	buildDirPath := filepath.Join(outputDirPath, "build")
	if err = os.MkdirAll(buildDirPath, os.ModePerm); err != nil {
		return err
	}

	if err = exec.Command("docker", "volume", "create", "edge-rootfs").Run(); err != nil {
		return err
	}
	defer exec.Command("docker", "volume", "rm", "edge-rootfs").Run()

	if err = p.writeBundle("ghcr.io/pluralsh/kairos-plural-bundle:0.1.4", "/rootfs/plural-bundle.tar"); err != nil {
		return err
	}

	if err = p.writeBundle("ghcr.io/pluralsh/kairos-plural-images-bundle:0.1.2", "/rootfs/plural-images-bundle.tar"); err != nil {
		return err
	}

	if err = p.writeBundle("ghcr.io/pluralsh/kairos-plural-trust-manager-bundle:0.1.0", "/rootfs/plural-trust-manager-bundle.tar"); err != nil {
		return err
	}

	if err = exec.Command("docker", "run", "-i", "--rm", "--privileged",
		"--mount", "source=edge-rootfs,target=/rootfs", "quay.io/luet/base",
		"util", "unpack", rpi4Image, "/rootfs").Run(); err != nil {
		return err
	}

	if err = exec.Command("docker", "run", "-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", buildDirPath+":/tmp/build",
		"-v", configPath+":/cloud-config.yaml",
		"--mount", "source=edge-rootfs,target=/rootfs",
		"--privileged", "-i", "--rm",
		"--entrypoint=/build-arm-image.sh", "quay.io/kairos/auroraboot:v0.4.3",
		"--model", "rpi4",
		"--directory", "/rootfs",
		"--config", "/cloud-config.yaml", "/tmp/build/kairos.img").Run(); err != nil {
		return err
	}

	utils.Success("successfully saved image to %s directory\n", outputDir)
	return nil
}

func (p *Plural) writeCloudConfig(username, password, path string) error {
	response, err := http.Get("https://raw.githubusercontent.com/pluralsh/edge/main/hack/cloud-config.yaml")
	if err != nil {
		return err
	}

	defer response.Body.Close()
	buffer := new(bytes.Buffer)
	if _, err = buffer.ReadFrom(response.Body); err != nil {
		return err
	}

	template := buffer.String()
	template = strings.ReplaceAll(template, "@USERNAME@", username)
	template = strings.ReplaceAll(template, "@PASSWORD@", password)
	template = strings.ReplaceAll(template, "@URL@", consoleURL)
	template = strings.ReplaceAll(template, "@TOKEN@", consoleToken)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(template)
	return err
}

func (p *Plural) writeBundle(bundleImage, targetPath string) error {
	return exec.Command(
		"docker", "run", "-i", "--rm", "--user", "root", "--mount", "source=edge-rootfs,target=/rootfs",
		"gcr.io/go-containerregistry/crane:latest", "--platform=linux/arm64",
		"pull", bundleImage, targetPath).Run()
}

func (p *Plural) handleEdgeBootstrap(c *cli.Context) error {
	machineID := c.String("machine-id")
	project := c.String("project")

	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	registrationAttributes, err := p.getClusterRegistrationAttributes(machineID, project)
	if err != nil {
		return err
	}

	utils.Highlight("registering new cluster on %s machine\n", machineID)
	_, err = p.ConsoleClient.CreateClusterRegistration(*registrationAttributes) // TODO: Handle the case when it already exists, i.e. after reboot.
	if err != nil {
		return err
	}

	utils.Highlight("waiting for registration to be completed\n")
	var complete bool
	var registration *gqlclient.ClusterRegistrationFragment
	_ = wait.PollUntilContextCancel(context.Background(), 30*time.Second, true, func(_ context.Context) (done bool, err error) {
		complete, registration = p.ConsoleClient.IsClusterRegistrationComplete(machineID)
		return complete, nil
	})

	clusterAttributes, err := p.getClusterAttributes(registration)
	if err != nil {
		return err
	}

	utils.Highlight("creating %s cluster\n", registration.Name)
	cluster, err := p.ConsoleClient.CreateCluster(*clusterAttributes)
	if err != nil {
		if errors.Like(err, "handle") {
			handle := lo.ToPtr(clusterAttributes.Name)
			if clusterAttributes.Handle != nil {
				handle = clusterAttributes.Handle
			}
			return p.ReinstallOperator(c, nil, handle)
		}
		return err
	}

	if cluster.CreateCluster.DeployToken == nil {
		return fmt.Errorf("could not fetch deploy token from cluster")
	}

	url := p.ConsoleClient.ExtUrl()
	if agentUrl, err := p.ConsoleClient.AgentUrl(cluster.CreateCluster.ID); err == nil {
		url = agentUrl
	}

	utils.Highlight("installing agent on %s cluster with %s URL\n", registration.Name, p.ConsoleClient.Url())
	return p.DoInstallOperator(url, *cluster.CreateCluster.DeployToken, "")
}

func (p *Plural) getClusterRegistrationAttributes(machineID, project string) (*gqlclient.ClusterRegistrationCreateAttributes, error) {
	attributes := gqlclient.ClusterRegistrationCreateAttributes{MachineID: machineID}

	if project != "" {
		proj, err := p.ConsoleClient.GetProject(project)
		if err != nil {
			return nil, err
		}
		if proj == nil {
			return nil, fmt.Errorf("cannot find %s project", project)
		}
		attributes.ProjectID = lo.ToPtr(proj.ID)
	}

	return &attributes, nil
}

func (p *Plural) getClusterAttributes(registration *gqlclient.ClusterRegistrationFragment) (*gqlclient.ClusterAttributes, error) {
	attributes := gqlclient.ClusterAttributes{
		Name:   registration.Name,
		Handle: &registration.Handle,
	}

	if registration.Tags != nil {
		attributes.Tags = lo.Map(registration.Tags, func(tag *gqlclient.ClusterTags, index int) *gqlclient.TagAttributes {
			if tag == nil {
				return nil
			}

			return &gqlclient.TagAttributes{
				Name:  tag.Name,
				Value: tag.Value,
			}
		})
	}

	if registration.Metadata != nil {
		metadata, err := json.Marshal(registration.Metadata)
		if err != nil {
			return nil, err
		}
		attributes.Metadata = lo.ToPtr(string(metadata))
	}

	if registration.Project != nil {
		attributes.ProjectID = &registration.Project.ID
	}

	return &attributes, nil
}
