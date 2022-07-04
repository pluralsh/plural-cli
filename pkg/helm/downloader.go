package helm

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"os"
)

var providers = []getter.Provider{ChartMuseumProvider}

func UpdateDependencies(path string) error {
	out := os.Stdout
	client := action.NewDependency()
	settings := cli.New()

	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptWriter(out),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)
	if err != nil {
		return err
	}

	gtrs := getter.All(settings)
	man := &downloader.Manager{
		Out:              out,
		ChartPath:        path,
		Keyring:          client.Keyring,
		SkipUpdate:       false,
		Getters:          append(providers, gtrs...),
		RegistryClient:   registryClient,
		RepositoryConfig: settings.RepositoryConfig,
		RepositoryCache:  settings.RepositoryCache,
		Debug:            settings.Debug,
	}
	return man.Update()
}