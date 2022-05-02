package crypto

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func backupKey(key string) error {
	p := getKeyPath()
	if utils.Exists(p) {
		aes, _ := Read(p)
		if aes.Key == key {
			return nil
		}

		ind := 0
		for {
			bp := backupPath(ind)
			if utils.Exists(bp) {
				ind++
				continue
			}

			utils.Highlight("===> backing up aes key to %s\n", bp)
			if err := os.MkdirAll(filepath.Dir(bp), os.ModePerm); err != nil {
				return err
			}
			return utils.CopyFile(p, bp)
		}
	}

	return nil
}

func backupPath(ind int) string {
	folder, _ := os.UserHomeDir()
	infix := ""
	if ind > 0 {
		infix = fmt.Sprintf("_%d.", ind)
	}

	return pathing.SanitizeFilepath(filepath.Join(folder, ".plural", "keybackups", fmt.Sprintf("key_backup%s", infix)))
}
