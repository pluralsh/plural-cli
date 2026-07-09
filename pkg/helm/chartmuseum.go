package helm

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/config"
	"helm.sh/helm/v3/pkg/getter"
)

type ChartMuseum struct{}

var ChartMuseumProvider = getter.Provider{
	Schemes: []string{"cm"},
	New: func(options ...getter.Option) (getter.Getter, error) {
		return &ChartMuseum{}, nil
	},
}

func (c *ChartMuseum) Get(fileUrl string, options ...getter.Option) (*bytes.Buffer, error) {
	parsedURL, err := url.Parse(fileUrl)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(parsedURL.Path, "/")
	numParts := len(parts)
	if numParts <= 1 {
		return nil, fmt.Errorf("invalid file url: %s", fileUrl)
	}

	filePath := parts[numParts-1]

	numRemoveParts := 1
	if parts[numParts-2] == "charts" {
		numRemoveParts++
		filePath = "charts/" + filePath
	}

	parsedURL.Path = strings.Join(parts[:numParts-numRemoveParts], "/")
	parsedURL.Scheme = "https"
	conf := config.Read()
	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.URL.Path = path.Join(req.URL.Path, filePath)
	if conf.Token != "" {
		req.Header.Set("Authorization", "Bearer "+conf.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("failed to download chart %s: status %s", fileUrl, resp.Status)
		}
		return nil, fmt.Errorf("failed to download chart %s: status %s: %s", fileUrl, resp.Status, strings.TrimSpace(string(body)))
	}
	var buff bytes.Buffer
	_, err = io.Copy(&buff, resp.Body)
	return &buff, err
}
