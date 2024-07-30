package cd

import (
	"fmt"

	gqlclient "github.com/pluralsh/console/go/client"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func SaveClusterKubeconfig(cluster *gqlclient.ClusterFragment, token string) error {
	configAccess := clientcmd.NewDefaultPathOptions()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("cannot read kubeconfig: %w", err)
	}
	if config == nil {
		config = &clientcmdapi.Config{}
	}

	// TODO: We should additionally set CertificateAuthority for Cluster.
	configCluster := clientcmdapi.NewCluster()
	configCluster.Server = *cluster.KasURL
	if config.Clusters == nil {
		config.Clusters = make(map[string]*clientcmdapi.Cluster)
	}
	config.Clusters[cluster.Name] = configCluster

	configAuthInfo := clientcmdapi.NewAuthInfo()
	configAuthInfo.Token = fmt.Sprintf("plrl:%s:%s", cluster.ID, token)
	if config.AuthInfos == nil {
		config.AuthInfos = make(map[string]*clientcmdapi.AuthInfo)
	}
	config.AuthInfos[cluster.Name] = configAuthInfo

	configContext := clientcmdapi.NewContext()
	configContext.Cluster = cluster.Name
	configContext.AuthInfo = cluster.Name
	if config.Contexts == nil {
		config.Contexts = make(map[string]*clientcmdapi.Context)
	}
	config.Contexts[cluster.Name] = configContext

	config.CurrentContext = cluster.Name

	if err := clientcmd.ModifyConfig(configAccess, *config, true); err != nil {
		return err
	}

	fmt.Printf("set your kubectl context to %s\n", cluster.Name)
	return nil
}
