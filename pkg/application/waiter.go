package application

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/application/api/v1beta1"
)

func ListAll(kubeConf *rest.Config) ([]v1beta1.Application, error) {
	apps, err := NewForConfig(kubeConf)
	if err != nil {
		return nil, err
	}

	client := apps.Applications("")
	l, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return l.Items, nil
}
