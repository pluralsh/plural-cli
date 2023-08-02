package bootstrap

import (
	"context"
	"time"

	"github.com/pluralsh/plural/pkg/cluster"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapi "sigs.k8s.io/cluster-api/api/v1beta1"
)

func CheckClusterReadiness(name, namespace string) bool {
	prov, err := provider.GetProvider()
	if err != nil {
		return false
	}

	err = prov.KubeConfig()
	if err != nil {
		return false
	}

	kubeConf, err := kubernetes.KubeConfig()
	if err != nil {
		return false
	}

	conf := config.Read()
	ctx := context.Background()
	clusters, err := cluster.NewForConfig(kubeConf)
	if err != nil {
		return false
	}

	client := clusters.Clusters(conf.Namespace(namespace))
	c, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false
	}

	for _, cond := range c.Status.Conditions {
		if cond.Type == clusterapi.ReadyCondition && cond.Status == "True" {
			return true
		}
	}

	return false
}

func CheckClusterReadinessWithRetries(name, namespace string, retries int, sleep time.Duration, log bool) bool {
	if log {
		utils.Highlight("Checking cluster status...\n")
	}

	if CheckClusterReadiness(name, namespace) {
		return true
	}

	if retries--; retries > 0 {
		time.Sleep(sleep)
		return CheckClusterReadinessWithRetries(name, namespace, retries, sleep, false)
	}

	return false
}
