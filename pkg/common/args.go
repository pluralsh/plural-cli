package common

import (
	"fmt"
	"strings"

	"github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/console"

	"github.com/samber/lo"
)

// ParseServiceIdentifier parses the given identifier and returns the service id, cluster name, and service name.
// If the identifier is in the format @{cluster-handle}/{service-name}, the cluster name and service name are returned.
// Otherwise, the service id is returned.
func ParseServiceIdentifier(id string) (serviceId, clusterName, serviceName *string, err error) {
	if strings.HasPrefix(id, "@") {
		i := strings.Trim(id, "@")
		split := strings.Split(i, "/")
		if len(split) != 2 {
			err = fmt.Errorf("expected format @{cluster-handle}/{service-name} or {service-id}, got %s", id)
			return
		}
		clusterName = &split[0]
		serviceName = &split[1]
	} else {
		serviceId = &id
	}

	return
}

// GetService returns the service deployment for the given identifier.
// Identifier should be in the format of @{cluster-handle}/{service-name} or {service-id}.
func GetService(c console.ConsoleClient, id string) (*client.ServiceDeploymentExtended, error) {
	serviceId, clusterName, serviceName, err := ParseServiceIdentifier(id)
	if err != nil {
		return nil, fmt.Errorf("could not parse identifier: %w", err)
	}

	service, err := c.GetClusterService(serviceId, serviceName, clusterName)
	if err != nil {
		return nil, fmt.Errorf("could not get service deployment: %w", err)
	}

	return service, lo.Ternary(service == nil, fmt.Errorf("could not find service deployment for %s", id), nil)
}
