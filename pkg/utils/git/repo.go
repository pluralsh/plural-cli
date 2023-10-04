package git

import (
	"bufio"
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

func CurrentBranch() (string, error) {
	return GitRaw("rev-parse", "--abbrev-ref", "HEAD")
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
