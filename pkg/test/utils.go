package test

import (
	"os"
	"strings"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func GenDefaultConfig() config.Config {
	return config.Config{
		Email:           "test@plural.sh",
		Token:           "abc",
		NamespacePrefix: "test",
		Endpoint:        "http://example.com",
		LockProfile:     "abc",
		ReportErrors:    false,
	}
}

func CheckFingerprint(t *testing.T, path string) {
	b, err := os.ReadFile(path)
	assert.NoError(t, err)
	keyID := string(b)
	if keyID == "" {
		t.Fatal("expected not empty file")
	}
	if !strings.HasPrefix(keyID, "keyid: SHA256:") {
		t.Fatalf("expected SHA256 format, got %s", keyID)
	}
	aesKey, err := crypto.Materialize()
	assert.NoError(t, err)
	var k crypto.KeyValidator
	err = yaml.Unmarshal([]byte(keyID), &k)
	assert.NoError(t, err)
	assert.Equal(t, aesKey.ID(), k.KeyID)
}
