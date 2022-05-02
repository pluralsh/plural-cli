package main

import (
	"os/exec"

	"github.com/pluralsh/plural/pkg/application"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
	"sigs.k8s.io/application/api/v1beta1"

	tm "github.com/buger/goterm"
)

func handleWatch(c *cli.Context) error {
	repo := c.Args().Get(0)
	kubeConf, err := utils.KubeConfig()
	if err != nil {
		return err
	}
	kube, err := utils.Kubernetes()
	if err != nil {
		return err
	}

	timeout := func() error { return nil }
	return application.Waiter(kubeConf, repo, func(app *v1beta1.Application) (bool, error) {
		tm.MoveCursor(1, 1)
		application.Print(kube.Kube, app)
		application.Flush()
		return false, nil
	}, timeout)
}

func handleWait(c *cli.Context) error {
	repo := c.Args().Get(0)
	kubeConf, err := utils.KubeConfig()
	if err != nil {
		return err
	}

	return application.Wait(kubeConf, repo)
}

func handleInfo(c *cli.Context) error {
	repo := c.Args().Get(0)
	conf := config.Read()
	cmd := exec.Command("k9s", "-n", conf.Namespace(repo))
	return cmd.Run()
}
