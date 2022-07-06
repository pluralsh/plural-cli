package helm

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	cm "github.com/chartmuseum/helm-push/pkg/chartmuseum"
	"github.com/pluralsh/plural/pkg/config"
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
	client, err := cm.NewClient(
		cm.URL(parsedURL.String()),
		cm.AccessToken(conf.Token),
	)

	if err != nil {
		return nil, err
	}

	resp, err := client.DownloadFile(filePath)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var buff bytes.Buffer
	_, err = io.Copy(&buff, resp.Body)
	return &buff, err
}
