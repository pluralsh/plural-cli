//go:build e2e

package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/pluralsh/polly/containers"

	"github.com/stretchr/testify/assert"
)

func TestApiListInstallations(t *testing.T) {

	cmd := exec.Command("plural", "api", "list", "installations")
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	installations := make([]string, 0)
	rows := strings.Split(string(cmdOutput[:]), "\n")
	for _, row := range rows[1:] { // Skip the heading row and iterate through the remaining rows
		row = strings.ReplaceAll(row, "|", "")
		cols := strings.Fields(row) // Extract each column from the row.
		if len(cols) == 3 {
			installations = append(installations, cols[0])
		}
	}
	expected := []string{"bootstrap"}
	expects := containers.ToSet(expected)
	assert.True(t, expects.Equal(containers.ToSet(installations)), fmt.Sprintf("the expected %s is different then %s", expected, installations))
}

func TestPackagesList(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	assert.NoError(t, err)

	testDir := path.Join(homeDir, "test")

	err = os.Chdir(testDir)
	assert.NoError(t, err)
	cmd := exec.Command("plural", "packages", "list", "bootstrap")
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	packages := make([]string, 0)
	rows := strings.Split(string(cmdOutput[:]), "\n")
	for _, row := range rows[3:] { // Skip the heading row and iterate through the remaining rows
		row = strings.ReplaceAll(row, "|", "")
		cols := strings.Fields(row) // Extract each column from the row.
		if len(cols) == 3 {
			packages = append(packages, cols[1])
		}
	}
	expected := []string{"bootstrap", "kind-bootstrap-cluster-api", "cluster-api-control-plane", "cluster-api-bootstrap", "plural-certmanager-webhook", "cluster-api-cluster", "cluster-api-core", "cluster-api-provider-docker"}
	expects := containers.ToSet(expected)
	assert.True(t, expects.Equal(containers.ToSet(packages)), fmt.Sprintf("the expected %s is different then %s", expected, packages))
}
