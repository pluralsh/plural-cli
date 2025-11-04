package gcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"cloud.google.com/go/container/apiv1/containerpb"
	"cloud.google.com/go/storage"
	"github.com/samber/lo"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider/permissions"
	"github.com/pluralsh/plural-cli/pkg/provider/preflights"
	utilerr "github.com/pluralsh/plural-cli/pkg/utils/errors"
)

type Provider struct {
	// InputProvider partially implements the Provider interface.
	InputProvider

	bucket string
	ctx    map[string]interface{}
	writer manifest.Writer
}

func (in *Provider) Name() string {
	return api.ProviderGCP
}

func (in *Provider) project() (string, error) {
	projectID := in.InputProvider.Project()

	exists, err := IsProjectExists(projectID)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", fmt.Errorf("project %s does not exist", projectID)
	}

	return projectID, nil
}

func (in *Provider) Bucket() string {
	return in.bucket
}

func (in *Provider) KubeConfig() error {
	if kubernetes.InKubernetes() {
		return nil
	}

	return in.setupKubeconfig()
}

func (in *Provider) setupKubeconfig() error {
	ctx := context.Background()

	c, err := ClusterManagerClient()
	if err != nil {
		return err
	}

	project, err := in.project()
	if err != nil {
		return err
	}

	// Determine cluster location (zone or region)
	//_, location := in.clusterLocation()

	// Get cluster details
	req := &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s",
			project, in.Region(), in.Cluster()),
	}

	cluster, err := c.GetCluster(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	// Decode cluster CA certificate
	caCert, err := base64.StdEncoding.DecodeString(cluster.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return fmt.Errorf("failed to decode CA certificate: %w", err)
	}

	// Build kubeconfig with gke-gcloud-auth-plugin
	contextName := in.KubeContext()

	kubeConfig := clientcmdapi.NewConfig()
	kubeConfig.Clusters[contextName] = &clientcmdapi.Cluster{
		Server:                   fmt.Sprintf("https://%s", cluster.Endpoint),
		CertificateAuthorityData: caCert,
	}
	kubeConfig.AuthInfos[contextName] = &clientcmdapi.AuthInfo{
		Exec: &clientcmdapi.ExecConfig{
			APIVersion:         "client.authentication.k8s.io/v1beta1",
			Command:            "gke-gcloud-auth-plugin",
			InstallHint:        "Install gke-gcloud-auth-plugin for use with kubectl by following https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl#install_plugin",
			ProvideClusterInfo: true,
		},
	}
	kubeConfig.Contexts[contextName] = &clientcmdapi.Context{
		Cluster:  contextName,
		AuthInfo: contextName,
	}
	kubeConfig.CurrentContext = contextName

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
	loadingRules := &clientcmd.ClientConfigLoadingRules{
		Precedence: []string{kubeconfigPath},
	}
	existingConfig, err := loadingRules.Load()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load existing kubeconfig: %w", err)
	}

	if existingConfig != nil {
		// Merge with existing config
		for k, v := range kubeConfig.Clusters {
			existingConfig.Clusters[k] = v
		}
		for k, v := range kubeConfig.AuthInfos {
			existingConfig.AuthInfos[k] = v
		}
		for k, v := range kubeConfig.Contexts {
			existingConfig.Contexts[k] = v
		}
		existingConfig.CurrentContext = contextName
		kubeConfig = existingConfig
	}

	// Write kubeconfig
	err = clientcmd.WriteToFile(*kubeConfig, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig: %w", err)
	}

	return nil
}

func (in *Provider) KubeContext() string {
	return fmt.Sprintf("gke_%s_%s_%s", in.Project(), in.Region(), in.Cluster())
}

func (in *Provider) CreateBucket() error {
	return utilerr.ErrorWrap(in.createBucket(in.bucket), fmt.Sprintf("Failed to create terraform state bucket %s", in.Bucket()))
}

func (in *Provider) Context() map[string]interface{} {
	return in.ctx
}

func (in *Provider) Preflights() []*preflights.Preflight {
	return []*preflights.Preflight{
		{Name: string(PreflightCheckEnabledServices), Callback: in.validateEnabled},
		{Name: string(PreflightCheckServiceAccountPermissions), Callback: in.validatePermissions},
	}
}

func (in *Provider) Permissions() (permissions.Checker, error) {
	projectID, err := in.project()
	if err != nil {
		return nil, err
	}

	return permissions.NewGcpChecker(context.Background(), projectID)
}

func (in *Provider) Flush() error {
	if in.writer == nil {
		return nil
	}
	return in.writer()
}

func (in *Provider) createBucket(name string) error {
	storageClient, err := StorageClient()
	if err != nil {
		return err
	}

	bkt := storageClient.Bucket(name)
	if _, err := bkt.Attrs(context.Background()); err != nil {
		return bkt.Create(context.Background(), in.Project(), &storage.BucketAttrs{
			Location: string(getBucketLocation(in.Region())),
		})
	}

	return nil
}

func (in *Provider) clusterLocation() (string, string) {
	if z, ok := in.ctx["clusterZone"]; ok {
		return "zone", z.(string)
	}

	return "region", in.Region()
}

func (in *Provider) ensure() (*Provider, error) {
	return in, lo.Ternary(
		len(in.bucket) == 0 || in.InputProvider == nil,
		fmt.Errorf("GCPProvider not initialized, NewGCPProvider(...) must be called with either a manifest or config option"),
		nil,
	)
}

func NewProvider(options ...Option) (*Provider, error) {
	result := new(Provider)

	err := printUserInfo()
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		if err := opt(result); err != nil {
			return nil, err
		}
	}

	return result.ensure()
}
