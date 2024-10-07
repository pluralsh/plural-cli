package git

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

func Root() (string, error) {
	return GitRaw("rev-parse", "--show-toplevel")
}

func Repo() (*gogit.Repository, error) {
	root, err := Root()
	if err != nil {
		return nil, err
	}

	return gogit.PlainOpen(root)
}

func PrintDiff() error {
	cmd := exec.Command("git", "--no-pager", "diff")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CurrentBranch() (string, error) {
	return GitRaw("rev-parse", "--abbrev-ref", "HEAD")
}

func Submodule(url string) error {
	_, err := GitRaw("submodule", "add", "-f", url)
	return err
}

func RemoveSubmodule(name string) error {
	_, err := GitRaw("rm", name)
	return err
}

func BranchedSubmodule(url, branch string) error {
	_, err := GitRaw("submodule", "add", "-f", "-b", branch, url)
	return err
}

func PathClone(url, branch, path string) error {
	_, err := GitRaw("clone", url, "-b", branch, path)
	return err
}

func Rm(path string) error {
	_, err := GitRaw("rm", path)
	return err
}

func HasUpstreamChanges() (bool, string, error) {
	headRef, err := GitRaw("rev-parse", "--symbolic-full-name", "HEAD")
	if err != nil {
		return false, "", err
	}

	headSha, err := GitRaw("rev-parse", "HEAD")
	if err != nil {
		return false, "", err
	}

	res, err := GitRaw("ls-remote", "origin", "-h", headRef)
	if err != nil {
		return false, "", err
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

	return remote == headSha, remote, nil
}

func Init() (string, error) {
	return GitRaw("init")
}

func GetURL() (string, error) {
	return GitRaw("ls-remote", "--get-url")
}
