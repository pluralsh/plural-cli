package main

import (
	"fmt"
)

var (
	errNoGit      = fmt.Errorf("Could not compare current workspace to origin. Do you have an `origin` remote configured, or does your repo not have an initial commit?")
	errRemoteDiff = fmt.Errorf("Your local workspace is not in sync with remote. Either `git pull` recent changes or `git push` any missed changes.")
	errUnlock     = fmt.Errorf("could not decrypt your repo, this is likely due to using the wrong key at ~/.plural/key. The original key might be in a backup or on your previous machine.")
)
