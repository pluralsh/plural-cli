package api

import (
	"os"
	"path"
	"path/filepath"

	"github.com/pluralsh/gqlclient"

	"github.com/pluralsh/gqlclient/pkg/utils"
	tarutils "github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/samber/lo"
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

func (client *client) GetTerraformVersions(id string) ([]*Version, error) {
	resp, err := client.pluralClient.GetTerraformVersions(client.ctx, id)
	if err != nil {
		return nil, err
	}

	versions := make([]*Version, 0)
	for _, version := range resp.Versions.Edges {
		versions = append(versions, convertVersion(version.Node))
	}

	return versions, nil
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

func tarDir(name, dir, regex string) (res string, err error) {
	fullPath, err := filepath.Abs(dir)
	if err != nil {
		return
	}

	cwd, _ := os.Getwd()
	res = pathing.SanitizeFilepath(filepath.Join(cwd, name+".tgz"))
	f, err := os.Create(res)
	if err != nil {
		return
	}
	defer f.Close()

	err = tarutils.Tar(fullPath, f, regex)
	return
}

func (client *client) UploadTerraform(dir, repoName string) (Terraform, error) {
	tf := Terraform{}
	name := path.Base(dir)
	tarFile, err := tarDir(name, dir, "\\.terraform")
	if err != nil {
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

	upload := resp.UploadTerraform
	tf.Name = lo.FromPtr(upload.Name)
	tf.Id = lo.FromPtr(upload.ID)
	tf.Description = lo.FromPtr(upload.Description)
	tf.ValuesTemplate = lo.FromPtr(upload.ValuesTemplate)
	tf.Package = lo.FromPtr(upload.Package)
	if upload.Dependencies != nil {
		tf.Dependencies = convertDependencies(upload.Dependencies)
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
