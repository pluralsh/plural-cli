package console

import "github.com/pluralsh/plural/pkg/api"

func (c *consoleClient) ListClusterServices() ([]ServiceDeployment, error) {
	output := []ServiceDeployment{}
	result, err := c.pluralClient.ListClusterServices(c.ctx)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListClusterServices")
	}

	for _, cs := range result.ClusterServices {
		output = append(output, *convertServiceDeployment(cs))
	}
	return output, nil
}
