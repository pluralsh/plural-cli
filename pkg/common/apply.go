package common

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/pluralfile"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"github.com/urfave/cli"
)

func Apply(c *cli.Context) error {
	path, _ := os.Getwd()
	var file = pathing.SanitizeFilepath(filepath.Join(path, "Pluralfile"))
	if c.IsSet("file") {
		file, _ = filepath.Abs(c.String("file"))
	}

	if err := os.Chdir(filepath.Dir(file)); err != nil {
		return err
	}

	plrl, err := pluralfile.Parse(file)
	if err != nil {
		return err
	}

	lock, err := plrl.Lock(file)
	if err != nil {
		return err
	}
	return plrl.Execute(file, lock)
}
