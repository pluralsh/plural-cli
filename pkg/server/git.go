package server

import (
	"os"
	"path/filepath"
	"github.com/pluralsh/plural/pkg/utils"
	homedir "github.com/mitchellh/go-homedir"
)

func setupGit(setup *SetupRequest) error {
	p, err := homedir.Expand("~/.ssh")
	if err != nil {
		return err
	}

	utils.WriteFileIfNotPresent(filepath.Join(p, "id_rsa"), setup.SshPrivateKey)
	utils.WriteFileIfNotPresent(filepath.Join(p, "id_rsa.pub"), setup.SshPublicKey)

	if err := execCmd("ssh-add", filepath.Join(p, "id_rsa")); err != nil {
		return err
	}


	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := execCmd("git", "clone", setup.GitUrl, "workspace"); err != nil {
		return err
	}

	os.Chdir(filepath.Join(dir, "workspace"))
	if err := execCmd("plural", "crypto", "init"); err != nil {
		return err
	}

	return execCmd("plural", "crypto", "unlock")
}