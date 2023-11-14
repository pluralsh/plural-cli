package tests

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

func testGit(ctx *manifest.Context, test *api.RecipeTest) error {
	args := collectArguments(test.Args, ctx)
	auth, err := authMethod(args)
	if err != nil {
		return err
	}
	url := args["url"].Val.(string)
	dir, err := os.MkdirTemp("", "repo")
	if err != nil {
		return err
	}

	defer os.RemoveAll(dir)
	fmt.Println("~~> Attempting to clone repo in a temporary directory...")
	_, err = git.Clone(auth, url, dir)
	return err
}

func authMethod(args map[string]*ContextValue) (transport.AuthMethod, error) {
	if arg, ok := args["password"]; ok && arg.Present {
		if pass, ok := arg.Val.(string); ok && pass != "" {
			if user, ok := args["username"].Val.(string); ok {
				return git.BasicAuth(user, pass)
			}
			return nil, fmt.Errorf("No valid username/password pair for basic auth")
		}
	}

	urlArg := args["url"]
	if !urlArg.Present {
		return nil, fmt.Errorf("requires a git url")
	}

	url, ok := urlArg.Val.(string)
	if !ok {
		return nil, fmt.Errorf("No valid git url")
	}

	privateKeyArg := args["private_key"]
	if !privateKeyArg.Present {
		return nil, fmt.Errorf("requires a ssh private key for authentication")
	}

	pk, ok := privateKeyArg.Val.(string)
	if !ok {
		return nil, fmt.Errorf("No valid git ssh private key")
	}

	passphrase := ""
	if passArg, ok := args["passphrase"]; ok && passArg.Present {
		if pass, ok := passArg.Val.(string); ok {
			passphrase = pass
		}
	}

	user, _, _, _, err := git.UrlComponents(url)
	if err != nil {
		return nil, err
	}
	return git.SSHAuth(user, pk, passphrase)
}
