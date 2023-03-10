package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/posthog/posthog-go"
	"gopkg.in/yaml.v2"
)

const (
	versionPlaceholder = "dev"
)

var (
	version = versionPlaceholder
)

type PosthogEvent string

const (
	InstallPosthogEvent PosthogEvent = "cli_install"
	BuildPosthogEvent   PosthogEvent = "cli_build"
	DeployPosthogEvent  PosthogEvent = "cli_deploy"
)

var posthogClient posthog.Client

type ProjectInfo struct {
	Spec *ProjectInfoSpec
}

type ProjectInfoSpec struct {
	SendMetrics bool
	Cluster     string
	Provider    string
	Owner       *ProjectInfoOwner
}
type ProjectInfoOwner struct {
	ID string
}

type PosthogProperties struct {
	ApplicationName string `json:"applicationName,omitempty"`
	ApplicationID   string `json:"applicationID,omitempty"`
	PackageType     string `json:"packageType,omitempty"`
	PackageName     string `json:"packageName,omitempty"`
	PackageId       string `json:"packageID,omitempty"`
	PackageVersion  string `json:"packageVersion,omitempty"`
	RecipeName      string `json:"recipeName,omitempty"`
	Error           error  `json:"error,omitempty"`
}

func newPosthogClient() (posthog.Client, error) {
	return posthog.NewWithConfig("phc_r0v4jbKz8Rr27mfqgO15AN5BMuuvnU8hCFedd6zpSDy", posthog.Config{
		Endpoint: "https://posthog.plural.sh",
	})
}

func posthogCapture(posthogClient posthog.Client, event PosthogEvent, property PosthogProperties) error {
	if project, err := getProjectInfo(); err == nil && project.Spec != nil && project.Spec.SendMetrics {
		var properties map[string]interface{}
		inrec, err := json.Marshal(property)
		if err != nil {
			return err
		}
		err = json.Unmarshal(inrec, &properties)
		if err != nil {
			return err
		}
		properties["clusterName"] = project.Spec.Cluster
		properties["provider"] = project.Spec.Provider
		properties["cliVersion"] = version
		userID := "cli-user"
		if project.Spec.Owner != nil {
			userID = project.Spec.Owner.ID
		}
		LogInfo().Printf("send posthog event %v \n", properties)
		return posthogClient.Enqueue(posthog.Capture{
			DistinctId: userID,
			Event:      string(event),
			Properties: properties,
		})
	}
	LogInfo().Println("sending events disabled")
	return nil
}

func PosthogCapture(event PosthogEvent, property PosthogProperties) {
	if posthogClient == nil {
		var err error
		posthogClient, err = newPosthogClient()
		if err != nil {
			LogError().Printf("Failed to create posthog client %v", err)
			return
		}
	}
	if err := posthogCapture(posthogClient, event, property); err != nil {
		LogError().Printf("Failed to send posthog event %v", err)
	}
}

func getProjectInfo() (*ProjectInfo, error) {
	root, found := ProjectRoot()
	if found {
		path := pathing.SanitizeFilepath(filepath.Join(root, "workspace.yaml"))
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("could not find workspace.yaml file")
		}
		var projectInfo ProjectInfo
		err = yaml.Unmarshal(contents, &projectInfo)
		if err != nil {
			return nil, err
		}
		return &projectInfo, nil
	}

	return nil, fmt.Errorf("you are not in the project directory")
}
