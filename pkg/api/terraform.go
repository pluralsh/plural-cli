package api

import (
	"os"
	"path"
	"path/filepath"

	"github.com/pluralsh/gqlclient"

	"github.com/pluralsh/gqlclient/pkg/utils"
	tarutils "github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func (client *client) GetTerraforma(repoId string) ([]*Terraform, error) {

	terraformResponse, err := client.pluralClient.GetTerraform(client.ctx, repoId)
	if err != nil {
		return nil, err
	}

	terraform := make([]*Terraform, 0)
	for _, edge := range terraformResponse.Terraform.Edges {
		terraform = append(terraform, convertTerraform(edge.Node))
	}
	return terraform, err
}

func (client *client) GetTerraformInstallations(repoId string) ([]*TerraformInstallation, error) {
	resp, err := client.pluralClient.GetTerraformInstallations(client.ctx, repoId)
	if err != nil {
		return nil, err
	}

	inst := make([]*TerraformInstallation, 0)
	for _, edge := range resp.TerraformInstallations.Edges {
		inst = append(inst, &TerraformInstallation{
			Id:        utils.ConvertStringPointer(edge.Node.ID),
			Terraform: convertTerraform(edge.Node.Terraform),
			Version:   convertVersion(edge.Node.Version),
		})
	}
	return inst, err
}

func (client *client) UninstallTerraform(id string) (err error) {
	_, err = client.pluralClient.UninstallTerraform(client.ctx, id)
	return
}

func (client *client) UploadTerraform(dir, repoName string) (Terraform, error) {
	name := path.Base(dir)
	fullPath, err := filepath.Abs(dir)
	tf := Terraform{}
	if err != nil {
		return tf, err
	}
	cwd, _ := os.Getwd()
	tarFile := pathing.SanitizeFilepath(filepath.Join(cwd, name+".tgz"))
	f, err := os.Create(tarFile)
	if err != nil {
		return tf, err
	}
	defer f.Close()

	if err := tarutils.Tar(fullPath, f, "\\.terraform"); err != nil {
		return tf, err
	}
	rf, err := os.Open(tarFile)
	if err != nil {
		return tf, err
	}
	defer rf.Close()
	defer os.Remove(tarFile)

	resp, err := client.pluralClient.UploadTerraform(client.ctx, repoName, name, "package", gqlclient.WithFiles([]gqlclient.Upload{
		{
			Field: "package",
			Name:  tarFile,
			R:     rf,
		},
	}))
	if err != nil {
		return tf, err
	}

	if resp.UploadTerraform.Name != nil {
		tf.Name = *resp.UploadTerraform.Name
	}
	if resp.UploadTerraform.ID != nil {
		tf.Id = *resp.UploadTerraform.ID
	}
	if resp.UploadTerraform.Description != nil {
		tf.Description = *resp.UploadTerraform.Description
	}
	if resp.UploadTerraform.ValuesTemplate != nil {
		tf.ValuesTemplate = *resp.UploadTerraform.ValuesTemplate
	}
	if resp.UploadTerraform.Package != nil {
		tf.Package = *resp.UploadTerraform.Package
	}
	if resp.UploadTerraform.Dependencies != nil {
		tf.Dependencies = convertDependencies(resp.UploadTerraform.Dependencies)
	}

	return tf, err
}

func convertTerraform(ter *gqlclient.TerraformFragment) *Terraform {
	if ter == nil {
		return nil
	}
	return &Terraform{
		Id:             utils.ConvertStringPointer(ter.ID),
		Name:           utils.ConvertStringPointer(ter.Name),
		Description:    utils.ConvertStringPointer(ter.Description),
		ValuesTemplate: utils.ConvertStringPointer(ter.ValuesTemplate),
		Dependencies:   convertDependencies(ter.Dependencies),
		Package:        utils.ConvertStringPointer(ter.Package),
	}
}
