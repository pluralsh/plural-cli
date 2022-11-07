package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/server"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/stretchr/testify/assert"
)

func TestGetConfiguration(t *testing.T) {
	tests := []struct {
		name               string
		expectedHTTPStatus int
		expectedResponse   string
	}{
		{
			name:               `update configuration console email address`,
			expectedHTTPStatus: http.StatusOK,
			expectedResponse:   `{"workspace":{},"git":{"url":"git@git.test.com:portfolio/space.space_name.git","root":"%s","name":"%s","branch":"master"},"context_configuration":{"console":{"email":"test@plural.sh","git_user":"test"},"minio":{"host":"minio.plural.sh","url":"https://test.plural.sh"}}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// create temp environment
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			err = os.Chdir(dir)
			assert.NoError(t, err)

			pm := genProjectManifest()
			io, err := json.Marshal(pm)
			assert.NoError(t, err)
			err = os.WriteFile(path.Join(dir, "workspace.yaml"), io, 0644)
			assert.NoError(t, err)

			context := manifest.NewContext()
			context.Configuration = genDefaultContextConfiguration()
			err = context.Write(path.Join(dir, "context.yaml"))
			assert.NoError(t, err)

			// To don't override ~/.gitconfig
			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")
			_, err = git.Init()
			assert.NoError(t, err)
			_, err = git.GitRaw("config", "--global", "user.email", "test@plural.com")
			assert.NoError(t, err)
			_, err = git.GitRaw("config", "--global", "user.name", "test")
			assert.NoError(t, err)
			_, err = git.GitRaw("add", "-A")
			assert.NoError(t, err)
			_, err = git.GitRaw("commit", "-m", "init")
			assert.NoError(t, err)
			_, err = git.GitRaw("remote", "add", "origin", "git@git.test.com:portfolio/space.space_name.git")
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/v1/configuration", strings.NewReader(""))
			res := httptest.NewRecorder()
			r := server.SetUpRouter()
			r.ServeHTTP(res, req)

			if res.Code != test.expectedHTTPStatus {
				t.Fatalf("Expected HTTP status code %d, got %d: %s", test.expectedHTTPStatus, res.Code, res.Body.String())
			}

			test.expectedResponse = fmt.Sprintf(test.expectedResponse, dir, filepath.Base(dir))

			if res.Code == http.StatusOK {
				CompareWithResult(t, res, test.expectedResponse)
			}
		})
	}
}

func genProjectManifest() *manifest.VersionedProjectManifest {
	return &manifest.VersionedProjectManifest{
		ApiVersion: "plural.sh/v1alpha1",
		Kind:       "ProjectManifest",
		Spec: &manifest.ProjectManifest{
			Cluster: "abc",
			Bucket:  "def",
			Project: "test",
		},
	}
}
