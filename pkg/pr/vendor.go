package pr

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/utils"
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

func applyVendoring(template *PrTemplate, ctx map[string]interface{}) error {
	if template == nil || template.Spec.Vendor == nil {
		return nil
	}

	if template.Spec.Vendor.Helm != nil {
		helmSpec := template.Spec.Vendor.Helm
		if dest, err := templateReplacement([]byte(helmSpec.Destination), ctx); err == nil {
			helmSpec.Destination = string(dest)
		}
		if chart, err := templateReplacement([]byte(helmSpec.Chart), ctx); err == nil {
			helmSpec.Chart = string(chart)
		}
		if version, err := templateReplacement([]byte(helmSpec.Version), ctx); err == nil {
			helmSpec.Version = string(version)
		}

		return downloadChart(helmSpec)
	}

	return nil
}

// downloadChart downloads a Helm chart tarball to the specified destination
func downloadChart(helmSpec *Helm) error {
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
	chart := helmSpec.Chart

	// Configure pull options
	client.RepoURL = helmSpec.URL
	if isOCIRegistry(helmSpec.URL) {
		client.RepoURL = ""
		chart = joinPreserve(helmSpec.URL, chart)
	}
	client.Version = helmSpec.Version
	client.Settings = settings
	client.Untar = true

	// Enable registry client if using OCI
	registryClient, err := registry.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}
	actionConfig.RegistryClient = registryClient

	// handle nested directories robustly
	if err := os.MkdirAll(helmSpec.Destination, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "helm-chart-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	client.DestDir = tempDir

	// remove the destination directory if it exists, to avoid conflicts and helm client error if exists
	_ = os.RemoveAll(client.DestDir)
	// pull the chart
	if _, err := client.Run(chart); err != nil {
		return fmt.Errorf("failed to pull chart: %w", err)
	}

	// copy from temp dir, this allows us to preserve additional files a repo has added to the chart
	if err := utils.CopyDir(tempDir, helmSpec.Destination); err != nil {
		return fmt.Errorf("failed to copy chart to destination directory: %w", err)
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
