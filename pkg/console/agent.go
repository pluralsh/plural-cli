package console

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/pluralsh/console/go/polly/algorithms"
	"github.com/pluralsh/plural-cli/pkg/helm"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/samber/lo"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
)

const (
	ChartName         = "deployment-operator"
	ReleaseName       = "deploy-operator"
	RepoUrl           = "https://pluralsh.github.io/deployment-operator"
	OperatorNamespace = "plrl-deploy-operator"

	crdFieldManager = "helm"
)

func fetchVendoredAgentChart(consoleURL string) (string, string, error) {
	parsedConsoleURL, err := url.Parse(consoleURL)
	if err != nil {
		return "", "", fmt.Errorf("cannot parse console URL: %s", err.Error())
	}

	directory, err := os.MkdirTemp("", "agent-chart-")
	if err != nil {
		return directory, "", fmt.Errorf("cannot create directory: %s", err.Error())
	}

	agentChartURL := fmt.Sprintf("https://%s/ext/v1/agent/chart", parsedConsoleURL.Host)
	agentChartPath := filepath.Join(directory, "agent-chart.tgz")
	if err = utils.DownloadFile(agentChartPath, agentChartURL); err != nil {
		return directory, "", fmt.Errorf("cannot download agent chart: %s", err.Error())
	}

	return directory, agentChartPath, nil
}

func getRepositoryAgentChart(install *action.Install) (*chart.Chart, error) {
	if err := helm.AddRepo(ReleaseName, RepoUrl); err != nil {
		return nil, err
	}

	chartName := fmt.Sprintf("%s/%s", ReleaseName, ChartName)
	path, err := install.LocateChart(chartName, cli.New())
	if err != nil {
		return nil, err
	}

	return loader.Load(path)
}

func getCustomAgentChart(install *action.Install, url string) (*chart.Chart, error) {
	cp, err := install.LocateChart(url, cli.New())
	if err != nil {
		return nil, err
	}

	return loader.Load(cp)
}

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

func InstallAgent(consoleURL, token, namespace, version, helmChartLoc, clusterID string, values map[string]interface{}) error {
	vals := algorithms.Merge(map[string]interface{}{
		"secrets":    map[string]string{"deployToken": token},
		"consoleUrl": consoleURL,
	}, values)

	if clusterID != "" {
		vals["clusterId"] = clusterID
	}

	config, err := helm.GetActionConfig(namespace)
	if err != nil {
		return err
	}

	install := action.NewInstall(config)
	install.Version = version

	var chart *chart.Chart
	if helmChartLoc != "" {
		fmt.Println("using custom Helm chart: ", helmChartLoc)
		chart, err = getCustomAgentChart(install, helmChartLoc)
		if err != nil {
			return err
		}
	} else {
		workingDir, chartPath, err := fetchVendoredAgentChart(consoleURL)
		if workingDir != "" {
			defer func(path string) {
				if err := os.RemoveAll(path); err != nil {
					panic(fmt.Sprintf("could not remove temporary working directory, got error: %s", err))
				}
			}(workingDir)
		}
		if err != nil {
			fmt.Printf("using default repo as vendored agent chart could not be fetched, got error: %s\n", err)
			chart, err = getRepositoryAgentChart(install)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("using vendored agent chart")
			chart, err = loader.Load(chartPath)
			if err != nil {
				return err
			}
		}
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
	if err := syncCRDs(config, chart); err != nil {
		return err
	}

	fmt.Println("upgrading deployment operator...")
	client := action.NewUpgrade(config)
	client.Namespace = namespace
	client.Timeout = time.Minute * 5
	_, err := client.Run(ReleaseName, chart, values)
	return err
}

// syncCRDs server-side applies the chart's crds/ directory, which helm upgrade never touches.
// Client-side apply can't be used since some CRDs exceed the 256KB last-applied-configuration annotation limit.
func syncCRDs(config *action.Configuration, chart *chart.Chart) error {
	crds := chart.CRDObjects()
	if len(crds) == 0 {
		return nil
	}

	fmt.Printf("syncing %d deployment operator CRDs...\n", len(crds))
	applied := kube.ResourceList{}
	for _, crd := range crds {
		resources, err := config.KubeClient.Build(bytes.NewBuffer(crd.File.Data), false)
		if err != nil {
			return fmt.Errorf("failed to build CRD %s: %w", crd.Name, err)
		}

		for _, info := range resources {
			data, err := runtime.Encode(unstructured.UnstructuredJSONScheme, info.Object)
			if err != nil {
				return fmt.Errorf("failed to encode CRD %s: %w", info.Name, err)
			}

			obj, err := resource.NewHelper(info.Client, info.Mapping).
				WithFieldManager(crdFieldManager).
				Patch(info.Namespace, info.Name, types.ApplyPatchType, data, &metav1.PatchOptions{Force: lo.ToPtr(true)})
			if err != nil {
				return fmt.Errorf("failed to apply CRD %s: %w", info.Name, err)
			}

			if err := info.Refresh(obj, true); err != nil {
				return err
			}
			applied = append(applied, info)
		}
	}

	if err := config.KubeClient.Wait(applied, 60*time.Second); err != nil {
		return fmt.Errorf("failed waiting for CRDs to be established: %w", err)
	}

	// Newly added CRDs won't be in the cached discovery data used to render the upgrade.
	if discoveryClient, err := config.RESTClientGetter.ToDiscoveryClient(); err == nil && discoveryClient != nil {
		discoveryClient.Invalidate()
	}
	if restMapper, err := config.RESTClientGetter.ToRESTMapper(); err == nil {
		if resettable, ok := restMapper.(meta.ResettableRESTMapper); ok {
			resettable.Reset()
		}
	}

	return nil
}

func UninstallAgent(namespace string) error {
	return helm.Uninstall(ReleaseName, namespace)
}
