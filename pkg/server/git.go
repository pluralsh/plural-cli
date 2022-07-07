package server

import (
	"io/ioutil"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func gitExists() (bool, error) {
	dir, err := homedir.Expand("~/workspace")
	if err != nil {
		return false, err
	}

	return utils.Exists(dir), nil
}

func setupGit(setup *SetupRequest) error {
	p, err := homedir.Expand("~/.ssh")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(p, 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(pathing.SanitizeFilepath(filepath.Join(p, "id_rsa")), []byte(setup.SshPrivateKey), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(pathing.SanitizeFilepath(filepath.Join(p, "id_rsa.pub")), []byte(setup.SshPublicKey), 0644); err != nil {
		return err
	}

	if err := execCmd("ssh-add", pathing.SanitizeFilepath(filepath.Join(p, "id_rsa"))); err != nil {
		return err
	}

	dir, err := homedir.Expand("~/workspace")
	if err != nil {
		return err
	}

	if err := execCmd("git", "clone", setup.GitUrl, dir); err != nil {
		return err
	}

	if err := os.Chdir(dir); err != nil {
		return err
	}
	if err := gitConfig("user.email", setup.User.Email); err != nil {
		return err
	}

	name := "plural-shell"
	if setup.User.GitUser != "" {
		name = setup.User.GitUser
	}
	if err := gitConfig("user.name", name); err != nil {
		return err
	}

	if err := execCmd("plural", "crypto", "init"); err != nil {
		return err
	}

	return execCmd("plural", "crypto", "unlock")
}

func gitConfig(args ...string) error {
	cmdArgs := append([]string{"config", "--global"}, args...)
	return execCmd("git", cmdArgs...)
}

func syncGit() error {
	dir, err := homedir.Expand("~/workspace")
	if err != nil {
		return err
	}

	if err := os.Chdir(dir); err != nil {
		return err
	}

	return git.Sync(dir, "pushing local cloud shell changes", true)
}
