package plural

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/utils"
)

const (
	versionPlaceholder = "dev"
)

var (
	Version = versionPlaceholder
	Commit  = ""
	Date    = ""
)

func checkRecency() error {
	if Version == versionPlaceholder || strings.Contains(Version, "-") {
		utils.Warn("\nThis is a development version, which can be significantly different from official releases")
		utils.Warn("\nYou can download latest release from https://github.com/pluralsh/plural-cli/releases/latest\n")
		return nil
	}

	utils.CheckLatestVersion(Version)

	return nil
}

func versionInfo(c *cli.Context) error {
	fmt.Println("PLURAL CLI:")
	fmt.Printf("   version\t%s\n", Version)
	fmt.Printf("   git commit\t%s\n", Commit)
	fmt.Printf("   compiled at\t%s\n", Date)
	fmt.Printf("   os/arch\t%s/%s\n", runtime.GOOS, runtime.GOARCH)

	return checkRecency()
}
