package tests

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils/git"
)

func testGit(ctx *manifest.Context, test *api.RecipeTest) error {
	args := collectArguments(test.Args, ctx)
	auth, err := authMethod(args)
	if err != nil {
		return err
	}
	url := args["url"].Val.(string)
	dir, err := ioutil.TempDir("", "repo")
	if err != nil {
		return err
	}

	defer os.RemoveAll(dir)
	fmt.Println("~~> Attempting to clone repo in a temporary directory...")
	_, err = git.Clone(auth, url, dir)
	return err
}

func authMethod(args map[string]*ContextValue) (transport.AuthMethod, error) {
	if arg, ok := args["password"]; ok && arg.Present && arg.Val.(string) != "" {
		username := args["username"]
		return git.BasicAuth(username.Val.(string), arg.Val.(string))
	}

	urlArg := args["url"]
	if !urlArg.Present {
		return nil, fmt.Errorf("requires a git url")
	}
	url := urlArg.Val.(string)
	privateKeyArg := args["private_key"]
	if !privateKeyArg.Present {
		return nil, fmt.Errorf("requires a ssh private key for authentication")
	}
	pk := privateKeyArg.Val.(string)
	passArg, ok := args["passphrase"]
	passphrase := ""
	if ok && passArg.Present {
		passphrase = passArg.Val.(string)
	}

	user, _, _, _, err := git.UrlComponents(url)
	if err != nil {
		return nil, err
	}
	return git.SSHAuth(user, pk, passphrase)
}
