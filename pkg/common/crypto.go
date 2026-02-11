package common

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/urfave/cli"
)

const (
	GitAttributesFile = ".gitattributes"
	GitIgnoreFile     = ".gitignore"
)

const Gitattributes = `context.yaml filter=plural-crypt diff=plural-crypt
workspace.yaml filter=plural-crypt diff=plural-crypt
workspace.yaml* filter=plural-crypt diff=plural-crypt
helm-values/*.yaml filter=plural-crypt diff=plural-crypt
.env filter=plural-crypt diff=plural-crypt
.gitattributes !filter !diff
`

func EnsureGitAttributes() error {
	if utils.Exists(GitAttributesFile) {
		return nil
	}

	return utils.WriteFile(GitAttributesFile, []byte(Gitattributes))
}

const Gitignore = `/**/.terraform
/**/.terraform*
/**/terraform.tfstate*
/bin
*~
.idea
*.swp
*.swo
.DS_STORE
.vscode
context.yaml*
temp
`

func EnsureGitIgnore() error {
	if utils.Exists(GitIgnoreFile) {
		return nil
	}

	return utils.WriteFile(GitIgnoreFile, []byte(Gitignore))
}

func CryptoInit(_ *cli.Context) error {
	encryptConfig := [][]string{
		{"filter.plural-crypt.smudge", "plural crypto decrypt"},
		{"filter.plural-crypt.clean", "plural crypto encrypt"},
		{"filter.plural-crypt.required", "true"},
		{"diff.plural-crypt.textconv", "plural crypto decrypt"},
	}

	utils.Highlight("Creating git encryption filters\n")
	for _, conf := range encryptConfig {
		if err := GitConfig(conf[0], conf[1]); err != nil {
			return err
		}
	}

	if err := EnsureGitAttributes(); err != nil {
		return err
	}

	if err := EnsureGitIgnore(); err != nil {
		return err
	}

	_, err := crypto.Build()
	return err
}

func HandleUnlock(_ *cli.Context) error {
	_, err := crypto.Build()
	if err != nil {
		return err
	}

	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	// fixes Invalid cross-device link when using os.Rename
	gitIndexDir, err := filepath.Abs(filepath.Join(repoRoot, ".git"))
	if err != nil {
		return err
	}
	gitIndex := filepath.Join(gitIndexDir, "index")
	dump, err := os.CreateTemp(gitIndexDir, "index.bak")
	if err != nil {
		return err
	}
	if err := os.Rename(gitIndex, dump.Name()); err != nil {
		return err
	}

	if err := GitCommand("checkout", "HEAD", "--", repoRoot).Run(); err != nil {
		_ = os.Rename(dump.Name(), gitIndex)
		return ErrUnlock
	}

	os.Remove(dump.Name())
	return nil
}
