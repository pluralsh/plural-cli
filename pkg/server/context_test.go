package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/server"
)

func TestContextConfiguration(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		expectedHTTPStatus int
		expectedResponse   string
	}{
		{
			name:               `update configuration console email address`,
			body:               `{"configuration": {"console":{"email":"newEmail@plural.sh"}}}`,
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   `{"bundles":[],"buckets":[],"domains":[],"smtp":null,"configuration":{"console":{"email":"newEmail@plural.sh"},"minio":{"host":"minio.plural.sh","url":"https://test.plural.sh"}}}`,
		},
		{
			name:               `add new entry to configuration`,
			body:               `{"configuration": {"newEntry":{"test":"test"}}}`,
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   `{"bundles":[],"buckets":[],"domains":[],"smtp":null,"configuration":{"console":{"email":"test@plural.sh","git_user":"test"},"minio":{"host":"minio.plural.sh","url":"https://test.plural.sh"},"newEntry":{"test":"test"}}}`,
		},
		{
			name:               `remove minio url from configuration`,
			body:               `{"configuration": {"console":{"email":"test@plural.sh","git_user":"test"},"minio":{"host":"minio.plural.sh"}}}`,
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   `{"bundles":[],"buckets":[],"domains":[],"smtp":null,"configuration":{"console":{"email":"test@plural.sh","git_user":"test"},"minio":{"host":"minio.plural.sh"}}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			err = os.Chdir(dir)
			assert.NoError(t, err)

			context := manifest.NewContext()
			context.Configuration = genDefaultContextConfiguration()
			err = context.Write(path.Join(dir, "context.yaml"))
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/context/configuration", strings.NewReader(test.body))
			res := httptest.NewRecorder()
			r := server.SetUpRouter()
			r.ServeHTTP(res, req)

			if res.Code != test.expectedHTTPStatus {
				t.Fatalf("Expected HTTP status code %d, got %d: %s", test.expectedHTTPStatus, res.Code, res.Body.String())
			}

			if res.Code == http.StatusOK {
				CompareWithResult(t, res, test.expectedResponse)
			}

			context, err = manifest.ReadContext(manifest.ContextPath())
			assert.NoError(t, err)

			contextBytes, err := json.Marshal(context)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse, string(contextBytes))

		})
	}
}

func genDefaultContextConfiguration() map[string]map[string]interface{} {
	configMap := make(map[string]map[string]interface{})
	configMap["console"] = map[string]interface{}{
		"email":    "test@plural.sh",
		"git_user": "test",
	}
	configMap["minio"] = map[string]interface{}{
		"url":  "https://test.plural.sh",
		"host": "minio.plural.sh",
	}
	return configMap
}
