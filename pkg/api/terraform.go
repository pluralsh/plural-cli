package api

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/gqlclient"

	"github.com/pluralsh/gqlclient/pkg/utils"
	tarutils "github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
)

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
