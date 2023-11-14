package backup

import (
	"fmt"
	"os"
	"path/filepath"

	apiclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	"github.com/pluralsh/plural-cli/pkg/config"
)

type CAPIBackup struct {
	dirPath string
}

func (this CAPIBackup) createDir() {
	if this.Exists() {
		return
	}

	_ = os.MkdirAll(this.dirPath, os.ModePerm)
}

func (this CAPIBackup) Exists() bool {
	_, err := os.Stat(this.dirPath)
	return !os.IsNotExist(err)
}

func (this CAPIBackup) Save(options apiclient.MoveOptions) error {
	client, err := apiclient.New("")
	if err != nil {
		return err
	}

	this.createDir()
	if len(options.FromKubeconfig.Context) == 0 || len(options.FromKubeconfig.Path) == 0 {
		return fmt.Errorf("both FromKubeconfig context and path have to be configured\n")
	}

	options.ToDirectory = this.dirPath
	options.Namespace = "bootstrap"

	return client.Move(options)
}

func (this CAPIBackup) Restore(options apiclient.MoveOptions) error {
	client, err := apiclient.New("")
	if err != nil {
		return err
	}

	if len(options.ToKubeconfig.Context) == 0 || len(options.ToKubeconfig.Path) == 0 {
		return fmt.Errorf("both ToKubeconfig context and path have to be configured\n")
	}

	if !this.Exists() {
		return fmt.Errorf("could not find move backup to restore from")
	}

	options.FromDirectory = this.dirPath
	options.Namespace = "bootstrap"

	return client.Move(options)
}

func (this CAPIBackup) Remove() error {
	if !this.Exists() {
		return nil
	}

	return os.RemoveAll(this.dirPath)
}

func NewCAPIBackup(cluster string) Backup[apiclient.MoveOptions] {
	path, _ := config.PluralDir()

	return CAPIBackup{
		dirPath: filepath.Join(path, backupsDir, cluster),
	}
}
