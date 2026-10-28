package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClusterIssuers(t *testing.T) {
	issuer := func(name string) runtime.Object {
		u := &unstructured.Unstructured{}
		u.SetAPIVersion("cert-manager.io/v1")
		u.SetKind("ClusterIssuer")
		u.SetName(name)
		return u
	}
	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{clusterIssuerGVR: "ClusterIssuerList"},
		issuer("plural"), issuer("letsencrypt-prod"))

	names, err := ClusterIssuers(dyn)
	assert.NoError(t, err)
	assert.Equal(t, []string{"letsencrypt-prod", "plural"}, names)
}

func TestIngressClasses(t *testing.T) {
	class := func(name string, def bool) runtime.Object {
		c := &networkingv1.IngressClass{ObjectMeta: metav1.ObjectMeta{Name: name}}
		if def {
			c.Annotations = map[string]string{defaultIngressClassAnnotation: "true"}
		}
		return c
	}

	tests := []struct {
		name    string
		objects []runtime.Object
		want    []string
	}{
		{name: "none", want: []string{}},
		{name: "sorted without default", objects: []runtime.Object{class("traefik", false), class("nginx", false)}, want: []string{"nginx", "traefik"}},
		{name: "default first", objects: []runtime.Object{class("alb", false), class("nginx", false), class("traefik", true)}, want: []string{"traefik", "alb", "nginx"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := IngressClasses(fake.NewClientset(tt.objects...))
			assert.NoError(t, err)
			assert.Equal(t, tt.want, names)
		})
	}
}
