package bundle

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/pluralsh/plural/pkg/api"
)

func fetchFunction(item *api.ConfigurationItem) (interface{}, error) {
	switch item.FunctionName {
	case "repoUrl":
		return repoUrl()
	case "repoRoot":
		return repoRoot()
	case "repoName":
		return repoName()
	case "branchName":
		return branchName()
	}

	return nil, fmt.Errorf("unsupported function %s, contact the application developer", item.FunctionName)
}

func repoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	res, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(res)), err
}

func repoName() (string, error) {
	root, err := repoRoot()
	return path.Base(root), err
}

func repoUrl() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	res, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(res)), err
}

func branchName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	res, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(res)), err
}
