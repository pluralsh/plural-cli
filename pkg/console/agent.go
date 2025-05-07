package console

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/pluralsh/plural-cli/pkg/helm"
	"github.com/pluralsh/polly/algorithms"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	ChartName         = "deployment-operator"
	ReleaseName       = "deploy-operator"
	RepoUrl           = "https://pluralsh.github.io/deployment-operator"
	OperatorNamespace = "plrl-deploy-operator"
)

func IsAlreadyAgentInstalled(k8sClient *kubernetes.Clientset) (bool, error) {
	dl, err := k8sClient.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=deployment-operator",
	})
	if err != nil {
		return false, err
	}
	for _, deployment := range dl.Items {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if strings.Contains(container.Image, "pluralsh/deployment-operator") {
				return true, nil
			}
		}
	}

	return false, nil
}

func InstallAgent(consoleURL, token, namespace, version, helmChartLoc string, values map[string]interface{}) error {
	settings := cli.New()
	vals := map[string]interface{}{
		"secrets":    map[string]string{"deployToken": token},
		"consoleUrl": consoleURL,
	}
	vals = algorithms.Merge(vals, values)

	config, err := helm.GetActionConfig(namespace)
	if err != nil {
		return err
	}

	chartLoc := fmt.Sprintf("%s/%s", ReleaseName, ChartName)
	if helmChartLoc == "" {
		fmt.Println("Adding default Repo for deployment operator chart:", RepoUrl)
		if err := helm.AddRepo(ReleaseName, RepoUrl); err != nil {
			return err
		}
	} else {
		fmt.Println("Using custom helm chart url:", chartLoc)
		chartLoc = helmChartLoc
	}

	newInstallAction := action.NewInstall(config)
	newInstallAction.Version = version

	cp, err := action.NewInstall(config).LocateChart(chartLoc, settings)
	if err != nil {
		return err
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return err
	}

	histClient := action.NewHistory(config)
	histClient.Max = 5
	_, err = histClient.Run(ReleaseName)

	if errors.Is(err, driver.ErrReleaseNotFound) {
		return installAgent(config, chart, namespace, vals)
	}
	return upgradeAgent(config, chart, namespace, vals)
}

func installAgent(config *action.Configuration, chart *chart.Chart, namespace string, values map[string]interface{}) error {
	fmt.Println("installing deployment operator...")
	instClient := action.NewInstall(config)
	instClient.Namespace = namespace
	instClient.ReleaseName = ReleaseName
	instClient.Timeout = time.Minute * 5
	_, err := instClient.Run(chart, values)
	return err
}

func upgradeAgent(config *action.Configuration, chart *chart.Chart, namespace string, values map[string]interface{}) error {
	fmt.Println("upgrading deployment operator...")
	client := action.NewUpgrade(config)
	client.Namespace = namespace
	client.Timeout = time.Minute * 5
	_, err := client.Run(ReleaseName, chart, values)
	return err
}

func UninstallAgent(namespace string) error {
	return helm.Uninstall(ReleaseName, namespace)
}
