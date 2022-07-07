package application

import (
	"context"
	"fmt"
	"time"

	tm "github.com/buger/goterm"
	"github.com/pluralsh/plural/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/application/api/v1beta1"
)

const (
	waitTime = 5 * 60 * time.Second
)

func Waiter(kubeConf *rest.Config, repo string, appFunc func(app *v1beta1.Application) (bool, error), timeout func() error) error {
	conf := config.Read()
	ctx := context.Background()
	apps, err := NewForConfig(kubeConf)
	if err != nil {
		return err
	}

	client := apps.Applications(conf.Namespace(repo))
	app, err := client.Get(ctx, repo, metav1.GetOptions{})
	if err != nil {
		return err
	}

	tm.Clear()
	if ready, err := appFunc(app); ready || err != nil {
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
			app, ok := event.Object.(*v1beta1.Application)
			if !ok {
				return fmt.Errorf("Failed to parse watch event")
			}

			if stop, err := appFunc(app); stop || err != nil {
				return err
			}
		case <-time.After(waitTime):
			if err := timeout(); err != nil {
				return err
			}
		}
	}
}

func Wait(kubeConf *rest.Config, repo string) error {
	timeout := func() error {
		return fmt.Errorf("Failed to become ready after 5 minutes, try running `plural watch %s` to get an idea where to debug", repo)
	}

	return Waiter(kubeConf, repo, func(app *v1beta1.Application) (bool, error) {
		tm.MoveCursor(1, 1)
		ready := Ready(app)
		Flush()
		return ready, nil
	}, timeout)
}
