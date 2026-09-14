// Package edge implements Raspberry Pi image build and flash, shared by CLI and TUI.
package edge

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/utils"
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
	dockerfile = "FROM scratch\nWORKDIR /build\nCOPY kairos.img /build"
)

// Configuration is the plural-config.yaml used to build an edge image.
type Configuration struct {
	Image           string            `json:"image"`
	AurorabootImage string            `json:"aurorabootImage"`
	CraneImage      string            `json:"craneImage"`
	Bundles         map[string]string `json:"bundles"`
}

// ImageOptions are the CLI flags for plural edge image.
type ImageOptions struct {
	OutputDir    string
	Project      string
	User         string
	PluralConfig string
	CloudConfig  string
	Username     string
	Password     string
	WifiSSID     string
	WifiPassword string
	Model        string
	OCIURL       string
	ConsoleURL   string
	WorkingDir   string
}

// ConsoleAPI is the Console subset used to mint a bootstrap token.
type ConsoleAPI interface {
	GetUser(email string) (*gqlclient.UserFragment, error)
	GetProject(name string) (*gqlclient.ProjectFragment, error)
	CreateBootstrapToken(attributes gqlclient.BootstrapTokenAttributes) (string, error)
}

// Service builds edge images using Console and local Docker.
type Service struct {
	Client           ConsoleAPI
	Exec             func(name string, args ...string) error
	Log              func(string)
	LoadConfig       func(override string) (*Configuration, error)
	FetchCloudConfig func() (string, error)
	Output           io.Writer
}

// NewService constructs an image builder. A nil client is allowed when CloudConfig is set.
func NewService(client ConsoleAPI) *Service {
	return &Service{Client: client}
}

func (s *Service) log(msg string) {
	if s != nil && s.Log != nil {
		s.Log(msg)
		return
	}
	utils.Highlight("%s\n", msg)
}

