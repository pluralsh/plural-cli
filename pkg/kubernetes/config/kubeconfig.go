package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural/pkg/utils"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func GetKubeconfig(path, context string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	if !utils.Exists(path) {
		return "", fmt.Errorf("the specified path does not exist")
	}

	config, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return "", err
	}

	if context != "" {
		if config.Contexts[context] == nil {
			return "", fmt.Errorf("the given context doesn't exist")
		}
		config.CurrentContext = context
	}
	newConfig := *clientcmdapi.NewConfig()
	newConfig.CurrentContext = config.CurrentContext
	newConfig.Contexts[config.CurrentContext] = config.Contexts[config.CurrentContext]
	newConfig.Clusters[config.CurrentContext] = config.Clusters[config.CurrentContext]
	newConfig.AuthInfos[config.CurrentContext] = config.AuthInfos[config.CurrentContext]
	newConfig.Extensions[config.CurrentContext] = config.Extensions[config.CurrentContext]
	newConfig.Preferences = config.Preferences
	result, err := clientcmd.Write(newConfig)
	if err != nil {
		return "", err
	}

	return string(result), nil
}
