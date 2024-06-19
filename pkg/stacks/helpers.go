package stacks

import (
	"fmt"

	gqlclient "github.com/pluralsh/console-client-go"

	"github.com/pluralsh/plural-cli/pkg/console"
)

func GetTerraformStateUrls(client console.ConsoleClient, stackID string) (*gqlclient.TerraformStateUrls, error) {
	stackRuns, err := client.ListStackRuns(stackID)
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

func toTerraformStateUrls(stackRuns []*gqlclient.ListStackRuns_InfrastructureStack_Runs_Edges) *gqlclient.TerraformStateUrls {
	for _, edge := range stackRuns {
		run := edge.Node
		if run.Type != gqlclient.StackTypeTerraform || run.StateUrls.Terraform == nil {
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
