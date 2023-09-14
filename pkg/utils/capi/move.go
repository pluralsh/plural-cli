package capi

import (
	"fmt"
	"os"
	"path/filepath"

	apiclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

var (
	clusterStateDir = ""
)

func init() {
	gitRootDir, err := git.Root()
	if err != nil {
		panic(err)
	}

	clusterStateDir = pathing.SanitizeFilepath(filepath.Join(gitRootDir, "bootstrap", "capi"))
}

func MoveBackupExists() bool {
	_, err := os.Stat(clusterStateDir)
	return !os.IsNotExist(err)
}

func SaveMoveBackup(options apiclient.MoveOptions) error {
	client, err := apiclient.New("")
	if err != nil {
		return err
	}

	if !MoveBackupExists() {
		err = os.Mkdir(clusterStateDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if len(options.FromKubeconfig.Context) == 0 || len(options.FromKubeconfig.Path) == 0 {
		return fmt.Errorf("both FromKubeconfig context and path have to be configured\n")
	}

	options.ToDirectory = clusterStateDir
	options.Namespace = "bootstrap"

	return client.Move(options)
}

func RestoreMoveBackup(options apiclient.MoveOptions) error {
	client, err := apiclient.New("")
	if err != nil {
		return err
	}

	if !MoveBackupExists() {
		return fmt.Errorf("could not find move backup to restore from")
	}

	if len(options.ToKubeconfig.Context) == 0 || len(options.ToKubeconfig.Path) == 0 {
		return fmt.Errorf("both ToKubeconfig context and path have to be configured\n")
	}

	options.FromDirectory = clusterStateDir
	options.Namespace = "bootstrap"

	return client.Move(options)
}

func RemoveStateBackup() error {
	if !MoveBackupExists() {
		return nil
	}

	return os.Remove(clusterStateDir)
}
