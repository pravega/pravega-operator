package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PravegaClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []PravegaCluster `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PravegaCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PravegaClusterSpec   `json:"spec"`
	Status            PravegaClusterStatus `json:"status,omitempty"`
}

type PravegaClusterSpec struct {
	ZookeeperUri string         `json:",zookeeper_uri"`
	Bookkeeper   BookkeeperSpec `json:",bookkeeper"`
	Pravega      PravegaSpec    `json:",pravega"`
}

type ImageSpec struct {
	Repository string        `json:",repository"`
	Tag        string        `json:",tag"`
	PullPolicy v1.PullPolicy `json:",pullPolicy"`
}

func (spec *ImageSpec) String() string {
	return spec.Repository + ":" + spec.Tag
}

type PravegaClusterStatus struct {
	// Fill me
}
