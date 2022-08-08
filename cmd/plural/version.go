package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
)

var (
	GitCommit  = "n/a"
	Version    = "dev"
	CompiledAt = time.Now().Format(time.RFC3339)
)

const latestUri = "https://api.github.com/repos/pluralsh/plural-cli/commits/master"

func latestVersion() (res string, err error) { //nolint:deadcode,unused
	resp, err := http.Get(latestUri)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var ghResp struct {
		Sha string
	}
	err = json.Unmarshal(body, &ghResp)
	res = ghResp.Sha
	return
}

func checkRecency() error { //nolint:deadcode,unused
	sha, err := latestVersion()
	if err != nil {
		return err
	}

	if !strings.HasPrefix(sha, GitCommit) {
		utils.Warn("Your cli version appears out of date, try updating it with your package manager\n\n")
	}

	return nil
}

func versionInfo(c *cli.Context) error {
	fmt.Println("Plural CLI:")
	fmt.Printf("  Version: %s\n", Version)
	fmt.Printf("  Git Commit: %s\n", GitCommit)
	fmt.Printf("  Compiled At: %s\n", CompiledAt)
	fmt.Printf("  OS: %s\n", runtime.GOOS)
	fmt.Printf("  Arch: %s\n", runtime.GOARCH)
	return nil
}
