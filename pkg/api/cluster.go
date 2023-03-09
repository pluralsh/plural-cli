package api

import (
	"github.com/pluralsh/gqlclient"
	"github.com/samber/lo"
)

func (client *client) DestroyCluster(domain, name, provider string) error {
	_, err := client.pluralClient.DestroyCluster(client.ctx, domain, name, gqlclient.Provider(NormalizeProvider(provider)))
	if err != nil {
		return err
	}

	return nil
}

func (client *client) CreateDependency(source, dest string) error {
	_, err := client.pluralClient.CreateDependency(client.ctx, source, dest)
	return err
}

func (client *client) PromoteCluster() error {
	_, err := client.pluralClient.PromoteCluster(client.ctx)
	return err
}

func (client *client) Clusters() ([]*Cluster, error) {
	resp, err := client.pluralClient.Clusters(client.ctx, nil)
	if err != nil {
		return nil, err
	}
	clusters := make([]*Cluster, 0)
	for _, edge := range resp.Clusters.Edges {
		node := edge.Node
		clusters = append(clusters, &Cluster{
			Id:       node.ID,
			Name:     node.Name,
			Provider: string(node.Provider),
			GitUrl:   lo.FromPtr(node.GitURL),
			Owner: &User{
				Id:    node.Owner.ID,
				Name:  node.Owner.Name,
				Email: node.Owner.Email,
			},
		})
	}

	return clusters, nil
}

func (client *client) Cluster(id string) (*Cluster, error) {
	resp, err := client.pluralClient.ClusterInfo(client.ctx, id)
	if err != nil {
		return nil, err
	}

	node := resp.Cluster
	upgradeInfo := make([]*UpgradeInfo, 0)
	for _, info := range node.UpgradeInfo {
		upgradeInfo = append(upgradeInfo, &UpgradeInfo{
			Count: lo.FromPtr(info.Count),
			Installation: &Installation{
				Repository: convertRepository(info.Installation.Repository),
			},
		})
	}

	cluster := &Cluster{
		Id:       node.ID,
		Name:     node.Name,
		Provider: string(node.Provider),
		GitUrl:   lo.FromPtr(node.GitURL),
		Owner: &User{
			Id:    node.Owner.ID,
			Name:  node.Owner.Name,
			Email: node.Owner.Email,
		},
		UpgradeInfo: upgradeInfo,
	}
	return cluster, nil
}
