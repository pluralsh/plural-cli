package bridge

import (
	"errors"
	"fmt"
)

func (e *Error) Error() string {
	if e.Operation == "" {
		return fmt.Sprintf("%s: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("%s: %s: %v", e.Operation, e.Code, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

// IsCode reports whether err contains a bridge error with code.
func IsCode(err error, code ErrorCode) bool {
	var bridgeErr *Error
	return errors.As(err, &bridgeErr) && bridgeErr.Code == code
}
