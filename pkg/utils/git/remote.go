package git

import (
	e "errors"
	"fmt"
	"os"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/pluralsh/plural-cli/pkg/utils/errors"
)

func Clone(auth transport.AuthMethod, url, path string) (*gogit.Repository, error) {
	return gogit.PlainClone(path, false, &gogit.CloneOptions{
		Auth:     auth,
		URL:      url,
		Progress: os.Stdout,
	})
}

func Sync(root, msg string, force bool) error {
	if res, err := git(root, "add", "."); err != nil {
		return errors.ErrorWrap(e.New(res), "`git add .` failed")
	}

	if res, err := git(root, "commit", "-m", msg); err != nil {
		return errors.ErrorWrap(e.New(res), "failed to commit changes")
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
		return errors.ErrorWrap(e.New(res), fmt.Sprintf("`git push origin %s` failed", branch))
	}

	return nil
}
