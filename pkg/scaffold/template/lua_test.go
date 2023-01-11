package template_test

import (
	"os"
	"path"
	"testing"

	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/scaffold/template"
	pluraltest "github.com/pluralsh/plural/pkg/test"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestFromLuaTemplateComplex(t *testing.T) {
	tests := []struct {
		name             string
		vals             map[string]interface{}
		script           string
		keyContent       string
		expectedResponse string
		expectedError    string
	}{
		{
			name:       `test complex azure`,
			keyContent: `key: "gKNJBnflqQA6lfUKLWMwl7CMJk4j+qqG9jnGYdTvwTk="`,
			vals: map[string]interface{}{
				"Values":        map[string]interface{}{"console_dns": "https://onplural.sh"},
				"Configuration": "",
				"License":       "abc",
				"Region":        "US",
				"Project":       "test",
				"Cluster":       "test",
				"Provider":      "azure",
				"Config":        map[string]interface{}{"Email": "test@plural.sh"},
				"Context":       map[string]interface{}{"SubscriptionId": "abc", "TenantId": "bca"},
				"console":       map[string]interface{}{"secrets": map[string]interface{}{"admin_password": "abc", "jwt": "abc", "admin_email": "", "erlang": "abc"}},
			},
			script: func() string {
				io, err := os.ReadFile("../../test/lua/values.yaml.lua")
				if err != nil {
					t.Fatal(err)
				}
				return string(io)
			}(),

			expectedResponse: `global:
  application:
    links:
    - description: console web ui
      url: https://onplural.sh
test:
  consoleIdentityClientId: '{{ .Import.Terraform.console_msi_client_id }}'
  consoleIdentityId: '{{ .Import.Terraform.console_msi_id }}'
  enabled: true
  extraEnv:
  - name: ARM_USE_MSI
    value: true
  - name: ARM_SUBSCRIPTION_ID
    value: abc
  - name: ARM_TENANT_ID
    value: bca
  ingress:
    annotations:
      kubernetes.io/tls-acme: "true"
    console_dns: https://onplural.sh
    ingressClass: nginx
  license: abc
  podLabels:
    aadpodidbinding: console
  provider: azure
  replicaCount: 2
  secrets:
    admin_email: ""
    admin_password: abc
    branch_name: master
    cluster_name: test
    config: |
      apiVersion: platform.plural.sh/v1alpha1
      kind: Config
      metadata: null
      spec:
        email: test@plural.sh
        token: abc
        namespacePrefix: test
        endpoint: http://example.com
        lockProfile: abc
        reportErrors: false
    erlang: abc
    git_access_token: ""
    git_email: console@plural.sh
    git_url: git@git.test.com:portfolio/space.space_name.git
    git_user: console
    id_rsa: ""
    id_rsa_pub: ""
    jwt: abc
    key: |
      key: "gKNJBnflqQA6lfUKLWMwl7CMJk4j+qqG9jnGYdTvwTk="
    repo_root: ""
    ssh_passphrase: ""
  serviceAccount:
    annotations:
      eks.amazonaws.com/role-arn: arn:aws:iam::test:role/test-console
    create: true
`,
		},
		{
			name:       `test complex aws`,
			keyContent: `key: "gKNJBnflqQA6lfUKLWMwl7CMJk4j+qqG9jnGYdTvwTk="`,
			vals: map[string]interface{}{
				"Values":        map[string]interface{}{"console_dns": "https://onplural.sh"},
				"Configuration": "",
				"License":       "abc",
				"Region":        "US",
				"Project":       "test",
				"Cluster":       "test",
				"Provider":      "aws",
				"Config":        map[string]interface{}{"Email": "test@plural.sh"},
				"Context":       map[string]interface{}{"SubscriptionId": "abc", "TenantId": "bca"},
				"console":       map[string]interface{}{"secrets": map[string]interface{}{"admin_password": "abc", "jwt": "abc", "admin_email": "", "erlang": "abc"}},
			},
			script: func() string {
				io, err := os.ReadFile("../../test/lua/values.yaml.lua")
				if err != nil {
					t.Fatal(err)
				}
				return string(io)
			}(),

			expectedResponse: `global:
  application:
    links:
    - description: console web ui
      url: https://onplural.sh
test:
  enabled: true
  ingress:
    annotations:
      kubernetes.io/tls-acme: "true"
    console_dns: https://onplural.sh
    ingressClass: nginx
  license: abc
  provider: aws
  replicaCount: 2
  secrets:
    admin_email: ""
    admin_password: abc
    branch_name: master
    cluster_name: test
    config: |
      apiVersion: platform.plural.sh/v1alpha1
      kind: Config
      metadata: null
      spec:
        email: test@plural.sh
        token: abc
        namespacePrefix: test
        endpoint: http://example.com
        lockProfile: abc
        reportErrors: false
    erlang: abc
    git_access_token: ""
    git_email: console@plural.sh
    git_url: git@git.test.com:portfolio/space.space_name.git
    git_user: console
    id_rsa: ""
    id_rsa_pub: ""
    jwt: abc
    key: |
      key: "gKNJBnflqQA6lfUKLWMwl7CMJk4j+qqG9jnGYdTvwTk="
    repo_root: ""
    ssh_passphrase: ""
  serviceAccount:
    annotations:
      eks.amazonaws.com/role-arn: arn:aws:iam::test:role/test-console
    create: true
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")

			err = os.Chdir(dir)
			assert.NoError(t, err)
			defaultConfig := pluraltest.GenDefaultConfig()
			err = defaultConfig.Save(config.ConfigName)
			assert.NoError(t, err)

			_, err = git.Init()
			assert.NoError(t, err)
			_, err = git.GitRaw("config", "--global", "user.email", "test@plural.com")
			assert.NoError(t, err)
			_, err = git.GitRaw("config", "--global", "user.name", "test")
			assert.NoError(t, err)
			_, err = git.GitRaw("add", "-A")
			assert.NoError(t, err)
			_, err = git.GitRaw("commit", "-m", "init")
			assert.NoError(t, err)
			_, err = git.GitRaw("remote", "add", "origin", "git@git.test.com:portfolio/space.space_name.git")
			assert.NoError(t, err)

			err = os.MkdirAll(path.Join(dir, ".plural"), os.ModePerm)
			assert.NoError(t, err)
			err = os.WriteFile(path.Join(dir, ".plural", "key"), []byte(test.keyContent), 0644)
			assert.NoError(t, err)

			globals := map[string]interface{}{}
			values := make(map[string]map[string]interface{})

			err = template.FromLuaTemplate(test.vals, globals, values, "test", test.script)
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

func TestFromLuaTemplate(t *testing.T) {
	tests := []struct {
		name             string
		vals             map[string]interface{}
		script           string
		expectedResponse string
		expectedError    string
	}{
		{
			name: `test globals`,
			vals: map[string]interface{}{
				"Values": map[string]interface{}{"console_dns": "https://onplural.sh"},
			},
			script: `output = {
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
			vals: map[string]interface{}{
				"Context": map[string]interface{}{"SubscriptionId": "abc", "TenantId": "cda"},
			},
			script: `output = {
		extraEnv={
			{
				name="ARM_USE_MSI",
				value = 'true'
	
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
			expectedResponse: `global: {}
test:
  enabled: true
  extraEnv:
  - name: ARM_USE_MSI
    value: "true"
  - name: ARM_SUBSCRIPTION_ID
    value: abc
  - name: ARM_TENANT_ID
    value: cda
`,
		},
		{
			name: `test annotations`,
			vals: map[string]interface{}{
				"Context": map[string]interface{}{"SubscriptionId": "abc", "TenantId": "cda"},
			},
			script: `output = {
					ingress={
						annotations={
							['kubernetes.io/tls-acme']='true',
						}
					}
}`,
			expectedResponse: `global: {}
test:
  enabled: true
  ingress:
    annotations:
      kubernetes.io/tls-acme: "true"
`,
		},
		{
			name: `test importValue`,
			vals: map[string]interface{}{
				"Values": map[string]interface{}{"console_dns": "https://onplural.sh"},
			},
			script: `output = {}
    output.consoleIdentityId=importValue("Terraform", "console_msi_id")
    output.consoleIdentityClientId=importValue("Terraform", "console_msi_client_id")
`,
			expectedResponse: `global: {}
test:
  consoleIdentityClientId: '{{ .Import.Terraform.console_msi_client_id }}'
  consoleIdentityId: '{{ .Import.Terraform.console_msi_id }}'
  enabled: true
`,
		},
		{
			name: `test when output doesn't exist`,
			vals: map[string]interface{}{
				"Values": map[string]interface{}{"console_dns": "https://onplural.sh"},
			},
			script: `
			somethingElse={}
`,
			expectedError: `the output variable is missing in the lua script`,
		},
		{
			name: `test encryption key`,
			vals: map[string]interface{}{},
			script: `
			output={
				secrets={
					key=''
                }
			}
			output.secrets.key=dedupe(Var, "console.secrets.key", readFile(pathJoin(homeDir(), ".plural", "key"))) .. "\n"
`,
			expectedResponse: `global: {}
test:
  enabled: true
  secrets:
    key: |
      key: 3fWQQvTe5w9Gu6uoJzw1C5SxnARtNT/QOvvISRFoxRs=
`,
		},
		{
			name: `test encryption key when already exists`,
			vals: map[string]interface{}{
				"console": map[string]interface{}{"secrets": map[string]interface{}{
					"key": "key: abc",
				}},
			},
			script: `
			output={
				secrets={
					key=''
                }
			}
			output.secrets.key=dedupe(Var, "console.secrets.key", readFile(pathJoin(homeDir(), ".plural", "key"))) .. "\n"
`,
			expectedResponse: `global: {}
test:
  enabled: true
  secrets:
    key: |
      key: abc
`,
		},
		{
			name: `test kebab case`,
			vals: map[string]interface{}{
				"Values": map[string]interface{}{
					"console_dns": "console.onplural.sh",
					"testBase": map[string]interface{}{
						"cypressEmail":    "test@plural.sh",
						"cypressPassword": "xyz",
					},
				},
			},
			script: `
			output={
				
			}
			output["test-base"]={
				enabled=true,
				secret={
					CYPRESS_EMAIL=Var.Values.testBase.cypressEmail,
					CYPRESS_PASSWORD=Var.Values.testBase.cypressPassword,
					CYPRESS_BASE_URL= "https://" .. Var.Values.console_dns .. "/",
				}
			}
`,
			expectedResponse: `global: {}
test:
  enabled: true
  test-base:
    enabled: true
    secret:
      CYPRESS_BASE_URL: https://console.onplural.sh/
      CYPRESS_EMAIL: test@plural.sh
      CYPRESS_PASSWORD: xyz
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")

			err = os.Chdir(dir)
			assert.NoError(t, err)
			err = os.MkdirAll(path.Join(dir, ".plural"), os.ModePerm)
			assert.NoError(t, err)
			err = os.WriteFile(path.Join(dir, ".plural", "key"), []byte("key: 3fWQQvTe5w9Gu6uoJzw1C5SxnARtNT/QOvvISRFoxRs="), 0644)
			assert.NoError(t, err)

			globals := map[string]interface{}{}
			values := make(map[string]map[string]interface{})
			err = template.FromLuaTemplate(test.vals, globals, values, "test", test.script)
			if test.expectedError != "" {
				assert.Equal(t, test.expectedError, err.Error())
			} else {
				assert.NoError(t, err)

				values["global"] = globals
				res, err := yaml.Marshal(values)
				assert.NoError(t, err)
				response := string(res)
				assert.Equal(t, test.expectedResponse, response)
			}
		})
	}
}
