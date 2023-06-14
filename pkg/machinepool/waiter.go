package machinepool

import (
	"context"
	"fmt"
	"time"

	tm "github.com/buger/goterm"
	"github.com/pluralsh/plural/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	clusterapiExp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
)

const (
	waitTime = 40 * 60 * time.Second
)

func ListAll(kubeConf *rest.Config) ([]clusterapiExp.MachinePool, error) {
	mps, err := NewForConfig(kubeConf)
	if err != nil {
		return nil, err
	}

	client := mps.MachinePools("")
	l, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return l.Items, nil
}

func Waiter(kubeConf *rest.Config, namespace string, name string, mpFunc func(mp *clusterapiExp.MachinePool) (bool, error), timeout func() error) error {
	conf := config.Read()
	ctx := context.Background()
	mps, err := NewForConfig(kubeConf)
	if err != nil {
		return err
	}

	client := mps.MachinePools(conf.Namespace(namespace))
	mp, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	tm.Clear()
	if ready, err := mpFunc(mp); ready || err != nil {
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
			mp, ok := event.Object.(*clusterapiExp.MachinePool)
			if !ok {
				return fmt.Errorf("Failed to parse watch event")
			}

			if stop, err := mpFunc(mp); stop || err != nil {
				return err
			}
		case <-time.After(waitTime):
			if err := timeout(); err != nil {
				return err
			}
		}
	}
}

func SilentWait(kubeConf *rest.Config, namespace string, name string) error {
	timeout := func() error {
		return fmt.Errorf("Failed to become ready after 40 minutes, try running `plural cluster mpwatch %s %s` to get an idea where to debug", namespace, name)
	}

	return Waiter(kubeConf, namespace, name, func(mp *clusterapiExp.MachinePool) (bool, error) {
		phase := findReadiness(mp)
		if phase == clusterapiExp.MachinePoolPhaseRunning {
			fmt.Printf("MachinePool %s is finally ready!", name)
			return true, nil
		}
		return false, nil
	}, timeout)
}

func Wait(kubeConf *rest.Config, namespace string, name string) error {
	timeout := func() error {
		return fmt.Errorf("Failed to become ready after 40 minutes, try running `plural cluster mpwatch %s %s` to get an idea where to debug", namespace, name)
	}

	return Waiter(kubeConf, namespace, name, func(mp *clusterapiExp.MachinePool) (bool, error) {
		tm.MoveCursor(1, 1)
		ready := Ready(mp)
		Flush()
		return ready, nil
	}, timeout)
}
