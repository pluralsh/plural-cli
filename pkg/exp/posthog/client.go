package posthog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/posthog/posthog-go"
)

const (
	publicAPIKey = "phc_r0v4jbKz8Rr27mfqgO15AN5BMuuvnU8hCFedd6zpSDy"
	endpoint     = "posthog.plural.sh"
)

type posthogClient struct {
	config      Config
	contentType string
	client      http.Client
}

func (this *posthogClient) IsFeatureEnabled(payload FeatureFlagPayload) (bool, error) {
	values := posthog.DecideRequestData{
		ApiKey:     this.config.APIKey,
		DistinctId: payload.DistinctId,
		PersonProperties: posthog.Properties{
			"email": payload.DistinctId,
		},
	}
	data, err := json.Marshal(values)
	if err != nil {
		return false, err
	}

	resp, err := this.client.Post(this.decideEndpoint(), this.contentType, bytes.NewBuffer(data))
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	decide := new(posthog.DecideResponse)
	err = json.NewDecoder(resp.Body).Decode(decide)
	if err != nil {
		return false, err
	}

	enabled := decide.FeatureFlags[payload.Key]

	switch v := enabled.(type) {
	case nil:
		return false, nil
	case bool:
		return v, nil
	case string:
		return v == "true", nil
	default:
		return true, nil
	}
}

func (this *posthogClient) decideEndpoint() string {
	return fmt.Sprintf("https://%s/decide/?v=3", this.config.Endpoint)
}

func New() Client {
	return &posthogClient{
		config: Config{
			APIKey:   publicAPIKey,
			Endpoint: endpoint,
		},
		contentType: "application/json",
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
}
