package crypto

import "fmt"

var (
	errFingerprint = fmt.Errorf("The fingerprint of your encryption key does not match this repos key, we cannot safely decrypt")
)
