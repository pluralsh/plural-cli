package scaffold_test

import (
	"os"
	"testing"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/scaffold"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestFromLuaTemplate(t *testing.T) {
	tests := []struct {
		name             string
		vals             map[string]interface{}
		w                *wkspace.Workspace
		chartInst        *api.ChartInstallation
		expectedResponse string
		expectedError    string
	}{
		{
			name: `test globals`,
			w:    &wkspace.Workspace{},
			vals: map[string]interface{}{
				"Values": map[string]interface{}{"console_dns": "https://onplural.sh"},
			},
			chartInst: &api.ChartInstallation{
				Chart: &api.Chart{
					Name: "test",
				},
				Version: &api.Version{
					ValuesTemplate: `valuesYaml = {
    global={
        application={
            links={
                {	description= "console web ui",
                     url=Var.Values.console_dns
                }
            }
        }
    }
}`,
				},
			},
			expectedResponse: `global:
  application:
    links:
    - description: console web ui
      url: https://onplural.sh
test:
  enabled: true
`,
		},
		{
			name: `test env var`,
			w:    &wkspace.Workspace{},
			vals: map[string]interface{}{
				"Context": map[string]interface{}{"SubscriptionId": "abc", "TenantId": "cda"},
			},
			chartInst: &api.ChartInstallation{
				Chart: &api.Chart{
					Name: "test",
				},
				Version: &api.Version{
					ValuesTemplate: `valuesYaml = {
		extraEnv={
			{
				name="ARM_USE_MSI",
				value = true
	
			},
			{
				name="ARM_SUBSCRIPTION_ID",
				value=Var.Context.SubscriptionId
			},
			{
				name="ARM_TENANT_ID",
				value= Var.Context.TenantId
			}
    	}
}`,
				},
			},
			expectedResponse: `global: {}
test:
  enabled: true
  extraEnv:
  - name: ARM_USE_MSI
    value: true
  - name: ARM_SUBSCRIPTION_ID
    value: abc
  - name: ARM_TENANT_ID
    value: cda
`,
		},
		{
			name: `test annotations`,
			w:    &wkspace.Workspace{},
			vals: map[string]interface{}{
				"Context": map[string]interface{}{"SubscriptionId": "abc", "TenantId": "cda"},
			},
			chartInst: &api.ChartInstallation{
				Chart: &api.Chart{
					Name: "test",
				},
				Version: &api.Version{
					ValuesTemplate: `valuesYaml = {
					ingress={
						annotations={
							"kubernetes.io/tls-acme: 'true'",
							"cert-manager.io/cluster-issuer: letsencrypt-prod",
							"nginx.ingress.kubernetes.io/affinity: cookie",
							"nginx.ingress.kubernetes.io/force-ssl-redirect: 'true'",
							"nginx.ingress.kubernetes.io/proxy-read-timeout: '3600'",
							"nginx.ingress.kubernetes.io/proxy-send-timeout: '3600'",
							"nginx.ingress.kubernetes.io/session-cookie-path: /socket",
						}
					}
}`,
				},
			},
			expectedResponse: `global: {}
test:
  enabled: true
  ingress:
    annotations:
    - kubernetes.io/tls-acme: "true"
    - cert-manager.io/cluster-issuer: letsencrypt-prod
    - nginx.ingress.kubernetes.io/affinity: cookie
    - nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    - nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    - nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
    - nginx.ingress.kubernetes.io/session-cookie-path: /socket
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			globals := map[string]interface{}{}
			values := make(map[string]map[string]interface{})
			err := scaffold.FromLuaTemplate(test.vals, globals, values, test.w, test.chartInst)
			if test.expectedError != "" {
				assert.Equal(t, test.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
			}
			values["global"] = globals
			res, err := yaml.Marshal(values)
			assert.NoError(t, err)
			response := string(res)
			assert.Equal(t, test.expectedResponse, response)
		})
	}
}

func TestFromLuaTemplateComplex(t *testing.T) {
	tests := []struct {
		name             string
		vals             map[string]interface{}
		w                *wkspace.Workspace
		chartInst        *api.ChartInstallation
		expectedResponse string
		expectedError    string
	}{
		{
			name: `test complex`,
			w:    &wkspace.Workspace{},
			vals: map[string]interface{}{
				"Values":        map[string]interface{}{"console_dns": "https://onplural.sh"},
				"Configuration": "",
				"License":       "abc",
				"OIDC":          "",
				"Region":        "US",
				"Project":       "test",
				"Cluster":       "test",
				"Config":        "",
				"Provider":      "azure",
				"Context":       "",
				"Network":       "",
				"Applications":  "",
			},
			chartInst: &api.ChartInstallation{
				Chart: &api.Chart{
					Name: "test",
				},
				Version: &api.Version{
					ValuesTemplate: func() string {
						io, err := os.ReadFile("../test/lua/values.yaml.lua")
						if err != nil {
							t.Fatal(err)
						}
						return string(io)
					}(),
				},
			},
			expectedResponse: `
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			globals := map[string]interface{}{}
			values := make(map[string]map[string]interface{})

			err := scaffold.FromLuaTemplate(test.vals, globals, values, test.w, test.chartInst)
			if test.expectedError != "" {
				assert.Equal(t, test.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
			}
			values["global"] = globals
			res, err := yaml.Marshal(values)
			assert.NoError(t, err)
			response := string(res)
			assert.Equal(t, test.expectedResponse, response)
		})
	}
}
