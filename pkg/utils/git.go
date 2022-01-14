package utils

import (
	"os"
	"fmt"
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
		if len(cols) > 1 {
			result = append(result, cols[1])
		}
	}
	return result, nil
}

func Sync(msg string, force bool) error {
	root, _ := ProjectRoot()
	if res, err := git(root, "add", "."); err != nil {
		return ErrorWrap(fmt.Errorf(res), "`git add .` failed")
	}

	if res, err := git(root, "commit", "-m", msg); err != nil {
		return ErrorWrap(fmt.Errorf(res), "failed to commit changes")
	}

	branch, err := CurrentBranch()
	if err != nil {
		return err
	}

	args := []string{"push", "origin", branch}
	if force {
		args = []string{"push", "-f", "origin", branch}
	}

	if res, err := git(root, args...); err != nil {
		return ErrorWrap(fmt.Errorf(res), fmt.Sprintf("`git push origin %s` failed", branch))
	}

	return nil
}

func CurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	res, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(res)), err
}

func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	res, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(res)), err
}