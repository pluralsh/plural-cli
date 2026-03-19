package cd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteLuaTemplate_EmptyScript(t *testing.T) {
	dir := t.TempDir()

	result, err := executeLuaTemplate("", dir, nil)

	require.NoError(t, err)
	assert.Empty(t, result["valuesFiles"])
}

func TestExecuteLuaTemplate_SetsValues(t *testing.T) {
	dir := t.TempDir()
	script := `
values["replicas"] = 3
values["image"] = "nginx"
`
	result, err := executeLuaTemplate(script, dir, nil)

	require.NoError(t, err)
	values, ok := result["values"].(map[string]interface{})
	require.True(t, ok, "values should be map[string]interface{}")
	assert.Equal(t, float64(3), values["replicas"])
	assert.Equal(t, "nginx", values["image"])
	assert.Empty(t, result["valuesFiles"])
}

func TestExecuteLuaTemplate_SetsValuesFiles(t *testing.T) {
	dir := t.TempDir()
	script := `
valuesFiles[1] = "base.yaml"
valuesFiles[2] = "override.yaml"
`
	result, err := executeLuaTemplate(script, dir, nil)

	require.NoError(t, err)
	valuesFiles, ok := result["valuesFiles"].([]string)
	require.True(t, ok, "valuesFiles should be []string")
	assert.Equal(t, []string{"base.yaml", "override.yaml"}, valuesFiles)
}

func TestExecuteLuaTemplate_SyntaxError(t *testing.T) {
	dir := t.TempDir()
	script := `this is not valid lua %%%`

	_, err := executeLuaTemplate(script, dir, nil)

	assert.Error(t, err)
}

func TestExecuteLuaTemplate_RuntimeError(t *testing.T) {
	dir := t.TempDir()
	script := `
local x = nil
x.field = "boom"
`
	_, err := executeLuaTemplate(script, dir, nil)

	assert.Error(t, err)
}

func TestExecuteLuaTemplate_NestedValues(t *testing.T) {
	dir := t.TempDir()
	script := `
values["db"] = { host = "localhost", port = 5432 }
`
	result, err := executeLuaTemplate(script, dir, nil)

	require.NoError(t, err)
	values, ok := result["values"].(map[string]interface{})
	require.True(t, ok)
	db, ok := values["db"].(map[string]interface{})
	require.True(t, ok, "db should be a nested map")
	assert.Equal(t, "localhost", db["host"])
	assert.Equal(t, float64(5432), db["port"])
}

func TestExecuteLuaTemplate_UsesTempDir(t *testing.T) {
	dir := t.TempDir()
	helperContent := `
function greet(name)
  return "hello " .. name
end
`
	helperPath := filepath.Join(dir, "helper.lua")
	err := os.WriteFile(helperPath, []byte(helperContent), 0600)
	require.NoError(t, err)

	// Use dofile with absolute path; require resolves relative to CWD, not dir.
	script := "dofile('" + helperPath + "')\n" + `
values["greeting"] = greet("world")
`
	result, err := executeLuaTemplate(script, dir, nil)

	require.NoError(t, err)
	values, ok := result["values"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "hello world", values["greeting"])
}

// Binding tests – each test mirrors what the real luaBindings helpers produce,
// then asserts that Lua can read those values as expected.

func TestExecuteLuaTemplate_ConfigurationBinding(t *testing.T) {
	dir := t.TempDir()
	// Mirrors luaConfigurationBinding: map[string]string keyed by config name.
	script := `
values["env"]    = configuration["env"]
values["region"] = configuration["region"]
`
	bindings := map[string]interface{}{
		"configuration": map[string]string{
			"env":    "production",
			"region": "us-east-1",
		},
	}

	result, err := executeLuaTemplate(script, dir, bindings)

	require.NoError(t, err)
	values := result["values"].(map[string]interface{})
	assert.Equal(t, "production", values["env"])
	assert.Equal(t, "us-east-1", values["region"])
}

func TestExecuteLuaTemplate_ServiceBinding(t *testing.T) {
	dir := t.TempDir()
	// Mirrors luaServiceBinding: both PascalCase and lowercase keys are present.
	script := `
values["name"]      = service["name"]
values["namespace"] = service["namespace"]
values["Name"]      = service["Name"]
`
	bindings := map[string]interface{}{
		"service": map[string]interface{}{
			"name":      "my-service",
			"Name":      "my-service",
			"namespace": "default",
			"Namespace": "default",
		},
	}

	result, err := executeLuaTemplate(script, dir, bindings)

	require.NoError(t, err)
	values := result["values"].(map[string]interface{})
	assert.Equal(t, "my-service", values["name"])
	assert.Equal(t, "default", values["namespace"])
	assert.Equal(t, "my-service", values["Name"])
}

