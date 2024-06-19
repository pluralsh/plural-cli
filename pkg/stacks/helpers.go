package stacks

import (
	"context"
	"fmt"

	console "github.com/pluralsh/console-client-go"
	"github.com/samber/lo"
)

func GetTerraformStateUrls(client console.ConsoleClient, stackID string) (*console.TerraformStateUrls, error) {
	stackRuns, err := client.ListStackRuns(context.Background(), stackID, nil, nil, lo.ToPtr(int64(100)), nil)
	if err != nil {
		return nil, err
	}

	if stackRuns.InfrastructureStack == nil ||
		stackRuns.InfrastructureStack.Runs == nil ||
		len(stackRuns.InfrastructureStack.Runs.Edges) == 0 {
		return nil, nil
	}

	stateUrls := toTerraformStateUrls(stackRuns.InfrastructureStack.Runs.Edges)
	if stateUrls == nil {
		return nil, fmt.Errorf("no terraform state urls found for stack %s", stackID)
	}

	return stateUrls, nil
}

func toTerraformStateUrls(stackRuns []*console.ListStackRuns_InfrastructureStack_Runs_Edges) *console.TerraformStateUrls {
	for _, edge := range stackRuns {
		run := edge.Node
		if run.Type != console.StackTypeTerraform || run.StateUrls.Terraform == nil {
			continue
		}

		return &console.TerraformStateUrls{
			Address: run.StateUrls.Terraform.Address,
			Lock:    run.StateUrls.Terraform.Lock,
			Unlock:  run.StateUrls.Terraform.Unlock,
		}
	}

	return nil
}
