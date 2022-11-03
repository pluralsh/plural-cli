package proxy

import (
	"fmt"

	"github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	"github.com/pluralsh/plural/pkg/kubernetes"

	v1 "k8s.io/api/core/v1"
)

type UserCredentials struct {
	User     string
	Password string
}

func fetchSecret(namespace string, k kubernetes.Kube, creds *v1alpha1.Credentials) (string, error) {
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

func fetchUserPassword(secret *v1.Secret, creds *v1alpha1.Credentials) (user *UserCredentials, err error) {
	pwd, ok := secret.Data[creds.Key]
	if !ok {
		err = fmt.Errorf("Could not find password key")
		return
	}

	username := creds.User
	if creds.UserKey != "" {
		uname, ok := secret.Data[creds.UserKey]
		if !ok {
			err = fmt.Errorf("Could not find password key")
			return
		}
		username = string(uname)
	}

	if username == "" {
		err = fmt.Errorf("No username found")
		return
	}

	user = &UserCredentials{User: username, Password: string(pwd)}
	return
}
