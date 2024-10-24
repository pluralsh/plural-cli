package pr_test

import (
	"os"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/pr"
	"github.com/stretchr/testify/assert"
)

func TestBuildCRD(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		envs        map[string]string
		expectedCtx map[string]interface{}
	}{
		{
			name: "test PR automation",
			path: "../../test/prautomation/prautomations.yaml",
			envs: map[string]string{"PLURAL__NAME": "test", "PLURAL__REGION": "eu-central-1", "PLURAL__TYPE": "s3"},
			expectedCtx: map[string]interface{}{
				"name": "test", "region": "eu-central-1", "type": "s3",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.envs {
				os.Setenv(k, v)
			}

			prTemplate, err := pr.BuildCRD(test.path, "")
			assert.NoError(t, err)
			assert.Equal(t, test.expectedCtx, prTemplate.Context)

		})
	}
}
