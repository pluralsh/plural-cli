package permissions

import (
	"context"
	"fmt"

	"cloud.google.com/go/iam/apiv1/iampb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"github.com/pluralsh/polly/containers"
)

type GcpChecker struct {
	project string
	ctx     context.Context
}

func (g *GcpChecker) requiredPermissions() []string {
	return []string{
		"compute.globalOperations.get",
		"compute.instanceGroupManagers.get",
		"compute.networks.create",
		"compute.networks.delete",
		"compute.networks.get",
		"compute.networks.updatePolicy",
		"compute.regionOperations.get",
		"compute.regions.get",
		"compute.routers.create",
		"compute.routers.delete",
		"compute.routers.get",
		"compute.subnetworks.create",
		"compute.subnetworks.delete",
		"compute.subnetworks.get",
		"compute.subnetworks.list",
		"compute.zones.list",
		"container.clusters.create",
		"container.clusters.delete",
		"container.clusters.get",
		"container.clusters.getCredentials",
		"container.clusters.update",
		"container.nodes.get",
		"container.nodes.list",
		"container.nodes.update",
		"container.pods.get",
		"iam.serviceAccounts.actAs",
		"iam.serviceAccounts.getAccessToken",
	}
}

func NewGcpChecker(ctx context.Context, project string) (*GcpChecker, error) {
	return &GcpChecker{project, ctx}, nil
}

func (g *GcpChecker) MissingPermissions() (result []string, err error) {
	svc, err := resourcemanager.NewProjectsClient(g.ctx)
	if err != nil {
		return
	}

	defer svc.Close()

	res, err := svc.TestIamPermissions(g.ctx, &iampb.TestIamPermissionsRequest{
		Resource:    fmt.Sprintf("projects/%s", g.project),
		Permissions: g.requiredPermissions(),
	})
	if err != nil {
		return
	}

	has := containers.ToSet(res.Permissions)
	result = containers.ToSet(g.requiredPermissions()).Difference(has).List()
	return
}
