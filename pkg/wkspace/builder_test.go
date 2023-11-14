package wkspace

import (
	"testing"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestRequiredCliVersion(t *testing.T) {
	w := &Workspace{
		Charts: []*api.ChartInstallation{
			{
				Version: &api.Version{
					Dependencies: &api.Dependencies{
						CliVsn: "0.1.0",
					},
				},
			},
		},
		Terraform: []*api.TerraformInstallation{
			{
				Version: &api.Version{
					Dependencies: &api.Dependencies{
						CliVsn: "0.2.0",
					},
				},
			},
		},
	}
	vsn, ok := w.RequiredCliVsn()
	assert.True(t, ok)
	assert.Equal(t, "v0.2.0", vsn)
}

func TestRequiredCliVersionEmpty(t *testing.T) {
	w := &Workspace{
		Charts: []*api.ChartInstallation{
			{
				Version: &api.Version{
					Dependencies: &api.Dependencies{
						CliVsn: "bogus",
					},
				},
			},
		},
		Terraform: []*api.TerraformInstallation{
			{
				Version: &api.Version{
					Dependencies: &api.Dependencies{
						CliVsn: "",
					},
				},
			},
		},
	}
	_, ok := w.RequiredCliVsn()
	assert.False(t, ok)
}
