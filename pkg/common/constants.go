package common

import "fmt"

const BackupMsg = "Would you like to back up your repo encryption key to plural?  If you chose to manage it yourself, you can find it at ~/.plural/key"

const (
	AffirmUp   = "Are you ready to set up your initial management cluster?  You can check the generated terraform/helm to confirm everything looks good first"
	AffirmDown = "Are you ready to destroy your plural infrastructure?  This wil destroy your management cluster, and any data stored within"
)

var (
	ErrUnlock = fmt.Errorf("could not decrypt your repo, this is likely due to using the wrong key at ~/.plural/key, the original key might be in a backup or on your previous machine")
)
