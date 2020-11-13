package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate controller-gen object paths=$GOFILE

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Proxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProxySpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Proxy `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ProxySpec struct {
	metav1.TypeMeta `json:",inline"`
	Type            string      `json:"type"`
	Target          string      `json:"target"`
	Credentials     Credentials `json:"credentials"`
	DbConfig        DbConfig    `json:"dbConfig"`
	ShConfig        ShConfig    `json:"shConfig"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Credentials struct {
	metav1.TypeMeta `json:",inline"`
	Secret          string `json:"secret"`
	Key             string `json:"key"`
	User            string `json:"user"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DbConfig struct {
	metav1.TypeMeta `json:",inline"`
	Name            string `json:"name"`
	Engine          string `json:"engine"`
	Port            int32  `json:"port"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ShConfig struct {
	metav1.TypeMeta `json:",inline"`
	Command         string   `json:"command"`
	Args            []string `json:"args"`
}
