package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli/v2"
)

const (
	versionPlaceholder = "dev"
)

var (
	version = versionPlaceholder
	commit  = ""
	date    = ""
)

func checkRecency() error {
	if version == versionPlaceholder || strings.Contains(version, "-") {
		utils.Warn("\nThis is a development version, which can be significantly different from official releases")
		utils.Warn("\nYou can download latest release from https://github.com/pluralsh/plural-cli/releases/latest\n")
		return nil
	}

	utils.CheckLatestVersion(version)

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
