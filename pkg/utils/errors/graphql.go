package errors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"
)

type graphQLError struct {
	err     error
	message string
}

func (e *graphQLError) Error() string { return e.message }
func (e *graphQLError) Unwrap() error { return e.err }

// GraphQL presents server error messages while retaining the response for errors.As.
// Errors without a usable message are returned unchanged.
func GraphQL(err error) error {
	var formatted *graphQLError
	if errors.As(err, &formatted) {
		return err
	}

	var response *clientv2.ErrorResponse
	if !errors.As(err, &response) || response == nil {
		return err
	}

	var messages []string
	if response.GqlErrors != nil {
		for _, gqlErr := range *response.GqlErrors {
			if gqlErr != nil && strings.TrimSpace(gqlErr.Message) != "" {
				messages = append(messages, gqlErr.Message)
			}
		}
	}

	// Non-2xx responses can contain the same GraphQL errors in the raw HTTP
	// body. Prefer the parsed messages and avoid printing that body again.
	if len(messages) == 0 && response.NetworkError != nil {
		if code := response.NetworkError.Code; code != 0 {
			messages = append(messages, strings.TrimSpace(fmt.Sprintf("HTTP %d %s", code, http.StatusText(code))))
		} else if message := strings.TrimSpace(response.NetworkError.Message); message != "" {
			messages = append(messages, message)
		}
	}
	if len(messages) == 0 {
		return err
	}

	return &graphQLError{
		err:     err,
		message: strings.Replace(err.Error(), response.Error(), strings.Join(messages, "; "), 1),
	}
}

// GraphQLInterceptor formats errors after all other request interceptors finish.
func GraphQLInterceptor(ctx context.Context, req *http.Request, info *clientv2.GQLRequestInfo, res any, next clientv2.RequestInterceptorFunc) error {
	return GraphQL(next(ctx, req, info, res))
}
