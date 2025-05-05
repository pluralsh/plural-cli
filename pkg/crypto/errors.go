package crypto

import "fmt"

var (
	errFingerprint = fmt.Errorf("the fingerprint of your encryption key does not match this repos key, we cannot safely decrypt")
)
