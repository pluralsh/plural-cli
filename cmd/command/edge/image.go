package edge

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/console"
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
	project := c.String("project")
	user := c.String("user")
	pluralConfig := c.String("plural-config")
	cloudConfig := c.String("cloud-config")
	username := c.String("username")
	password := c.String("password")
	wifiSsid := c.String("wifi-ssid")
	wifiPassword := c.String("wifi-password")
	model := c.String("model")
	imagePushURL := c.String("image-push-url")

	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
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
	if err = p.writeCloudConfig(project, user, username, password, wifiSsid, wifiPassword, cloudConfigPath, cloudConfig); err != nil {
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
		"--model", model,
		"--directory", volumeMountPath,
		"--config", "/cloud-config.yaml", "/tmp/build/kairos.img"); err != nil {
		return err
	}

	if imagePushURL != "" {
		dockerfile := "FROM scratch\nWORKDIR /build\nCOPY kairos.img /build"
		dockerfilePath := filepath.Join(buildDirPath, "Dockerfile")
		if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
			return fmt.Errorf("cannot create Docker file: %w", err)
		}
		if err = utils.Exec("docker", "build", "-t", imagePushURL, "-f", dockerfilePath, buildDirPath); err != nil {
			return err
		}
		if err = utils.Exec("docker", "push", imagePushURL); err != nil {
			return err
		}

		utils.Success("image pushed successfully to %s\n", imagePushURL)
	}

	if err = utils.CopyDir(buildDirPath, outputDirPath); err != nil {
		return fmt.Errorf("cannot move output files: %w", err)
	}

	utils.Success("image saved to %s directory\n", outputDir)

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

func (p *Plural) writeCloudConfig(project, user, username, password, wifiSsid, wifiPassword, path, override string) error {
	if override != "" {
		return utils.CopyFile(override, path)
	}

	url := consoleURL
	if url == "" {
		url = console.ReadConfig().Url // Read URL from config if it was not provided via args or env var
	}

	token, err := p.createBootstrapToken(project, user)
	if err != nil {
		return err
	}

	if url == "" {
		return fmt.Errorf("url cannot be empty when cloud config is not specified")
	}

	if token == "" {
		return fmt.Errorf("token cannot be empty when cloud config is not specified")
	}

	if username == "" {
		return fmt.Errorf("username cannot be empty when cloud config is not specified")
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
	template = strings.ReplaceAll(template, "@URL@", url)
	template = strings.ReplaceAll(template, "@TOKEN@", token)
	template = strings.ReplaceAll(template, "@USERNAME@", username)
	template = strings.ReplaceAll(template, "@PASSWORD@", password)

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

func (p *Plural) handleEdgeDownload(c *cli.Context) error {
	outputDir := c.String("to")
	url := c.String("url")

	utils.Highlight("unpacking image contents\n")
	if err := utils.Exec("docker", "run", "-i", "--rm", "--privileged", "-v", fmt.Sprintf("%s:/image", outputDir),
		"quay.io/luet/base", "util", "unpack", url, "/image"); err != nil {
		return err
	}

	return nil
}
