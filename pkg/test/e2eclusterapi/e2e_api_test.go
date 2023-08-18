//go:build e2e

package e2e_test

import (
	"fmt"
	"os/exec"
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
