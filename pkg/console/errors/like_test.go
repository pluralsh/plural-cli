package errors_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/gqlerror"

	consoleerrors "github.com/pluralsh/plural-cli/pkg/console/errors"
	clierrors "github.com/pluralsh/plural-cli/pkg/utils/errors"
)

func TestLike(t *testing.T) {
	response := &clientv2.ErrorResponse{GqlErrors: &gqlerror.List{
		nil,
		{Message: "name is required"},
		{Message: "handle has already been taken"},
	}}
	var nilResponse *clientv2.ErrorResponse
	for _, tt := range []struct {
		name string
		err  error
		msg  string
		want bool
	}{
		{name: "nil error", msg: "handle"},
		{name: "typed nil response", err: nilResponse, msg: "handle"},
		{name: "unrelated error with matching text", err: errors.New("handle has already been taken"), msg: "handle"},
		{name: "network error only", err: &clientv2.ErrorResponse{NetworkError: &clientv2.HTTPError{Code: 502}}, msg: "handle"},
		{name: "empty list", err: &clientv2.ErrorResponse{GqlErrors: &gqlerror.List{}}, msg: "handle"},
		{name: "nil entries", err: &clientv2.ErrorResponse{GqlErrors: &gqlerror.List{nil}}, msg: "handle"},
		{name: "match after nil and nonmatching entries", err: response, msg: "handle", want: true},
		{name: "no match", err: response, msg: "permission denied"},
		{name: "wrapped formatted response", err: fmt.Errorf("creating cluster: %w", clierrors.GraphQL(response)), msg: "handle", want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, consoleerrors.Like(tt.err, tt.msg))
		})
	}
}
