package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/pluralsh/plural/pkg/helm"
	"github.com/pluralsh/plural/pkg/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

var settings = cli.New()

func InstallCilium(cluster string) error {
	namespace := "kube-system"
	cmd := exec.Command(
		"kind", "export", "kubeconfig", "--name", cluster)
	if err := utils.Execute(cmd); err != nil {
		return err
	}

	if err := addCiliumRepo(); err != nil {
		return err
	}

	helmConfig, err := helm.GetActionConfig(namespace)
	if err != nil {
		return nil
	}
	instClient := action.NewInstall(helmConfig)

	cp, err := instClient.ChartPathOptions.LocateChart("cilium/cilium", settings)
	if err != nil {
		return err
	}
	chart, err := loader.Load(cp)
	if err != nil {
		return err
	}

	instClient.Namespace = namespace
	instClient.ReleaseName = "cilium"
	instClient.Timeout = time.Minute * 10

	_, err = instClient.Run(chart, map[string]interface{}{})

	return err
}

func addCiliumRepo() error {
	name := "cilium"
	url := "https://helm.cilium.io/"
	repoFile := envOr("HELM_REPOSITORY_CONFIG", helmpath.ConfigPath("repositories.yaml"))

	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Acquire a file lock for process synchronization
	repoFileExt := filepath.Ext(repoFile)
	var lockPath string
	if len(repoFileExt) > 0 && len(repoFileExt) < len(repoFile) {
		lockPath = strings.TrimSuffix(repoFile, repoFileExt) + ".lock"
	} else {
		lockPath = repoFile + ".lock"
	}
	fileLock := flock.New(lockPath)
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		return err
	}

	b, err := os.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}

	c := repo.Entry{
		Name:                  name,
		URL:                   url,
		InsecureSkipTLSverify: true,
	}

	// If the repo exists do one of two things:
	// 1. If the configuration for the name is the same continue without error
	// 2. When the config is different require --force-update
	if f.Has(name) {
		return nil
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return err
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		return fmt.Errorf("looks like %q is not a valid chart repository or cannot be reached", url)
	}

	f.Update(&c)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		return err
	}

	return nil
}

func envOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}
