package console

import "github.com/pluralsh/plural/pkg/api"

func (c *consoleClient) ListClusterServices(clusterId string) ([]ServiceDeployment, error) {
	output := []ServiceDeployment{}
	result, err := c.pluralClient.ListServiceDeployment(c.ctx, nil, nil, nil, &clusterId)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListClusterServices")
	}

	for _, cs := range result.ServiceDeployments.Edges {
		output = append(output, *convertServiceDeployment(cs.Node))
	}
	return output, nil
}
