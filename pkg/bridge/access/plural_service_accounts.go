package access

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/config"
)

// PluralServiceAccountSource queries the Plural App API without exposing its
// GraphQL transport to the Access screen.
type PluralServiceAccountSource struct {
	Credentials bridge.CredentialStore
	Client      *http.Client
}

func (s PluralServiceAccountSource) ListServiceAccounts(ctx context.Context, profile Profile, query string) ([]ServiceAccount, error) {
	credential, err := s.Credentials.Get(ctx, profile.ID)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(map[string]any{
		"query":     `query TUIServiceAccounts($q: String, $serviceAccount: Boolean!) { users(q: $q, serviceAccount: $serviceAccount, first: 100) { edges { node { id email } } } }`,
		"variables": map[string]any{"q": query, "serviceAccount": true},
	})
	if err != nil {
		return nil, err
	}
	conf := config.Config{Endpoint: profile.Endpoint}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, conf.Url(), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+credential)
	request.Header.Set("Content-Type", "application/json")
	client := s.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("service-account query returned %s", response.Status)
	}
	var result struct {
		Data struct {
			Users struct {
				Edges []struct {
					Node struct{ ID, Email string } `json:"node"`
				} `json:"edges"`
			} `json:"users"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("service-account query failed: %s", result.Errors[0].Message)
	}
	accounts := make([]ServiceAccount, 0, len(result.Data.Users.Edges))
	for _, edge := range result.Data.Users.Edges {
		accounts = append(accounts, ServiceAccount{ID: edge.Node.ID, Email: edge.Node.Email})
	}
	return accounts, nil
}
