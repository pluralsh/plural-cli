package kubernetes

import (
	"context"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const defaultIngressClassAnnotation = "ingressclass.kubernetes.io/is-default-class"

var clusterIssuerGVR = schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "clusterissuers"}

// ClusterIssuers lists cert-manager cluster issuer names. It fails if cert-manager isn't installed.
func ClusterIssuers(dyn dynamic.Interface) ([]string, error) {
	list, err := dyn.Resource(clusterIssuerGVR).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		names = append(names, item.GetName())
	}
	sort.Strings(names)
	return names, nil
}

// IngressClasses lists ingress class names, with the cluster default first when one is annotated.
func IngressClasses(cs kubernetes.Interface) ([]string, error) {
	list, err := cs.NetworkingV1().IngressClasses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	def := ""
	names := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		if def == "" && item.Annotations[defaultIngressClassAnnotation] == "true" {
			def = item.Name
			continue
		}
		names = append(names, item.Name)
	}
	sort.Strings(names)

	if def != "" {
		names = append([]string{def}, names...)
	}
	return names, nil
}
