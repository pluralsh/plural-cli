package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/urfave/cli"
)

var (
	GitCommit string
	Version   string
)

var BuildDate = time.Now()

func versionInfo(*cli.Context) error {
	fmt.Println("Plural CLI:")
	fmt.Printf("  Version: %s\n", Version)
	fmt.Printf("  Git Commit: %s\n", GitCommit)
	fmt.Printf("  Compiled At: %s\n", BuildDate.String())
	fmt.Printf("  OS: %s\n", runtime.GOOS)
	fmt.Printf("  Arch: %s\n", runtime.GOARCH)
	return nil
}
