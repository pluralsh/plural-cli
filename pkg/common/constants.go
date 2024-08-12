package common

import "fmt"

const BackupMsg = "Would you like to back up your encryption key to plural?  If you chose to manage it yourself, you can find it at ~/.plural/key"

const (
	AffirmUp   = "Are you ready to set up your initial management cluster?  You can check the generated terraform/helm to confirm everything looks good first"
	AffirmDown = "Are you ready to destroy your plural infrastructure?  This will destroy all k8s clusters and any data stored within"
)

var (
	ErrNoGit      = fmt.Errorf("Could not compare current workspace to origin. Do you have an `origin` remote configured, or does your repo not have an initial commit?")
	ErrRemoteDiff = fmt.Errorf("Your local workspace is not in sync with remote. Either `git pull` recent changes or `git push` any missed changes.  Also confirm you can authenticate to the origin remote, which you can see with `git remote -v`")
	ErrUnlock     = fmt.Errorf("could not decrypt your repo, this is likely due to using the wrong key at ~/.plural/key. The original key might be in a backup or on your previous machine.")
)
