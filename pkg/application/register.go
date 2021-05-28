package application

import (
	"sigs.k8s.io/application/api/v1beta1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
)

var (
    SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
    AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
    scheme.AddKnownTypes(v1beta1.GroupVersion,
        &v1beta1.Application{},
        &v1beta1.ApplicationList{},
    )

    metav1.AddToGroupVersion(scheme, v1beta1.GroupVersion)
    return nil
}