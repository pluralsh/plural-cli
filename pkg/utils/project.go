package utils

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
)

func ProjectRoot() (root string, found bool) {
	root, _ = os.Getwd()
	found = false

	for {
		if runtime.GOOS == "windows" {
			if root == "C:\\" {
				root, _ = git.Root()
				break
			}
		} else {
			if root == "/" {
				root, _ = git.Root()
				break
			}
		}

		if Exists(pathing.SanitizeFilepath(filepath.Join(root, "workspace.yaml"))) {
			found = true
			return
		}

		root = filepath.Dir(root)
	}

	return
}
