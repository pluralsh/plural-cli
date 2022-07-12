package api

import "github.com/pluralsh/gqlclient"

type VersionSpec struct {
	Repository string
	Chart      *string
	Terraform  *string
	Version    string
}

type TagAttributes struct {
	Tag string
}

type VersionAttributes struct {
	Tags []*TagAttributes
}

func (client *Client) UpdateVersion(spec *VersionSpec, tags []string) error {
	tagAttrs := make([]*gqlclient.VersionTagAttributes, 0)
	for _, tag := range tags {
		tagAttrs = append(tagAttrs, &gqlclient.VersionTagAttributes{
			Tag: tag,
		})
	}
	_, err := client.pluralClient.UpdateVersion(client.ctx, &gqlclient.VersionSpec{
		Chart:      spec.Chart,
		Repository: &spec.Repository,
		Terraform:  spec.Terraform,
		Version:    &spec.Version,
	}, gqlclient.VersionAttributes{
		Tags: tagAttrs,
	})
	if err != nil {
		return err
	}
	return nil
}
