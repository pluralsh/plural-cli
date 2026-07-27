package bridge

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/config"
)

func TestLegacyProfileStorePersistsCredentialsWithOwnerOnlyPermissions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	conf := &config.Config{Email: "dev@example.com", Token: "secret"}
	if err := (LegacyProfileStore{}).Persist(t.Context(), conf); err != nil {
		t.Fatalf("Persist() error = %v", err)
	}
	info, err := os.Stat(filepath.Join(home, ".plural", config.ConfigName))
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("config permissions = %04o", got)
	}
}
