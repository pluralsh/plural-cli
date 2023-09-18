package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/pluralsh/plural/pkg/cluster"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
)

const (
	ClusterNotReadyError = "cluster exists but it is not ready yet"
)

// getCluster returns Cluster resource.
func getCluster(kubeContext, name, namespace string) (*capi.Cluster, error) {
	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	err = prov.KubeConfig()
	if err != nil {
		return nil, err
	}

	var kubeConf *rest.Config
	if len(kubeContext) > 0 {
		kubeConf, err = kubernetes.KubeConfigWithContext(kubeContext)
	} else {
		kubeConf, err = kubernetes.KubeConfig()
	}

	if err != nil {
		return nil, err
	}

	conf := config.Read()
	ctx := context.Background()
	clusters, err := cluster.NewForConfig(kubeConf)
	if err != nil {
		return nil, err
	}

	client := clusters.Clusters(conf.Namespace(namespace))
	return client.Get(ctx, name, metav1.GetOptions{})
}

func IsClusterPhase(context, name, namespace string, phase capi.ClusterPhase) (bool, error) {
	c, err := getCluster(context, name, namespace)
	if err != nil {
		return false, err
	}

	return c.Status.GetTypedPhase() == phase, nil
}


// CheckClusterReadiness checks if Cluster API cluster is in ready state.
func CheckClusterReadiness(name, namespace string) (bool, error) {
	utils.Highlight("Checking cluster status")

	err := util.WaitFor(10*time.Second, time.Second, func() (bool, error) {
		utils.Highlight(".")

		c, err := getCluster("", name, namespace)
		if err != nil {
			return false, err
		}

		for _, cond := range c.Status.Conditions {
			if cond.Type == capi.ReadyCondition && cond.Status == "True" {
				return true, nil
			}
		}

		return true, fmt.Errorf(ClusterNotReadyError)
	})

	utils.Highlight("\n")

	return err == nil, err
}
