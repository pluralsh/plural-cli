package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"

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

	if err := os.WriteFile(pathing.SanitizeFilepath(filepath.Join(p, "id_rsa")), []byte(setup.SshPrivateKey), 0600); err != nil {
		return fmt.Errorf("error writing ssh private key: %w", err)
	}
	if err := os.WriteFile(pathing.SanitizeFilepath(filepath.Join(p, "id_rsa.pub")), []byte(setup.SshPublicKey), 0644); err != nil {
		return fmt.Errorf("error writing ssh public key: %w", err)
	}

	if err := execCmd("ssh-add", pathing.SanitizeFilepath(filepath.Join(p, "id_rsa"))); err != nil {
		return fmt.Errorf("error adding ssh key to agent: %w", err)
	}

	dir, err := homedir.Expand("~/workspace")
	if err != nil {
		return fmt.Errorf("error getting the workspace: %w", err)
	}

	if err := execCmd("git", "clone", setup.GitUrl, dir); err != nil {
		return fmt.Errorf("error cloning the repository: %w", err)
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("error changing directory: %w", err)
	}
	if err := gitConfig("user.email", setup.User.Email); err != nil {
		return fmt.Errorf("error during git config: %w", err)
	}

	name := "plural-shell"
	if setup.User.GitUser != "" {
		name = setup.User.GitUser
	}
	if err := gitConfig("user.name", name); err != nil {
		return fmt.Errorf("error during git config: %w", err)
	}

	if err := execCmd("plural", "crypto", "init"); err != nil {
		return fmt.Errorf("error running plural crypt init: %w", err)
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
