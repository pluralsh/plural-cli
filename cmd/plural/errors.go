package main

import (
	"fmt"
)

var (
	noGit      = fmt.Errorf("could not compare current workspace to origin, do you have an `origin` remote configured, or does your repo not have an initial commit")
	remoteDiff = fmt.Errorf("your local workspace is not in sync with remote, either `git pull` recent changes or `git push` any missed changes")
)
