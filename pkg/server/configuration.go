package server

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils/git"
)

type Configuration struct {
	Workspace WorkspaceConfiguration `json:"workspace"`
	Git       GitConfiguration       `json:"git"`
}

type WorkspaceConfiguration struct {
	Network      *NetworkConfiguration `json:"network,omitempty"`
	BucketPrefix string                `json:"bucket_prefix,omitempty"`
	Cluster      string                `json:"cluster,omitempty"`
}

type NetworkConfiguration struct {
	PluralDns bool   `json:"plural_dns,omitempty"`
	Subdomain string `json:"subdomain,omitempty"`
}

type GitConfiguration struct {
	Url    string `json:"url,omitempty"`
	Root   string `json:"root,omitempty"`
	Name   string `json:"name,omitempty"`
	Branch string `json:"branch,omitempty"`
}

func configuration(c *gin.Context) error {
	path := manifest.ProjectManifestPath()
	project, err := manifest.ReadProject(path)
	if err != nil {
		return err
	}
	configuration := Configuration{
		Workspace: WorkspaceConfiguration{
			BucketPrefix: project.BucketPrefix,
			Cluster:      project.Cluster,
		},
	}
	if project.Network != nil {
		configuration.Workspace.Network = &NetworkConfiguration{
			PluralDns: project.Network.PluralDns,
			Subdomain: project.Network.Subdomain,
		}
	}
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}
	branch, err := git.CurrentBranch()
	if err != nil {
		return err
	}
	url, err := git.GetURL()
	if err != nil {
		return err
	}

	configuration.Git = GitConfiguration{
		Url:    url,
		Root:   repoRoot,
		Name:   filepath.Base(repoRoot),
		Branch: branch,
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, configuration)
	return nil
}
