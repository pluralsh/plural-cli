package provider

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider/permissions"
	"github.com/pluralsh/plural-cli/pkg/provider/preflights"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ByokProvider struct {
	cluster string
	ctx     map[string]interface{}
	writer  manifest.Writer
}

func ByokFromManifest(man *manifest.ProjectManifest) (*ByokProvider, error) {
	prov := &ByokProvider{
		cluster: man.Cluster,
		ctx:     man.Context,
		writer:  nil,
	}

	return prov, nil
}

func mkBYOK(conf config.Config, name string) (prov *ByokProvider, err error) {
	prov = &ByokProvider{
		cluster: name,
		ctx:     map[string]interface{}{},
	}

	kubeconfigPath, err := askKubeconfig()
	if err != nil {
		return nil, err
	}

	// Expand tilde and resolve the path
	if strings.HasPrefix(kubeconfigPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		kubeconfigPath = filepath.Join(homeDir, kubeconfigPath[2:])
	} else if !filepath.IsAbs(kubeconfigPath) {
		// Convert relative path to absolute
		kubeconfigPath, err = filepath.Abs(kubeconfigPath)
		if err != nil {
			return nil, err
		}
	}
	kubeconfigData, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	// Convert to base64
	kubeconfigBase64 := base64.StdEncoding.EncodeToString(kubeconfigData)

	prov.ctx["kubeconfig"] = kubeconfigBase64

	projectManifest := manifest.ProjectManifest{
		Cluster:  name,
		Provider: api.BYOK,
		Owner:    &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint},
		Context:  prov.Context(),
	}
	prov.writer = projectManifest.Configure(cloudFlag, prov.Cluster())
	return prov, nil
}

func (b *ByokProvider) Name() string {
	return api.BYOK
}

func (b *ByokProvider) Cluster() string {
	return b.cluster
}

func (b *ByokProvider) Project() string {
	return ""
}

func (b *ByokProvider) Region() string {
	return ""
}

func (b *ByokProvider) Bucket() string {
	return ""
}

func (b *ByokProvider) KubeConfig() error {
	if kubernetes.InKubernetes() {
		return nil
	}

	// Decode the new kubeconfig
	newKubeconfigData, err := base64.StdEncoding.DecodeString(utils.ToString(b.ctx["kubeconfig"]))
	if err != nil {
		return err
	}

	// Parse the new kubeconfig
	newConfig, err := clientcmd.Load(newKubeconfigData)
	if err != nil {
		return fmt.Errorf("failed to parse new kubeconfig: %w", err)
	}

	// Get kubeconfig path
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	// Load existing kubeconfig if it exists
	existingConfig, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load existing kubeconfig: %w", err)
	}

	// If no existing config, use empty one
	if existingConfig == nil {
		existingConfig = clientcmdapi.NewConfig()
	}

	// Merge the new config into existing
	for name, cluster := range newConfig.Clusters {
		existingConfig.Clusters[name] = cluster
	}
	for name, authInfo := range newConfig.AuthInfos {
		existingConfig.AuthInfos[name] = authInfo
	}
	for name, context := range newConfig.Contexts {
		existingConfig.Contexts[name] = context
	}

	// Set the current context from the new config
	if newConfig.CurrentContext != "" {
		existingConfig.CurrentContext = newConfig.CurrentContext
	}

	// Write merged kubeconfig
	err = clientcmd.WriteToFile(*existingConfig, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig: %w", err)
	}

	return nil
}

func (b *ByokProvider) KubeContext() string {
	kubeconfigData, err := base64.StdEncoding.DecodeString(utils.ToString(b.ctx["kubeconfig"]))
	if err != nil {
		return ""
	}

	// Parse the kubeconfig
	load, err := clientcmd.Load(kubeconfigData)
	if err != nil {
		return ""
	}

	return load.CurrentContext
}

func (b *ByokProvider) CreateBucket() error {
	
	return nil
}

func (b *ByokProvider) Context() map[string]interface{} {
	return b.ctx
}

func (b *ByokProvider) Preflights() []*preflights.Preflight {
	return []*preflights.Preflight{
		{Name: "Test cluster connection", Callback: b.testClusterConnectivity},
	}
}

func (b *ByokProvider) Permissions() (permissions.Checker, error) {
	return permissions.NullChecker(), nil
}

func (b *ByokProvider) Flush() error {
	if b == nil || b.writer == nil {
		return nil
	}
	return b.writer()
}

func (b *ByokProvider) testClusterConnectivity() error {
	if err := b.KubeConfig(); err != nil {
		return err
	}
	kube, err := kubernetes.Kubernetes()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}
	if _, err := kube.Nodes(); err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	return nil
}

func askKubeconfig() (string, error) {
	location := "~/.kube/config"
	if err := survey.AskOne(&survey.Input{Message: "Enter the path to the kubeconfig file", Default: location}, &location); err != nil {
		return "", err
	}

	return location, nil
}
