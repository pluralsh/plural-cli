package console

import (
	consoleclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/gqlclient/pkg/utils"
	"github.com/pluralsh/plural/pkg/api"
)

func (c *consoleClient) ListClusters() ([]Cluster, error) {
	clusters := []Cluster{}

	result, err := c.pluralClient.ListClusters(c.ctx, nil, nil, nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListClusters")
	}
	for _, edge := range result.Clusters.Edges {
		if edge.Node != nil {
			clusters = append(clusters, convertClusters(edge.Node))
		}
	}
	return clusters, nil
}

func convertClusters(c *consoleclient.ClusterFragment) Cluster {
	newCluster := Cluster{
		Id:             c.ID,
		Name:           c.Name,
		Version:        c.Version,
		CurrentVersion: utils.ConvertStringPointer(c.CurrentVersion),
		NodePools:      []NodePool{},
	}
	if c.Provider != nil {
		newCluster.Provider = convertProvider(c.Provider)
	}
	for _, np := range c.NodePools {
		newCluster.NodePools = append(newCluster.NodePools, NodePool{
			Id:           np.ID,
			Name:         np.Name,
			MinSize:      np.MinSize,
			MaxSize:      np.MaxSize,
			InstanceType: np.InstanceType,
		})
	}

	return newCluster
}

func convertProvider(c *consoleclient.ClusterProviderFragment) *ClusterProvider {
	output := &ClusterProvider{
		Id:        c.ID,
		Name:      c.Name,
		Namespace: c.Namespace,
		Cloud:     c.Cloud,
	}
	if c.Editable != nil {
		output.Editable = *c.Editable
	}
	if c.Repository != nil {
		output.Repository = convertGitRepository(c.Repository)
	}
	if c.Service != nil {
		output.Service = convertServiceDeployment(c.Service)
	}
	return output
}

func convertServiceDeployment(sd *consoleclient.ServiceDeploymentFragment) *ServiceDeployment {
	output := &ServiceDeployment{
		Id:        sd.ID,
		Name:      sd.Name,
		Namespace: sd.Namespace,
		Version:   sd.Version,
		DeletedAt: sd.DeletedAt,
		Git: GitRef{
			Folder: sd.Git.Folder,
			Ref:    sd.Git.Ref,
		},
		Sha:        utils.ConvertStringPointer(sd.Sha),
		Tarball:    utils.ConvertStringPointer(sd.Tarball),
		Components: []Component{},
	}
	if sd.Editable != nil {
		output.Editable = *sd.Editable
	}
	if sd.Repository != nil {
		output.Repository = convertGitRepository(sd.Repository)
	}
	for _, c := range sd.Components {
		output.Components = append(output.Components, Component{
			Id:        c.ID,
			Name:      c.Name,
			Group:     c.Group,
			Kind:      c.Kind,
			Namespace: c.Namespace,
			State: func() ComponentState {
				if c.State != nil {
					return ComponentState(*c.State)
				}
				return ComponentStateUnknown
			}(),
			Synced:  c.Synced,
			Version: c.Version,
		})
	}

	return output
}

func convertGitRepository(gitRepo *consoleclient.GitRepositoryFragment) *GitRepository {
	git := &GitRepository{
		Id:  gitRepo.ID,
		URL: gitRepo.URL,
	}
	if gitRepo.Editable != nil {
		git.Editable = *gitRepo.Editable
	}
	if gitRepo.Health != nil {
		git.Health = GitHealth(*gitRepo.Health)
	}
	if gitRepo.AuthMethod != nil {
		git.AuthMethod = AuthMethod(*gitRepo.AuthMethod)
	}

	return git
}
