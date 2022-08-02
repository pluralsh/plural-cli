package api

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/gqlclient/pkg/utils"
	file "github.com/pluralsh/plural/pkg/utils"
	"gopkg.in/yaml.v3"
)

type ArtifactAttributes struct {
	Name     string
	Readme   string
	Type     string
	Platform string
	Blob     string
	Arch     string
}

func (client *client) ListArtifacts(repo string) ([]Artifact, error) {

	result := make([]Artifact, 0)

	resp, err := client.pluralClient.ListArtifacts(client.ctx, repo)
	if err != nil {
		return result, err
	}
	for _, artifact := range resp.Repository.Artifacts {
		ar := Artifact{
			Id:     utils.ConvertStringPointer(artifact.ID),
			Name:   utils.ConvertStringPointer(artifact.Name),
			Readme: utils.ConvertStringPointer(artifact.Readme),
			Blob:   utils.ConvertStringPointer(artifact.Blob),
			Sha:    utils.ConvertStringPointer(artifact.Sha),
			Arch:   utils.ConvertStringPointer(artifact.Arch),
		}
		if artifact.Platform != nil {
			ar.Platform = string(*artifact.Platform)
		}
		if artifact.Filesize != nil {
			ar.Filesize = int(*artifact.Filesize)
		}
		result = append(result, ar)
	}
	return result, nil
}

func (client *client) CreateArtifact(repo string, attrs ArtifactAttributes) (Artifact, error) {
	var artifact Artifact
	fullPath, _ := filepath.Abs(attrs.Blob)
	rf, err := os.Open(fullPath)
	if err != nil {
		return artifact, err
	}
	defer rf.Close()

	readmePath, _ := filepath.Abs(attrs.Readme)
	readme, err := file.ReadFile(readmePath)
	if err != nil {
		return artifact, err
	}

	createArtifact, err := client.pluralClient.CreateArtifact(context.Background(), repo, attrs.Name, readme, attrs.Type, attrs.Platform, "blob", &attrs.Arch, gqlclient.WithFiles([]gqlclient.Upload{{
		Field: "blob",
		Name:  attrs.Blob,
		R:     rf,
	}}))
	if err != nil {
		return artifact, err
	}
	artifact.Id = utils.ConvertStringPointer(createArtifact.CreateArtifact.ID)
	artifact.Name = utils.ConvertStringPointer(createArtifact.CreateArtifact.Name)
	artifact.Readme = utils.ConvertStringPointer(createArtifact.CreateArtifact.Readme)
	artifact.Arch = utils.ConvertStringPointer(createArtifact.CreateArtifact.Arch)
	artifact.Sha = utils.ConvertStringPointer(createArtifact.CreateArtifact.Sha)
	if createArtifact.CreateArtifact.Platform != nil {
		platform := createArtifact.CreateArtifact.Platform
		artifact.Platform = string(*platform)
	}

	return artifact, err
}

func ConstructArtifactAttributes(marshalled []byte) (ArtifactAttributes, error) {
	var attrs ArtifactAttributes
	err := yaml.Unmarshal(marshalled, &attrs)
	return attrs, err
}
