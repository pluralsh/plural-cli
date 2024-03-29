package console

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/pluralsh/plural-cli/pkg/helm"
	"github.com/pluralsh/polly/algorithms"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

const (
	ChartName         = "deployment-operator"
	ReleaseName       = "deploy-operator"
	RepoUrl           = "https://pluralsh.github.io/deployment-operator"
	OperatorNamespace = "plrl-deploy-operator"
)

func InstallAgent(url, token, namespace string, values map[string]interface{}) error {
	settings := cli.New()
	vals := map[string]interface{}{
		"secrets": map[string]string{
			"deployToken": token,
		},
		"consoleUrl": url,
	}
	vals = algorithms.Merge(vals, values)

	if err := helm.AddRepo(ReleaseName, RepoUrl); err != nil {
		return err
	}

	helmConfig, err := helm.GetActionConfig(namespace)
	if err != nil {
		return err
	}

	cp, err := action.NewInstall(helmConfig).ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", ReleaseName, ChartName), settings)
	if err != nil {
		return err
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return err
	}

	histClient := action.NewHistory(helmConfig)
	histClient.Max = 5

	if _, err = histClient.Run(ReleaseName); errors.Is(err, driver.ErrReleaseNotFound) {
		fmt.Println("installing deployment operator...")
		instClient := action.NewInstall(helmConfig)
		instClient.Namespace = namespace
		instClient.ReleaseName = ReleaseName
		instClient.Timeout = time.Minute * 5
		_, err = instClient.Run(chart, vals)
		if err != nil {
			return err
		}
		return nil
	}
	fmt.Println("upgrading deployment operator...")
	client := action.NewUpgrade(helmConfig)
	client.Namespace = namespace
	client.Timeout = time.Minute * 5
	_, err = client.Run(ReleaseName, chart, vals)
	return err
}

func UninstallAgent(namespace string) error {
	return helm.Uninstall(ReleaseName, namespace)
}
