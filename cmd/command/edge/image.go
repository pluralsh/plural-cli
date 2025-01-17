package edge

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

func (p *Plural) handleEdgeImage(c *cli.Context) error {
	image := c.String("image")
	username := c.String("username")
	password := c.String("password")
	outputDir := c.String("output-dir")
	_ = c.String("override-config")  // TODO
	_ = c.String("override-bundles") // TODO

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

	if err = utils.Exec("docker", "volume", "create", "edge-rootfs"); err != nil {
		return err
	}
	defer utils.Exec("docker", "volume", "rm", "edge-rootfs")

	if err = p.writeBundle("ghcr.io/pluralsh/kairos-plural-bundle:0.1.4", "/rootfs/plural-bundle.tar"); err != nil {
		return err
	}

	if err = p.writeBundle("ghcr.io/pluralsh/kairos-plural-images-bundle:0.1.2", "/rootfs/plural-images-bundle.tar"); err != nil {
		return err
	}

	if err = p.writeBundle("ghcr.io/pluralsh/kairos-plural-trust-manager-bundle:0.1.0", "/rootfs/plural-trust-manager-bundle.tar"); err != nil {
		return err
	}

	if err = utils.Exec("docker", "run", "-i", "--rm", "--privileged",
		"--mount", "source=edge-rootfs,target=/rootfs", "quay.io/luet/base",
		"util", "unpack", image, "/rootfs"); err != nil {
		return err
	}

	if err = utils.Exec("docker", "run", "-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", buildDirPath+":/tmp/build",
		"-v", configPath+":/cloud-config.yaml",
		"--mount", "source=edge-rootfs,target=/rootfs",
		"--privileged", "-i", "--rm",
		"--entrypoint=/build-arm-image.sh", "quay.io/kairos/auroraboot:v0.4.3",
		"--model", "rpi4",
		"--directory", "/rootfs",
		"--config", "/cloud-config.yaml", "/tmp/build/kairos.img"); err != nil {
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
	return utils.Exec(
		"docker", "run", "-i", "--rm", "--user", "root", "--mount", "source=edge-rootfs,target=/rootfs",
		"gcr.io/go-containerregistry/crane:latest", "--platform=linux/arm64",
		"pull", bundleImage, targetPath)
}
