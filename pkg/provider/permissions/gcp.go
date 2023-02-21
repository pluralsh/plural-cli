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

var gcpExpected = []string{
	"storage.buckets.create",
	"storage.buckets.setIamPolicy",
	"iam.serviceAccounts.create",
	"iam.serviceAccounts.setIamPolicy",
	"container.clusters.create",
	"compute.networks.create",
	"compute.subnetworks.create",
}

func NewGcpChecker(ctx context.Context, project string) (*GcpChecker, error) {
	return &GcpChecker{project, ctx}, nil
}

func (g *GcpChecker) MissingPermissions() (result []string, err error) {
	svc, err := resourcemanager.NewProjectsClient(g.ctx)
	if err != nil {
		return
	}

	res, err := svc.TestIamPermissions(g.ctx, &iampb.TestIamPermissionsRequest{
		Resource:    fmt.Sprintf("projects/%s", g.project),
		Permissions: gcpExpected,
	})
	if err != nil {
		return
	}

	has := containers.ToSet(res.Permissions)
	result = containers.ToSet(gcpExpected).Difference(has).List()
	return
}
