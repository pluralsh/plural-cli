package access

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalAccessRepositoryImportsLegacyProfilesOnce(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".plural")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	legacy := "apiVersion: platform.plural.sh/v1alpha1\nkind: Config\nmetadata:\n  name: personal\nspec:\n  email: dev@example.com\n  endpoint: app.plural.sh\n  token: legacy-secret\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	credentials := &memoryCredentials{}
	repository := NewLocalRepository(home, credentials)
	state, err := repository.Load(t.Context())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(state.Profiles) != 1 || state.ActiveProfileID == "" {
		t.Fatalf("state = %#v", state)
	}
	if credentials.values[state.ActiveProfileID] != "legacy-secret" {
		t.Fatal("legacy secret was not moved to credential storage")
	}
	contents, err := os.ReadFile(filepath.Join(dir, accessRegistryName))
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) == "" || containsSecret(string(contents), "legacy-secret") {
		t.Fatalf("registry contains secret:\n%s", contents)
	}
	if err := os.Remove(filepath.Join(dir, "config.yml")); err != nil {
		t.Fatal(err)
	}
	state, err = repository.Load(t.Context())
	if err != nil || len(state.Profiles) != 1 {
		t.Fatalf("second Load() = %#v, %v", state, err)
	}
}

func containsSecret(value, secret string) bool {
	for i := 0; i+len(secret) <= len(value); i++ {
		if value[i:i+len(secret)] == secret {
			return true
		}
	}
	return false
}
