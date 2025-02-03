package edge

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

const (
	cloudConfigURL     = "https://raw.githubusercontent.com/pluralsh/edge/main/cloud-config.yaml"
	pluralConfigURL    = "https://raw.githubusercontent.com/pluralsh/edge/main/plural-config.yaml"
	buildDir           = "build"
	cloudConfigFile    = "cloud-config.yaml"
	volumeName         = "edge-rootfs"
	volumeMountPath    = "/rootfs"
	volumeMount        = "source=edge-rootfs,target=/rootfs"
	wifiConfigTemplate = `
stages:
  boot:
    - name: Setup Wi-Fi
      commands:
        - connmanctl enable wifi
        - wpa_passphrase '@WIFI_SSID@' '@WIFI_PASSWORD@' > /etc/wpa_supplicant/wpa_supplicant.conf
        - wpa_supplicant -B -i wlan0 -c /etc/wpa_supplicant/wpa_supplicant.conf
        - udhcpc -i wlan0 &`
	defaults = `#cloud-config
stages:
  boot:
    - name: Delete default Kairos user
      commands:
        - deluser --remove-home kairos`
)

type Configuration struct {
	Image           string            `json:"image"`
	AurorabootImage string            `json:"aurorabootImage"`
	CraneImage      string            `json:"craneImage"`
	Bundles         map[string]string `json:"bundles"`
}

func (p *Plural) handleEdgeImage(c *cli.Context) error {
	outputDir := c.String("output-dir")
	device := c.String("device")
	project := c.String("project")
	user := c.String("user")
	pluralConfig := c.String("plural-config")
	cloudConfig := c.String("cloud-config")
	username := c.String("username")
	password := c.String("password")
	wifiSsid := c.String("wifi-ssid")
	wifiPassword := c.String("wifi-password")

	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	utils.Highlight("creating bootstrap token for %s project\n", project)
	token, err := p.createBootstrapToken(project, user)
	if err != nil {
		return err
	}

	utils.Highlight("reading configuration\n")
	config, err := p.readPluralConfig(pluralConfig)
	if err != nil {
		return err
	}

	utils.Highlight("preparing output directory\n")
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	outputDirPath := filepath.Join(currentDir, outputDir)
	if err = os.MkdirAll(outputDirPath, os.ModePerm); err != nil {
		return err
	}

	buildDirPath := filepath.Join(outputDirPath, buildDir)
	if err = os.MkdirAll(buildDirPath, os.ModePerm); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(buildDirPath)
	}()

	utils.Highlight("writing configuration\n")
	cloudConfigPath := filepath.Join(outputDirPath, cloudConfigFile)
	if err = p.writeCloudConfig(token, username, password, wifiSsid, wifiPassword, cloudConfigPath, cloudConfig); err != nil {
		return err
	}

	utils.Highlight("overwriting default configuration to remove default user\n")
	defaultsPath := filepath.Join(outputDirPath, "defaults.yaml")
	if err := utils.WriteFile(defaultsPath, []byte(defaults)); err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(defaultsPath)
	}()

	utils.Highlight("preparing %s volume\n", volumeName)
	if err = utils.Exec("docker", "volume", "create", volumeName); err != nil {
		return err
	}
	defer func() {
		utils.Highlight("removing %s volume\n", volumeName)
		_ = utils.Exec("docker", "volume", "rm", volumeName)
	}()

	for bundle, image := range config.Bundles {
		utils.Highlight("writing %s bundle\n", bundle)
		if err = utils.Exec(
			"docker", "run", "-i", "--rm", "--user", "root", "--mount", volumeMount,
			config.CraneImage, "--platform=linux/arm64", "pull", image, fmt.Sprintf("%s/%s.tar", volumeMountPath, bundle)); err != nil {
			return err
		}
	}

	utils.Highlight("unpacking image contents\n")
	if err = utils.Exec("docker", "run", "-i", "--rm", "--privileged", "--mount", volumeMount,
		"quay.io/luet/base", "util", "unpack", config.Image, volumeMountPath); err != nil {
		return err
	}

	utils.Highlight("building image\n")
	if err = utils.Exec("docker", "run", "-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", buildDirPath+":/tmp/build",
		"-v", cloudConfigPath+":/cloud-config.yaml",
		"-v", defaultsPath+":/defaults.yaml",
		"--mount", volumeMount,
		"--privileged", "-i", "--rm",
		"--entrypoint=/build-arm-image.sh", config.AurorabootImage,
		"--model", "rpi4",
		"--directory", volumeMountPath,
		"--config", "/cloud-config.yaml", "/tmp/build/kairos.img"); err != nil {
		return err
	}

	if err = utils.CopyDir(buildDirPath, outputDirPath); err != nil {
		return fmt.Errorf("cannot move output files: %v", err)
	}
	utils.Success("image saved to %s directory\n", outputDir)

	if device != "" {
		if err = p.flashImage(filepath.Join(outputDirPath, "kairos.img"), device); err != nil {
			return err
		}
		utils.Success("image flashed on %s device\n", device)
	}

	return nil
}

func (p *Plural) createBootstrapToken(project, user string) (string, error) {
	attrributes := gqlclient.BootstrapTokenAttributes{}

	if user != "" {
		usr, err := p.ConsoleClient.GetUser(user)
		if err != nil {
			return "", err
		}
		if usr == nil {
			return "", fmt.Errorf("cannot find %s user", user)
		}
		attrributes.UserID = &usr.ID
	}

	proj, err := p.ConsoleClient.GetProject(project)
	if err != nil {
		return "", err
	}
	if proj == nil {
		return "", fmt.Errorf("cannot find %s project", project)
	}
	attrributes.ProjectID = proj.ID

	return p.ConsoleClient.CreateBootstrapToken(attrributes)
}

func (p *Plural) readPluralConfig(override string) (config *Configuration, err error) {
	if override != "" {
		err = utils.YamlFile(override, &config)
	} else {
		err = utils.RemoteYamlFile(pluralConfigURL, &config)
	}
	return config, err
}

func (p *Plural) writeCloudConfig(token, username, password, wifiSsid, wifiPassword, path, override string) error {
	if override != "" {
		return utils.CopyFile(override, path)
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty when cloud config is not specified")
	}

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
	template = strings.ReplaceAll(template, "@TOKEN@", token)

	if wifiSsid != "" && wifiPassword != "" {
		wifiConfig := strings.ReplaceAll(wifiConfigTemplate, "@WIFI_SSID@", wifiSsid)
		wifiConfig = strings.ReplaceAll(wifiConfig, "@WIFI_PASSWORD@", wifiPassword)
		template += "\n" + wifiConfig
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(template)
	return err
}
