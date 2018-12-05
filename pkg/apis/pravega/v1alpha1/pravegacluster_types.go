/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Temporary solution, blocked by Operator SDK issue 727
const (
	APIVERSION = "pravega.pravega.io/v1alpha1"
	KIND       = "PravegaCluster"
)

func init() {
	SchemeBuilder.Register(&PravegaCluster{}, &PravegaClusterList{})
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PravegaClusterList contains a list of PravegaCluster
type PravegaClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PravegaCluster `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PravegaCluster is the Schema for the pravegaclusters API
// +k8s:openapi-gen=true
type PravegaCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PravegaClusterSpec   `json:"spec,omitempty"`
	Status PravegaClusterStatus `json:"status,omitempty"`
}

// PravegaClusterSpec defines the desired state of PravegaCluster
type PravegaClusterSpec struct {
	ZookeeperUri   string         `json:"zookeeperUri"`
	ExternalAccess ExternalAccess `json:"externalAccess"`
	Bookkeeper     BookkeeperSpec `json:"bookkeeper"`
	Pravega        PravegaSpec    `json:"pravega"`
}

type ExternalAccess struct {
	Enabled bool           `json:"enabled"`
	Type    v1.ServiceType `json:"type,omitempty"`
}

type ImageSpec struct {
	Repository string        `json:"repository"`
	Tag        string        `json:"tag"`
	PullPolicy v1.PullPolicy `json:"pullPolicy"`
}

func (spec *ImageSpec) String() string {
	return spec.Repository + ":" + spec.Tag
}

// PravegaClusterStatus defines the observed state of PravegaCluster
type PravegaClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}
