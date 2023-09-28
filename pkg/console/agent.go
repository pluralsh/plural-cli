package console

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pluralsh/plural/pkg/helm"
	"helm.sh/helm/v3/pkg/action"

	"github.com/gofrs/flock"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
	"sigs.k8s.io/yaml"
)

var settings = cli.New()

const (
	repoName  = "deployop"
	repoUrl   = "https://pluralsh.github.io/deployment-operator"
	namespace = "console"
)

// getEnvVar gets value of environment variable, if it is not set then default value is returned instead.
func getEnvVar(name, defaultValue string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}

	return defaultValue
}

func InstallAgent(url, token string) error {
	helmConfig, err := helm.GetActionConfig(namespace)
	if err != nil {
		return nil
	}

	cp, err := action.NewInstall(helmConfig).ChartPathOptions.LocateChart(repoName, settings)
	if err != nil {
		return err
	}

	chart, err := loader.Load(cp)
	if err != nil {
		return err
	}

	histClient := action.NewHistory(helmConfig)
	histClient.Max = 5
	if _, err := histClient.Run(repoName); errors.Is(err, driver.ErrReleaseNotFound) {
		instClient := action.NewInstall(helmConfig)
		instClient.Namespace = namespace
		instClient.ReleaseName = repoName
		instClient.Timeout = time.Minute * 10

		_, err = instClient.Run(chart, map[string]interface{}{})
		return err
	}
	client := action.NewUpgrade(helmConfig)
	client.Namespace = namespace
	client.Timeout = time.Minute * 10
	_, err = client.Run(repoName, chart, map[string]interface{}{})
	return nil
}

func addAgentRepo() error {
	repoFile := getEnvVar("HELM_REPOSITORY_CONFIG", helmpath.ConfigPath("repositories.yaml"))
	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Acquire a file lock for process synchronization.
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
		defer func(fileLock *flock.Flock) {
			_ = fileLock.Unlock()
		}(fileLock)
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
		Name:                  repoName,
		URL:                   repoUrl,
		InsecureSkipTLSverify: true,
	}

	// If the repo exists do one of two things:
	// 1. If the configuration for the name is the same continue without error.
	// 2. When the config is different require --force-update.
	if f.Has(repoName) {
		return nil
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return err
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		return fmt.Errorf("looks like %q is not a valid chart repository or cannot be reached", repoUrl)
	}

	f.Update(&c)
	return f.WriteFile(repoFile, 0644)
}
