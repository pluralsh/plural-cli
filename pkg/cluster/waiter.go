package cluster

import (
	"context"
	"fmt"
	"time"

	tm "github.com/buger/goterm"
	"github.com/pluralsh/plural-cli/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	clusterapi "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	waitTime = 40 * 60 * time.Second
)

func Waiter(kubeConf *rest.Config, namespace string, name string, clustFunc func(cluster *clusterapi.Cluster) (bool, error), timeout func() error) error {
	conf := config.Read()
	ctx := context.Background()
	clusters, err := NewForConfig(kubeConf)
	if err != nil {
		return err
	}

	client := clusters.Clusters(conf.Namespace(namespace))
	cluster, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	tm.Clear()
	if ready, err := clustFunc(cluster); ready || err != nil {
		return err
	}

	watcher, err := WatchNamespace(ctx, client)
	if err != nil {
		return err
	}

	ch := watcher.ResultChan()
	for {
		select {
		case event := <-ch:
			tm.Clear()
			cluster, ok := event.Object.(*clusterapi.Cluster)
			if !ok {
				return fmt.Errorf("failed to parse watch event")
			}

			if stop, err := clustFunc(cluster); stop || err != nil {
				return err
			}
		case <-time.After(waitTime):
			if err := timeout(); err != nil {
				return err
			}
		}
	}
}

func Wait(kubeConf *rest.Config, namespace string, name string) error {
	timeout := func() error {
		return fmt.Errorf("Failed to become ready after 40 minutes, try running `plural cluster watch %s %s` to get an idea where to debug", namespace, name)
	}

	return Waiter(kubeConf, namespace, name, func(cluster *clusterapi.Cluster) (bool, error) {
		tm.MoveCursor(1, 1)
		ready := Ready(cluster)
		Flush()
		return ready, nil
	}, timeout)
}
