package git

import (
	"bufio"
	"fmt"
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

func Root() (string, error) {
	return gitRaw("rev-parse", "--show-toplevel")
}

func Repo() (*gogit.Repository, error) {
	root, err := Root()
	if err != nil {
		return nil, err
	}

	return gogit.PlainOpen(root)
}

func Worktree() (*gogit.Worktree, error) {
	repo, err := Repo()
	if err != nil {
		return nil, err
	}

	return repo.Worktree()
}

func CurrentBranch() (b string, err error) {
	repo, err := Repo()
	if err != nil {
		return
	}

	ref, err := repo.Head()
	if err != nil {
		return
	}

	b = ref.Name().Short()
	return
}

func HasUpstreamChanges() (bool, error) {
	repo, err := Repo()
	if err != nil {
		return false, err
	}

	ref, err := repo.Head()
	if err != nil {
		return false, err
	}

	res, err := gitRaw("ls-remote", "origin", "-h", fmt.Sprintf("refs/heads/%s", ref.Name().Short()))
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

	return remote == ref.Hash().String(), nil
}
