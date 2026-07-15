package console_test

import (
	"testing"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAgentHelmValuesWithoutTemplating(t *testing.T) {
	t.Parallel()

	settings := &gqlclient.DeploymentSettingsFragment{
		AgentHelmValues: lo.ToPtr("foo: bar\n"),
	}

	vals, err := console.ResolveAgentHelmValues(settings, nil)
	require.NoError(t, err)
	assert.Equal(t, "bar", vals["foo"])
}

func TestResolveAgentHelmValuesWithTemplating(t *testing.T) {
	t.Parallel()

	settings := &gqlclient.DeploymentSettingsFragment{
		AgentHelmValues:             lo.ToPtr("cluster:\n  name: {{ cluster.name }}\n"),
		AgentHelmValuesTemplateable: lo.ToPtr(true),
	}
	cluster := &gqlclient.ClusterFragment{
		ID:   "cluster-id",
		Name: "test",
	}

	vals, err := console.ResolveAgentHelmValues(settings, cluster)
	require.NoError(t, err)

	clusterVals, ok := vals["cluster"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test", clusterVals["name"])
}

func TestResolveAgentHelmValuesRequiresClusterWhenTemplating(t *testing.T) {
	t.Parallel()

	settings := &gqlclient.DeploymentSettingsFragment{
		AgentHelmValues:             lo.ToPtr("cluster:\n  id: {{ cluster.id }}\n"),
		AgentHelmValuesTemplateable: lo.ToPtr(true),
	}

	_, err := console.ResolveAgentHelmValues(settings, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster context is required")
}

func TestResolveAgentHelmValuesRendersClusterMetadataAndTags(t *testing.T) {
	t.Parallel()

	settings := &gqlclient.DeploymentSettingsFragment{
		AgentHelmValues: lo.ToPtr(`name: {{ cluster.name }}
region: {{ cluster.metadata.region }}
env: {{ cluster.tags.env }}
id: {{ cluster.id }}
kas: {{ cluster.kasUrl }}
`),
		AgentHelmValuesTemplateable: lo.ToPtr(true),
	}
	cluster := &gqlclient.ClusterFragment{
		ID:     "cluster-1",
		Name:   "prod-cluster",
		KasURL: lo.ToPtr("https://kas.example.com"),
		Metadata: map[string]any{
			"region": "eu-central-1",
		},
		Tags: []*gqlclient.ClusterTags{
			{Name: "env", Value: "production"},
		},
	}

	vals, err := console.ResolveAgentHelmValues(settings, cluster)
	require.NoError(t, err)
	assert.Equal(t, "prod-cluster", vals["name"])
	assert.Equal(t, "eu-central-1", vals["region"])
	assert.Equal(t, "production", vals["env"])
	assert.Equal(t, "cluster-1", vals["id"])
	assert.Equal(t, "https://kas.example.com", vals["kas"])
}

func TestResolveAgentHelmValuesNilSettings(t *testing.T) {
	t.Parallel()

	vals, err := console.ResolveAgentHelmValues(nil, nil)
	require.NoError(t, err)
	assert.Empty(t, vals)
}