func (s *Service) exec(name string, args ...string) error {
	if s != nil && s.Exec != nil {
		return s.Exec(name, args...)
	}
	cmd := exec.Command(name, args...)
	var stdout, stderr io.Writer = os.Stdout, os.Stderr
	if s != nil && s.Output != nil {
		stdout, stderr = s.Output, s.Output
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func (s *Service) loadConfig(override string) (*Configuration, error) {
	if s != nil && s.LoadConfig != nil {
		return s.LoadConfig(override)
	}
	var config *Configuration
	var err error
	if override != "" {
		err = utils.YamlFile(override, &config)
	} else {
		err = utils.RemoteYamlFile(pluralConfigURL, &config)
	}
	return config, err
}

func (s *Service) fetchCloudConfig() (string, error) {
	if s != nil && s.FetchCloudConfig != nil {
		return s.FetchCloudConfig()
	}
	response, err := http.Get(cloudConfigURL)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Build prepares a Raspberry Pi image the same way as plural edge image.
func (s *Service) Build(options ImageOptions) error {
	if options.Model == "" {
		options.Model = "rpi5"
	}
	if options.OutputDir == "" {
		options.OutputDir = "image"
	}
	if options.Username == "" {
		options.Username = "plural"
	}
	if options.Project == "" {
		options.Project = "default"
	}

	workingDir := options.WorkingDir
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	s.log("reading configuration")
	config, err := s.loadConfig(options.PluralConfig)
	if err != nil {
		return err
	}

	s.log("preparing output directory")
	outputDirPath := options.OutputDir
	if !filepath.IsAbs(outputDirPath) {
		outputDirPath = filepath.Join(workingDir, options.OutputDir)
	}
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

	s.log("writing configuration")
	cloudConfigPath := filepath.Join(outputDirPath, cloudConfigFile)
	if err = s.writeCloudConfig(options, cloudConfigPath); err != nil {
		return err
	}

	s.log("overwriting default configuration to remove default user")
	defaultsPath := filepath.Join(outputDirPath, "defaults.yaml")
	if err := utils.WriteFile(defaultsPath, []byte(defaults)); err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(defaultsPath)
	}()

	s.log("preparing " + volumeName + " volume")
	if err = s.exec("docker", "volume", "create", volumeName); err != nil {
		return err
	}
	defer func() {
		s.log("removing " + volumeName + " volume")
		_ = s.exec("docker", "volume", "rm", volumeName)
	}()

	for bundle, image := range config.Bundles {
		s.log("writing " + bundle + " bundle")
		if err = s.exec(
			"docker", "run", "-i", "--rm", "--user", "root", "--mount", volumeMount,
			config.CraneImage, "--platform=linux/arm64", "pull", image, fmt.Sprintf("%s/%s.tar", volumeMountPath, bundle)); err != nil {
			return err
		}
	}

	s.log("unpacking image contents")
	if err = s.exec("docker", "run", "-i", "--rm", "--privileged", "--mount", volumeMount,
		"quay.io/luet/base", "util", "unpack", config.Image, volumeMountPath); err != nil {
		return err
	}

	s.log("building image")
	if err = s.exec("docker", "run", "-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", buildDirPath+":/tmp/build",
		"-v", cloudConfigPath+":/cloud-config.yaml",
		"-v", defaultsPath+":/defaults.yaml",
		"--mount", volumeMount,
		"--privileged", "-i", "--rm",
		"--entrypoint=/build-arm-image.sh", config.AurorabootImage,
		"--model", options.Model,
		"--directory", volumeMountPath,
		"--config", "/cloud-config.yaml", "/tmp/build/kairos.img"); err != nil {
		return err
	}

	if options.OCIURL != "" {
		dockerfilePath := filepath.Join(buildDirPath, "Dockerfile")
		if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
			return fmt.Errorf("cannot create Dockerfile: %w", err)
		}
		if err = s.exec("docker", "build", "-t", options.OCIURL, "-f", dockerfilePath, buildDirPath); err != nil {
			return err
		}
		if err = s.exec("docker", "push", options.OCIURL); err != nil {
			return err
		}
		s.log("image pushed successfully to " + options.OCIURL)
	}

	if err = utils.CopyDir(buildDirPath, outputDirPath); err != nil {
		return fmt.Errorf("cannot move output files: %w", err)
	}

	s.log("image saved to " + options.OutputDir + " directory")
	return nil
}

func (s *Service) writeCloudConfig(options ImageOptions, path string) error {
	if options.CloudConfig != "" {
		return utils.CopyFile(options.CloudConfig, path)
	}

	token, err := s.createBootstrapToken(options.Project, options.User)
	if err != nil {
		return err
	}
	if options.ConsoleURL == "" {
		return fmt.Errorf("url cannot be empty when cloud config is not specified")
	}
	if token == "" {
		return fmt.Errorf("token cannot be empty when cloud config is not specified")
	}
	if options.Username == "" {
		return fmt.Errorf("username cannot be empty when cloud config is not specified")
	}
	if options.Password == "" {
		return fmt.Errorf("password cannot be empty when cloud config is not specified")
	}

	template, err := s.fetchCloudConfig()
	if err != nil {
		return err
	}
	template = strings.ReplaceAll(template, "@URL@", options.ConsoleURL)
	template = strings.ReplaceAll(template, "@TOKEN@", token)
	template = strings.ReplaceAll(template, "@USERNAME@", options.Username)
	template = strings.ReplaceAll(template, "@PASSWORD@", options.Password)

	if options.WifiSSID != "" && options.WifiPassword != "" {
		wifiConfig := strings.ReplaceAll(wifiConfigTemplate, "@WIFI_SSID@", options.WifiSSID)
		wifiConfig = strings.ReplaceAll(wifiConfig, "@WIFI_PASSWORD@", options.WifiPassword)
		template += "\n" + wifiConfig
	}

	return os.WriteFile(path, []byte(template), 0644)
}

func (s *Service) createBootstrapToken(project, user string) (string, error) {
	if s == nil || s.Client == nil {
		return "", fmt.Errorf("console client is not configured")
	}
	attributes := gqlclient.BootstrapTokenAttributes{}
	if user != "" {
		usr, err := s.Client.GetUser(user)
		if err != nil {
			return "", err
		}
		if usr == nil {
			return "", fmt.Errorf("cannot find %s user", user)
		}
		attributes.UserID = &usr.ID
	}

	proj, err := s.Client.GetProject(project)
	if err != nil {
		return "", err
	}
	if proj == nil {
		return "", fmt.Errorf("cannot find %s project", project)
	}
	attributes.ProjectID = proj.ID

	return s.Client.CreateBootstrapToken(attributes)
}
