package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pluralsh/plural/pkg/server"
)

func TestHealthcheck(t *testing.T) {
	tests := []struct {
		name               string
		expectedHTTPStatus int
		expectedResponse   string
	}{
		{
			name:               `test health check`,
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   "OK",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/health", strings.NewReader(""))
			res := httptest.NewRecorder()
			r := server.SetUpRouter()
			r.ServeHTTP(res, req)

			if res.Code != test.expectedHTTPStatus {
				t.Fatalf("Expected HTTP status code %d, got %d: %s", test.expectedHTTPStatus, res.Code, res.Body.String())
			}

			if res.Code == http.StatusOK {
				CompareWithResult(t, res, test.expectedResponse)
			}
		})
	}
}

// CompareWithResult a convenience function for comparing http.Body content with response.
func CompareWithResult(t *testing.T, res *httptest.ResponseRecorder, response string) {
	t.Helper()
	bBytes, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal("Unable to read response body")
	}

	r := strings.TrimSpace(response)
	b := strings.TrimSpace(string(bBytes))

	if r != b {
		t.Fatalf("Expected response body to be \n%s \ngot \n%s", r, b)
	}
}
