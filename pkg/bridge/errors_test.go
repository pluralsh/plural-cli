package bridge

import (
	"errors"
	"testing"
)

func TestBridgeErrorSupportsTypedRecoveryAndUnwrap(t *testing.T) {
	cause := errors.New("connection refused")
	err := &Error{Code: ErrorUnavailable, Operation: "load profile", Err: cause}

	if !IsCode(err, ErrorUnavailable) {
		t.Fatal("IsCode() = false")
	}
	if !errors.Is(err, cause) {
		t.Fatal("bridge error does not unwrap its cause")
	}
}
