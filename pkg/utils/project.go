package utils

import (
	"os"
	"runtime"
	"path/filepath"
	"github.com/pluralsh/plural/pkg/utils/git"
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

		if Exists(filepath.Join(root, "workspace.yaml")) {
			found = true
			return
		}

		root = filepath.Dir(root)
	}

	return
}
