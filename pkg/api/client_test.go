package api_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Yamashou/gqlgenc/clientv2"
	consoleclient "github.com/pluralsh/console/go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/console"
	consoleerrors "github.com/pluralsh/plural-cli/pkg/console/errors"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientErrorMessages(t *testing.T) {
	originalTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	for _, tt := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"GraphQL", 200, `{"errors":[{"message":"handle has already been taken","path":["createCluster"],"extensions":{"code":"invalid"}},{"message":"name is required"}]}`, "handle has already been taken; name is required"},
		{"GraphQL with HTTP failure", 400, `{"errors":[{"message":"handle has already been taken"}]}`, "handle has already been taken"},
		{"HTTP failure", 502, `<html>upstream unavailable</html>`, "HTTP 502 Bad Gateway"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			http.DefaultTransport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("Authorization") == "Token token" {
					assert.NotEmpty(t, req.URL.Query().Get("documentId"), "persisted query interceptor must still run")
				}
				return &http.Response{StatusCode: tt.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tt.body))}, nil
			})
			apiClient := api.FromConfig(&config.Config{Endpoint: "api.example.com", Token: "token"})
			consoleClient, err := console.NewConsoleClient("token", "https://console.example.com")
			require.NoError(t, err)

			_, err = apiClient.Me()
			assert.EqualError(t, err, tt.want)
			_, err = consoleClient.CreateCluster(consoleclient.ClusterAttributes{})
			assert.EqualError(t, err, tt.want)
			assert.Equal(t, tt.status != 502, consoleerrors.Like(err, "handle"))
			_, err = consoleClient.ListClusters()
			assert.EqualError(t, err, "ListClusters: "+tt.want)
			var response *clientv2.ErrorResponse
			assert.ErrorAs(t, err, &response)
			assert.Equal(t, tt.status != 502, consoleerrors.Like(err, "handle"))
		})
	}
}

func TestGetErrorResponse(t *testing.T) {
	response := &clientv2.ErrorResponse{GqlErrors: &gqlerror.List{{Message: "permission denied"}}}
	wrapped := fmt.Errorf("fetching cluster: %w", response)
	got := api.GetErrorResponse(wrapped, "GetCluster")
	assert.EqualError(t, got, "GetCluster: fetching cluster: permission denied")
	assert.ErrorIs(t, got, wrapped)
	var preserved *clientv2.ErrorResponse
	assert.ErrorAs(t, got, &preserved)
	assert.Same(t, response, preserved)
	assert.Nil(t, api.GetErrorResponse(nil, "GetCluster"))
	for _, err := range []error{errors.New("connection refused"), errors.New(`{"message":"unrelated"}`)} {
		assert.Same(t, err, api.GetErrorResponse(err, "GetCluster"))
	}
}
