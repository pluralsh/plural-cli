package utils

import (
	"os"
	"fmt"
	"bufio"
	"path/filepath"
	"os/exec"
	"strings"
	"regexp"
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
	return gitRaw("rev-parse", "--show-toplevel")
}

func ChangedFiles() ([]string, error) {
	res, err := gitRaw("status", "--porcelain")
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

func IsSha(str string) bool {
	matches, _ := regexp.MatchString("[a-f0-9]{40}", str)
	return matches
}

func RemoteDiff() (bool, error) {
	branch, err := CurrentBranch()
	if err != nil {
		return false, err
	}

	res, err := gitRaw("ls-remote", "origin", "-h", fmt.Sprintf("refs/heads/%s", branch))
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(strings.NewReader(res))
	var remote string
	for scanner.Scan() {
		line := scanner.Text()
		remote = strings.Fields(line)[0]
		if IsSha(remote) {
			break
		}
	}

	current, err := CurrentSha(branch)
	if err != nil {
		return false, err
	}

	return remote == current, nil
}

func CurrentSha(branch string) (string, error) {
	return gitRaw("rev-list", "--max-count=1", fmt.Sprintf("origin/%s", branch))
}

func CurrentBranch() (string, error) {
	return gitRaw("rev-parse", "--abbrev-ref", "HEAD")
}

func RepoName(url string) string {
	reg, _ := regexp.Compile(".*/")
	base := reg.ReplaceAllString(url, "")

	return strings.TrimSuffix(base, ".git")
}

func gitRaw(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	res, err := ExecuteWithOutput(cmd)
	return strings.TrimSpace(string(res)), err
}

func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	res, err := ExecuteWithOutput(cmd)
	return strings.TrimSpace(string(res)), err
}