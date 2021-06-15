package api

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

const updateVersion = `
mutation Update($spec: VersionSpec, $attributes: VersionAttributes!) {
	updateVersion(spec: $spec, attributes: $attributes) { id }
}
`

func (client *Client) UpdateVersion(spec *VersionSpec, tags []string) error {
	var resp struct {
		UpdateVersion struct {
			Id string
		}
	}

	tagAttrs := make([]*TagAttributes, 0)
	for _, tag := range tags {
		tagAttrs = append(tagAttrs, &TagAttributes{Tag: tag})
	}

	req := client.Build(updateVersion)
	req.Var("spec", spec)
	req.Var("attributes", &VersionAttributes{Tags: tagAttrs})
	return client.Run(req, &resp)
}