func TestExecuteLuaTemplate_ClusterBinding(t *testing.T) {
	dir := t.TempDir()
	// Mirrors luaClusterBinding: both PascalCase and lowercase keys, plus tags sub-map.
	script := `
values["clusterName"]    = cluster["name"]
values["clusterHandle"]  = cluster["handle"]
values["tagEnv"]         = cluster["tags"]["env"]
`
	bindings := map[string]interface{}{
		"cluster": map[string]interface{}{
			"name":   "prod-cluster",
			"Name":   "prod-cluster",
			"handle": "prod",
			"Handle": "prod",
			"tags": map[string]string{
				"env": "production",
			},
			"Tags": map[string]string{
				"env": "production",
			},
		},
	}

	result, err := executeLuaTemplate(script, dir, bindings)

	require.NoError(t, err)
	values := result["values"].(map[string]interface{})
	assert.Equal(t, "prod-cluster", values["clusterName"])
	assert.Equal(t, "prod", values["clusterHandle"])
	assert.Equal(t, "production", values["tagEnv"])
}

func TestExecuteLuaTemplate_ContextsBinding(t *testing.T) {
	dir := t.TempDir()
	// Mirrors luaContextsBinding: map[contextName]map[string]interface{}.
	script := `
values["dbHost"] = contexts["db-context"]["host"]
values["dbPort"] = contexts["db-context"]["port"]
`
	bindings := map[string]interface{}{
		"contexts": map[string]map[string]interface{}{
			"db-context": {
				"host": "db.internal",
				"port": 5432,
			},
		},
	}

	result, err := executeLuaTemplate(script, dir, bindings)

	require.NoError(t, err)
	values := result["values"].(map[string]interface{})
	assert.Equal(t, "db.internal", values["dbHost"])
	assert.Equal(t, float64(5432), values["dbPort"])
}

func TestExecuteLuaTemplate_ImportsBinding(t *testing.T) {
	dir := t.TempDir()
	// Mirrors luaImportsBinding: map[stackName]map[outputName]string.
	script := `
values["vpcId"]  = imports["network-stack"]["vpc_id"]
values["subnetId"] = imports["network-stack"]["subnet_id"]
`
	bindings := map[string]interface{}{
		"imports": map[string]map[string]string{
			"network-stack": {
				"vpc_id":    "vpc-abc123",
				"subnet_id": "subnet-def456",
			},
		},
	}

	result, err := executeLuaTemplate(script, dir, bindings)

	require.NoError(t, err)
	values := result["values"].(map[string]interface{})
	assert.Equal(t, "vpc-abc123", values["vpcId"])
	assert.Equal(t, "subnet-def456", values["subnetId"])
}

func TestExecuteLuaTemplate_MultipleBindingsUsedTogether(t *testing.T) {
	dir := t.TempDir()
	// Verifies that all binding types are available in a single script execution.
	script := `
values["svcName"]   = service["name"]
values["cfgEnv"]    = configuration["env"]
values["cluster"]   = cluster["name"]
values["ctxHost"]   = contexts["infra"]["endpoint"]
values["importOut"] = imports["infra-stack"]["bucket"]
`
	bindings := map[string]interface{}{
		"service":       map[string]interface{}{"name": "api", "Name": "api"},
		"configuration": map[string]string{"env": "staging"},
		"cluster":       map[string]interface{}{"name": "staging-cluster", "Name": "staging-cluster"},
		"contexts": map[string]map[string]interface{}{
			"infra": {"endpoint": "https://infra.internal"},
		},
		"imports": map[string]map[string]string{
			"infra-stack": {"bucket": "my-bucket"},
		},
	}

	result, err := executeLuaTemplate(script, dir, bindings)

	require.NoError(t, err)
	values := result["values"].(map[string]interface{})
	assert.Equal(t, "api", values["svcName"])
	assert.Equal(t, "staging", values["cfgEnv"])
	assert.Equal(t, "staging-cluster", values["cluster"])
	assert.Equal(t, "https://infra.internal", values["ctxHost"])
	assert.Equal(t, "my-bucket", values["importOut"])
}

func TestExecuteLuaTemplate_MissingBindingKeyIsNil(t *testing.T) {
	dir := t.TempDir()
	// Accessing a missing key in a binding map returns nil in Lua, not an error.
	script := `
if configuration["missing"] == nil then
  values["result"] = "nil as expected"
end
`
	bindings := map[string]interface{}{
		"configuration": map[string]string{"existing": "value"},
	}

	result, err := executeLuaTemplate(script, dir, bindings)

	require.NoError(t, err)
	values := result["values"].(map[string]interface{})
	assert.Equal(t, "nil as expected", values["result"])
}
