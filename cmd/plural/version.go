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
	version = "dev"
	commit  = ""
	date    = time.Now().Format(time.RFC3339)
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

	if !strings.HasPrefix(sha, commit) {
		utils.Warn("Your cli version appears out of date, try updating it with your package manager\n\n")
	}

	return nil
}

func versionInfo(c *cli.Context) error {
	fmt.Println("PLURAL CLI:")
	fmt.Printf("   version\t%s\n", version)
	fmt.Printf("   git commit\t%s\n", commit)
	fmt.Printf("   compiled at\t%s\n", date)
	fmt.Printf("   os/arch\t%s/%s\n", runtime.GOOS, runtime.GOARCH)
	return nil
}
