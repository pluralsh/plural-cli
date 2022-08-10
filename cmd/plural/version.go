package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
)

const (
	versionPlaceholder = "dev"
	latestURI          = "https://api.github.com/repos/pluralsh/plural-cli/releases/latest"
)

var (
	version = versionPlaceholder
	commit  = ""
	date    = ""
)

func getLatestVersion() (res string, err error) {
	resp, err := http.Get(latestURI)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var ghResp struct {
		Tag_Name string
	}
	err = json.Unmarshal(body, &ghResp)

	res = ghResp.Tag_Name
	return
}

func checkRecency() error {
	if version == versionPlaceholder || strings.Contains(version, "-") {
		utils.Warn("\nThis is a development version, which can be significantly different from official releases")
		utils.Warn("\nYou can download latest release from https://github.com/pluralsh/plural-cli/releases/latest\n")
		return nil
	}

	latestVersion, err := getLatestVersion()
	if err != nil {
		return err
	}

	if !strings.HasSuffix(latestVersion, version) {
		utils.Warn("\nYour version appears out of date, try updating it with your package manager\n")
	}

	return nil
}

func versionInfo(c *cli.Context) error {
	fmt.Println("PLURAL CLI:")
	fmt.Printf("   version\t%s\n", version)
	fmt.Printf("   git commit\t%s\n", commit)
	fmt.Printf("   compiled at\t%s\n", date)
	fmt.Printf("   os/arch\t%s/%s\n", runtime.GOOS, runtime.GOARCH)

	return checkRecency()
}
