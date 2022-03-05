package git

import (
	"fmt"
	"github.com/pluralsh/plural/pkg/utils/errors"
)

func Sync(root, msg string, force bool) error {
	if res, err := git(root, "add", "."); err != nil {
		return errors.ErrorWrap(fmt.Errorf(res), "`git add .` failed")
	}

	if res, err := git(root, "commit", "-m", msg); err != nil {
		return errors.ErrorWrap(fmt.Errorf(res), "failed to commit changes")
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
		return errors.ErrorWrap(fmt.Errorf(res), fmt.Sprintf("`git push origin %s` failed", branch))
	}

	return nil
}