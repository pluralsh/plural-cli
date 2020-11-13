package utils

import (
	"context"
	"os"
	"path/filepath"

	"github.com/michaeljguarino/forge/clientset/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

const tokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"

func InKubernetes() bool {
	return Exists(tokenFile)
}

type Kube struct {
	Kube  *kubernetes.Clientset
	Forge v1alpha1.Clientset
}

func Kubernetes() (*Kube, error) {
	homedir, _ := os.UserHomeDir()
	conf := filepath.Join(homedir, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", conf)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	forgeclient, err := v1alpha1.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Kube{Kube: clientset, Forge: forgeclient}, nil
}

func (k *Kube) Secret(namespace string, name string) (*v1.Secret, error) {
	return k.Kube.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
}
