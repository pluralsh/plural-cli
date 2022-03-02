package main

import (
	"fmt"
)

var (
 noGit = fmt.Errorf("Could not compare current workspace to origin, do you have an `origin` remote configured, or does your repo not have an inital commit?")
 remoteDiff = fmt.Errorf("Your local workspace is not in sync with remote, either `git pull` recent changes or `git push` any missed changes")
)