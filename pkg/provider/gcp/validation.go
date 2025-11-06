package gcp

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/serviceusage/apiv1/serviceusagepb"
	"github.com/pluralsh/polly/algorithms"

	"github.com/pluralsh/plural-cli/pkg/provider/permissions"
	provUtils "github.com/pluralsh/plural-cli/pkg/provider/utils"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

type PreflightCheck string

const (
	PreflightCheckEnabledServices           = PreflightCheck("[User] Enabled Services")
	PreflightCheckServiceAccountPermissions = PreflightCheck("[User] Test Permissions")
)

func (in *Provider) validateEnabled() error {
	ctx := context.Background()
	c, err := ServiceUsageClient()
	if err != nil {
		return err
	}

	errEnabled := fmt.Errorf("you don't have necessary services enabled, please run: `gcloud services enable serviceusage.googleapis.com cloudresourcemanager.googleapis.com container.googleapis.com` with an owner of the project to enable or enable them in the GCP console")
	services := algorithms.Map([]string{
		"serviceusage.googleapis.com",
		"cloudresourcemanager.googleapis.com",
		"container.googleapis.com",
	}, func(name string) string { return fmt.Sprintf("projects/%s/services/%s", in.Project(), name) })
	parent := fmt.Sprintf("projects/%s", in.Project())
	req := &serviceusagepb.BatchGetServicesRequest{
		Parent: parent,
		Names:  services,
	}
	resp, err := c.BatchGetServices(ctx, req)
	if err != nil {
		utils.LogError().Println(err)
		return fmt.Errorf("could not fetch services information for project %s, make sure your service account does have appropriate permissions", in.Project())
	}

	missing := algorithms.Filter(resp.Services, func(svc *serviceusagepb.Service) bool {
		return svc.State != serviceusagepb.State_ENABLED
	})

	if len(missing) > 0 {
		services := algorithms.Map(missing, func(svc *serviceusagepb.Service) string { return svc.Name })
		enableReq := &serviceusagepb.BatchEnableServicesRequest{
			Parent:     parent,
			ServiceIds: services,
		}
		utils.LogError().Printf("Attempting to enable services %v", services)
		if err := tryToEnableServices(ctx, c, enableReq); err != nil {
			return errEnabled
		}
	}

	return nil
}

func (in *Provider) validatePermissions() error {
	utils.LogInfo().Println("Validate GCP roles/permissions")
	ctx := context.Background()

	projectID, err := in.project()
	if err != nil {
		return err
	}

	checker, _ := permissions.NewGcpChecker(ctx, projectID)
	missing, err := checker.MissingPermissions()
	if err != nil {
		return err
	}

	if len(missing) == 0 {
		return nil
	}

	utils.Error("\u2700\n")
	for _, perm := range missing {
		utils.LogError().Printf("Recommended GCP permissions %s \n", perm)
		provUtils.FailedPermission(perm)
	}

	return fmt.Errorf(
		"your GCP user is missing permissions for project %s: %s, if you aren't comfortable granting these permissions, consider creating a separate GCP project for plural resources and adding required roles to your identity",
		projectID,
		strings.Join(missing, ", "),
	)
}
