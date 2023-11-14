package bootstrap

import (
	"os/exec"
	"time"

	"github.com/pkg/errors"
	"github.com/pluralsh/plural-cli/pkg/helm"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var settings = cli.New()

const (
	ciliumRepoName = "cilium"
	ciliumRepoUrl  = "https://helm.cilium.io/"
)

func installCilium(cluster string) error {
	namespace := "kube-system"
	cmd := exec.Command("kind", "export", "kubeconfig", "--name", cluster)
	if err := utils.Execute(cmd); err != nil {
		return err
	}

	if err := helm.AddRepo(ciliumRepoName, ciliumRepoUrl); err != nil {
		return err
	}

	helmConfig, err := helm.GetActionConfig(namespace)
	if err != nil {
		return nil
	}

	cp, err := action.NewInstall(helmConfig).ChartPathOptions.LocateChart("cilium/cilium", settings)
	if err != nil {
		return err
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return err
	}

	histClient := action.NewHistory(helmConfig)
	histClient.Max = 5
	if _, err := histClient.Run(ciliumRepoName); errors.Is(err, driver.ErrReleaseNotFound) {
		instClient := action.NewInstall(helmConfig)
		instClient.Namespace = namespace
		instClient.ReleaseName = ciliumRepoName
		instClient.Timeout = time.Minute * 10

		_, err = instClient.Run(chart, map[string]interface{}{})
		return err
	}
	client := action.NewUpgrade(helmConfig)
	client.Namespace = namespace
	client.Timeout = time.Minute * 10
	_, err = client.Run(ciliumRepoName, chart, map[string]interface{}{})

	return err
}
