package access

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestPluralServiceAccountSourceUsesActiveBaseCredential(t *testing.T) {
	credentials := &memoryCredentials{values: map[string]string{"app": "base-secret"}}
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got := request.Header.Get("Authorization"); got != "Bearer base-secret" {
			t.Fatalf("Authorization = %q", got)
		}
		body, _ := io.ReadAll(request.Body)
		if !strings.Contains(string(body), `"serviceAccount":true`) || !strings.Contains(string(body), `"q":"deploy"`) {
			t.Fatalf("request body = %s", body)
		}
		return &http.Response{StatusCode: 200, Status: "200 OK", Body: io.NopCloser(strings.NewReader(`{"data":{"users":{"edges":[{"node":{"id":"sa-1","email":"deploy@example.com"}}]}}}`)), Header: make(http.Header)}, nil
	})}
	source := PluralServiceAccountSource{Credentials: credentials, Client: client}
	accounts, err := source.ListServiceAccounts(context.Background(), Profile{ID: "app", Endpoint: "app.plural.sh"}, "deploy")
	if err != nil {
		t.Fatalf("ListServiceAccounts() error = %v", err)
	}
	if len(accounts) != 1 || accounts[0].Email != "deploy@example.com" {
		t.Fatalf("accounts = %#v", accounts)
	}
}
