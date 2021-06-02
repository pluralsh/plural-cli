package utils

import (
	"os"
	"path/filepath"
	"os/exec"
	"strings"
)

func ProjectRoot() (root string, found bool) {
	root, _ = os.Getwd()
	found = false

	for {
		if root == "/" {
			root, _ := RepoRoot()
			break
		}

		if Exists(filepath.Join(root, "workspace.yaml")) {
			found = true
			return
		}

		root = filepath.Dir(root)
	}

	return
}

func RepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(res)), nil
}
