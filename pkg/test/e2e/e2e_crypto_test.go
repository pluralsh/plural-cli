//go:build e2e

package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	utiltest "github.com/pluralsh/plural/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestKeyValidation(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	assert.NoError(t, err)

	testDir := path.Join(homeDir, "test")
	keyFingerprintFile := path.Join(testDir, ".keyid")
	err = os.Chdir(testDir)
	assert.NoError(t, err)

	// generate new validation file
	cmd := exec.Command("plural", "crypto", "fingerprint")
	cmdOutput, err := cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Equal(t, "", string(cmdOutput))
	utiltest.CheckFingerprint(t, keyFingerprintFile)

	// enforce a new key creation. Validation should fail
	err = os.Remove(path.Join(homeDir, ".plural", "key"))
	assert.NoError(t, err)
	cmd = exec.Command("plural", "crypto", "unlock")
	cmdOutput, err = cmd.CombinedOutput()
	assert.Error(t, err, "expected: 'the key fingerprint doesn't match' error")
	fmt.Println(string(cmdOutput))
}
