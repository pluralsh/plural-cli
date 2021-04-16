package proxy

import (
	"fmt"
	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	"github.com/pluralsh/plural/pkg/utils"
)

func fetchSecret(namespace string, k *utils.Kube, creds *v1alpha1.Credentials) (string, error) {
	secret, err := k.Secret(namespace, creds.Secret)
	if err != nil {
		return "", err
	}

	val, ok := secret.Data[creds.Key]
	if !ok {
		return "", fmt.Errorf("Could not find credential key")
	}

	return string(val), nil
}
