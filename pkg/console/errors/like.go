package errors

import (
	"errors"
	"strings"

	client "github.com/Yamashou/gqlgenc/clientv2"
)

func Like(err error, msg string) bool {
	if err == nil {
		return false
	}

	errorResponse := new(client.ErrorResponse)
	ok := errors.As(err, &errorResponse)
	if !ok {
		return false
	}

	return isLike(errorResponse, msg)
}

func isLike(err *client.ErrorResponse, msg string) bool {
	for _, g := range *err.GqlErrors {
		if strings.Contains(g.Message, msg) {
			return true
		}
	}

	return false
}
