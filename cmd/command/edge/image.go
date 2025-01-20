package edge

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
	"sigs.k8s.io/yaml"
)

const (
	cloudConfigURL  = "https://raw.githubusercontent.com/pluralsh/edge/main/cloud-config.yaml"
	pluralConfigURL = "https://raw.githubusercontent.com/pluralsh/edge/main/plural-config.yaml"
	buildDir        = "build"
	cloudConfigFile = "cloud-config.yaml"
)

type Configuration struct {
	Image   string            `json:"image"`
	Bundles map[string]string `json:"bundles"`
}

func (p *Plural) handleEdgeImage(c *cli.Context) error {
	username := c.String("username")
	password := c.String("password")
	outputDir := c.String("output-dir")
	cloudConfig := c.String("cloud-config")
	pluralConfig := c.String("plural-config")

	utils.Highlight("reading configuration\n")
	config, err := p.readConfig(pluralConfig)
	if err != nil {
		return err
	}

	utils.Highlight("preparing output directory\n")
	currentDir, err := os.Getwd()
	outputDirPath := filepath.Join(currentDir, outputDir)
	if err = os.MkdirAll(outputDirPath, os.ModePerm); err != nil {
		return err
	}

	buildDirPath := filepath.Join(outputDirPath, buildDir)
	if err = os.MkdirAll(buildDirPath, os.ModePerm); err != nil {
		return err
	}

	utils.Highlight("writing configuration\n")
	cloudConfigPath := filepath.Join(outputDirPath, cloudConfigFile)
	if err = p.writeCloudConfig(username, password, cloudConfigPath, cloudConfig); err != nil {
		return err
	}

	utils.Highlight("preparing build volume\n")
	if err = utils.Exec("docker", "volume", "create", "edge-rootfs"); err != nil {
		return err
	}
	defer utils.Exec("docker", "volume", "rm", "edge-rootfs")

	// TODO
	// https://github.com/moby/moby/issues/9527
	utils.Highlight("writing bundles\n")
	if err = p.writeBundles(); err != nil {
		return err
	}

	utils.Highlight("unpacking image contents\n")
	if err = utils.Exec("docker", "run", "-i", "--rm", "--privileged",
		"--mount", "source=edge-rootfs,target=/rootfs", "quay.io/luet/base",
		"util", "unpack", config.Image, "/rootfs"); err != nil {
		return err
	}

	utils.Highlight("building image\n")
	if err = utils.Exec("docker", "run", "-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", buildDirPath+":/tmp/build",
		"-v", cloudConfigPath+":/cloud-config.yaml",
		"--mount", "source=edge-rootfs,target=/rootfs",
		"--privileged", "-i", "--rm",
		"--entrypoint=/build-arm-image.sh", "quay.io/kairos/auroraboot:v0.4.3",
		"--model", "rpi4",
		"--directory", "/rootfs",
		"--config", "/cloud-config.yaml", "/tmp/build/kairos.img"); err != nil {
		return err
	}

	utils.Success("image saved to %s directory\n", outputDir)
	return nil
}

func (p *Plural) readConfig(override string) (*Configuration, error) {
	var content []byte
	var err error
	if override == "" {
		content, err = p.readDefaultConfig()
	} else {
		content, err = p.readFile(override)
	}
	if err != nil {
		return nil, err
	}

	var config *Configuration
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (p *Plural) readDefaultConfig() ([]byte, error) {
	response, err := http.Get(pluralConfigURL)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	buffer := new(bytes.Buffer)
	if _, err = buffer.ReadFrom(response.Body); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (p *Plural) readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func (p *Plural) writeCloudConfig(username, password, path, override string) error { // TODO: Override.
	response, err := http.Get(cloudConfigURL)
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

func (p *Plural) writeBundles() error {
	if err := p.writeBundle("ghcr.io/pluralsh/kairos-plural-bundle:0.1.4", "/rootfs/plural-bundle.tar"); err != nil {
		return err
	}

	if err := p.writeBundle("ghcr.io/pluralsh/kairos-plural-images-bundle:0.1.2", "/rootfs/plural-images-bundle.tar"); err != nil {
		return err
	}

	if err := p.writeBundle("ghcr.io/pluralsh/kairos-plural-trust-manager-bundle:0.1.0", "/rootfs/plural-trust-manager-bundle.tar"); err != nil {
		return err
	}

	return nil
}

func (p *Plural) writeBundle(bundleImage, targetPath string) error {
	return utils.Exec(
		"docker", "run", "-i", "--rm", "--user", "root", "--mount", "source=edge-rootfs,target=/rootfs",
		"gcr.io/go-containerregistry/crane:latest", "--platform=linux/arm64",
		"pull", bundleImage, targetPath)
}
