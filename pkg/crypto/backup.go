package crypto

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
)

func BackupKey(client api.Client) error {
	p := getKeyPath()
	aes, err := Read(p)
	if err != nil {
		return err
	}

	host, _ := os.Hostname()
	name, err := utils.ReadLineDefault("Give your key backup a name", host)
	if err != nil {
		return err
	}

	repos := []string{}
	if repo, err := git.GetURL(); err == nil {
		repos = append(repos, repo)
	}

	utils.Highlight("===> backing up aes key to plural\n")
	return client.CreateKeyBackup(api.KeyBackupAttributes{
		Name:         name,
		Repositories: repos,
		Key:          aes.Key,
	})
}

func DownloadBackup(client api.Client, name string) error {
	backup, err := client.GetKeyBackup(name)
	if err != nil {
		return api.GetErrorResponse(err, "GetKeyBackup")
	}

	if backup == nil {
		return fmt.Errorf("no backup found for %s", name)
	}

	return Setup(backup.Value)
}

func backupKey(key string) error {
	p := getKeyPath()
	if utils.Exists(p) {
		aes, err := Read(p)
		if err != nil {
			return err
		}
		if aes != nil && aes.Key == key {
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
