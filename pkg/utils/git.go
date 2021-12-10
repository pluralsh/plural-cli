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
			root, _ = RepoRoot()
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

func ChangedFiles() ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(res)), "\n") {
		cols := strings.Split(strings.TrimSpace(line), " ")
		result = append(result, cols[1])
	}
	return result, nil
}

func Sync(msg string) error {
	root, _ := ProjectRoot()
	if err := git(root, "add", "."); err != nil {
		return err
	}

	if err := git(root, "commit", "-m", msg); err != nil {
		return err
	}

	return git(root, "push")
}

func git(root string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	return cmd.Run()
}