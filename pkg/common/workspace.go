package common

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
	v1 "k8s.io/api/core/v1"
)

func KubeInit(_ *cli.Context) error {
	_, found := utils.ProjectRoot()
	if !found {
		return fmt.Errorf("Project not initialized, run `plural init` to set up a workspace")
	}

	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}

	return prov.KubeConfig()
}

func PrintListNodes(nodes *v1.NodeList) error {
	headers := []string{"Name", "CPU", "Memory", "Region", "Zone"}
	return utils.PrintTable(nodes.Items, headers, func(node v1.Node) ([]string, error) {
		status := node.Status
		labels := node.Labels
		cpu, mem := status.Capacity["cpu"], status.Capacity["memory"]
		return []string{
			node.Name,
			cpu.String(),
			mem.String(),
			labels["topology.kubernetes.io/region"],
			labels["topology.kubernetes.io/zone"],
		}, nil
	})
}
