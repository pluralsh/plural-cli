package pr

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
)

func newEnvSettings() (*cli.EnvSettings, string, error) {
	dir, err := os.MkdirTemp("", "repositories")
	if err != nil {
		return nil, dir, err
	}

	settings := cli.New()
	settings.RepositoryCache = dir
	settings.RepositoryConfig = path.Join(dir, "repositories.yaml")
	settings.KubeInsecureSkipTLSVerify = true

	return settings, dir, nil
}

// downloadChart downloads a Helm chart tarball to the specified destination
func downloadChart(template *PrTemplate) error {
	if template == nil {
		return nil
	}
	if template.Spec.Vendor == nil {
		return nil
	}
	if template.Spec.Vendor.Helm == nil {
		return nil
	}

	// Create Helm environment settings
	settings, dir, err := newEnvSettings()
	if err != nil {
		return fmt.Errorf("failed to create helm environment settings: %w", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(dir)

	// Create action configuration
	actionConfig := new(action.Configuration)

	// Initialize with empty namespace
	if err := actionConfig.Init(settings.RESTClientGetter(), "", os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		fmt.Printf(format+"\n", v...)
	}); err != nil {
		return fmt.Errorf("failed to initialize action config: %w", err)
	}

	// Create pull action
	client := action.NewPullWithOpts(action.WithConfig(actionConfig))
	chart := template.Spec.Vendor.Helm.Chart
	dirPath := joinPreserve(template.Spec.Vendor.Helm.Destination, chart)
	// Configure pull options
	client.RepoURL = template.Spec.Vendor.Helm.URL
	if isOCIRegistry(template.Spec.Vendor.Helm.URL) {
		client.RepoURL = ""
		chart = joinPreserve(template.Spec.Vendor.Helm.URL, chart)
	}
	client.Version = template.Spec.Vendor.Helm.Version
	client.DestDir = template.Spec.Vendor.Helm.Destination
	client.Settings = settings
	client.Untar = true

	// Enable registry client if using OCI
	registryClient, err := registry.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}
	actionConfig.RegistryClient = registryClient

	// remove the destination directory if it exists, to avoid conflicts and helm client error if exists
	_ = os.RemoveAll(dirPath)
	// pull the chart
	// currently output is always empty, so we ignore it. We have to remove the tarball manually.
	if _, err := client.Run(chart); err != nil {
		return fmt.Errorf("failed to pull chart: %w", err)
	}
	return removeTarball(client.DestDir, chart)
}

func removeTarball(dir, prefix string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && (strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz")) {
			return os.Remove(filepath.Join(dir, name))
		}
	}
	return nil
}

func isOCIRegistry(url string) bool {
	return strings.HasPrefix(url, "oci://")
}

// JoinPreserve joins two path segments without normalizing them.
// Unlike filepath.Join or path.Join, this function preserves leading
// prefixes such as "./", "../", and URL schemes like "oci://". It
// simply concatenates the two strings with exactly one slash between
// them, making it safe for:
//   - relative filesystem paths  (e.g., "./dir/file")
//   - parent-relative paths      (e.g., "../dir/file")
//   - OCI / URL-like paths       (e.g., "oci://host/repo/tag")
//
// This is useful when the exact prefix format must remain unchanged.
func joinPreserve(base, sub string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(sub, "/")
}
