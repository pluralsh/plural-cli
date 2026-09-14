package stacks

import (
	"fmt"

	gqlclient "github.com/pluralsh/console/go/client"
)

// StackRunsLister is the Console surface needed to resolve terraform backend URLs.
type StackRunsLister interface {
	ListStackRuns(stackID string) (*gqlclient.ListStackRuns, error)
}

func GetTerraformStateUrls(client StackRunsLister, stackID string) (*gqlclient.TerraformStateUrls, error) {
	stackRuns, err := client.ListStackRuns(stackID)
	if err != nil {
		return nil, err
	}

	if stackRuns == nil ||
		stackRuns.InfrastructureStack == nil ||
		stackRuns.InfrastructureStack.Runs == nil ||
		len(stackRuns.InfrastructureStack.Runs.Edges) == 0 {
		return nil, fmt.Errorf("no terraform state urls found for stack %s", stackID)
	}

	stateUrls := toTerraformStateUrls(stackRuns.InfrastructureStack.Runs.Edges)
	if stateUrls == nil {
		return nil, fmt.Errorf("no terraform state urls found for stack %s", stackID)
	}

	return stateUrls, nil
}

func toTerraformStateUrls(stackRuns []*gqlclient.ListStackRuns_InfrastructureStack_Runs_Edges) *gqlclient.TerraformStateUrls {
	for _, edge := range stackRuns {
		if edge == nil || edge.Node == nil {
			continue
		}
		run := edge.Node
		if run.Type != gqlclient.StackTypeTerraform || run.StateUrls == nil || run.StateUrls.Terraform == nil {
			continue
		}

		return &gqlclient.TerraformStateUrls{
			Address: run.StateUrls.Terraform.Address,
			Lock:    run.StateUrls.Terraform.Lock,
			Unlock:  run.StateUrls.Terraform.Unlock,
		}
	}

	return nil
}
