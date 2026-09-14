package errors_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/gqlerror"

	clierrors "github.com/pluralsh/plural-cli/pkg/utils/errors"
)

func TestGraphQL(t *testing.T) {
	response := &clientv2.ErrorResponse{GqlErrors: &gqlerror.List{
		{Message: "handle has already been taken"},
		nil,
		{Message: " "},
		{Message: "name is required"},
	}}
	want := "handle has already been taken; name is required"
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "multiple messages", err: response, want: want},
		{name: "wrapped response", err: fmt.Errorf("creating cluster: %w", response), want: "creating cluster: " + want},
		{name: "already formatted", err: fmt.Errorf("creating cluster: %w", clierrors.GraphQL(response)), want: "creating cluster: " + want},
		{name: "HTTP error", err: &clientv2.ErrorResponse{NetworkError: &clientv2.HTTPError{Code: 502, Message: "Response body <html>Bad gateway</html>"}}, want: "HTTP 502 Bad Gateway"},
		{name: "unknown HTTP status", err: &clientv2.ErrorResponse{NetworkError: &clientv2.HTTPError{Code: 599}}, want: "HTTP 599"},
		{name: "network message without status", err: &clientv2.ErrorResponse{NetworkError: &clientv2.HTTPError{Message: "connection failed"}}, want: "connection failed"},
		{name: "GraphQL over HTTP error", err: &clientv2.ErrorResponse{GqlErrors: response.GqlErrors, NetworkError: &clientv2.HTTPError{Code: 400, Message: "Response body raw JSON"}}, want: want},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clierrors.GraphQL(tt.err)
			assert.EqualError(t, got, tt.want)
			assert.ErrorIs(t, got, tt.err)
			var original, preserved *clientv2.ErrorResponse
			assert.ErrorAs(t, tt.err, &original)
			assert.ErrorAs(t, got, &preserved)
			assert.Same(t, original, preserved)
		})
	}
}

func TestGraphQLUnchanged(t *testing.T) {
	for _, err := range []error{
		nil,
		errors.New("connection refused"),
		errors.New(`{"message":"unrelated JSON error"}`),
		&clientv2.ErrorResponse{},
		&clientv2.ErrorResponse{GqlErrors: &gqlerror.List{}},
		&clientv2.ErrorResponse{GqlErrors: &gqlerror.List{nil, {Message: ""}}},
	} {
		assert.Equal(t, err, clierrors.GraphQL(err))
	}
}
