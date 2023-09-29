package console

import (
	"errors"
	"fmt"
	"time"

	"github.com/pluralsh/plural/pkg/helm"
	"helm.sh/helm/v3/pkg/action"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var settings = cli.New()

const (
	repoName = "deployop"
	repoUrl  = "https://pluralsh.github.io/deployment-operator"
)

func InstallAgent(url, token, namespace string) error {

	vals := map[string]interface{}{
		"secrets": map[string]string{
			"deployToken": token,
		},
		"consoleUrl": url,
	}

	if err := helm.AddRepo(repoName, repoUrl); err != nil {
		return err
	}

	helmConfig, err := helm.GetActionConfig(namespace)
	if err != nil {
		return nil
	}

	cp, err := action.NewInstall(helmConfig).ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", repoName, "deployment-operator"), settings)
	if err != nil {
		return err
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return err
	}

	histClient := action.NewHistory(helmConfig)
	histClient.Max = 5
	if _, err := histClient.Run(repoName); errors.Is(err, driver.ErrReleaseNotFound) {
		instClient := action.NewInstall(helmConfig)
		instClient.Namespace = namespace
		instClient.ReleaseName = repoName
		instClient.Timeout = time.Minute * 10

		_, err = instClient.Run(chart, vals)
		return err
	}
	client := action.NewUpgrade(helmConfig)
	client.Namespace = namespace
	client.Timeout = time.Minute * 10
	_, err = client.Run(repoName, chart, vals)
	return nil
}
