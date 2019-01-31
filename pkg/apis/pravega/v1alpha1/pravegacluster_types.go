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

const (
	// DefaultZookeeperUri is the default ZooKeeper URI in the form of "hostname:port"
	DefaultZookeeperUri = "zk-client:2181"

	// DefaultServiceType is the default service type for external access
	DefaultServiceType = v1.ServiceTypeLoadBalancer

	// DefaultPravegaVersion is the default tag used for for the Pravega
	// Docker image
	DefaultPravegaVersion = "0.3.2"
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

// WithDefaults set default values when not defined in the spec.
func (p *PravegaCluster) WithDefaults() {
	p.Spec.withDefaults(p)
}

// PravegaClusterSpec defines the desired state of PravegaCluster
type PravegaClusterSpec struct {
	// ZookeeperUri specifies the hostname/IP address and port in the format
	// "hostname:port".
	// By default, the value "zk-client:2181" is used, that corresponds to the
	// default Zookeeper service created by the Pravega Zookkeeper operator
	// available at: https://github.com/pravega/zookeeper-operator
	ZookeeperUri string `json:"zookeeperUri"`

	// ExternalAccess specifies whether or not to allow external access
	// to clients and the service type to use to achieve it
	// By default, external access is not enabled
	ExternalAccess *ExternalAccess `json:"externalAccess"`

	// Version is the expected version of the Pravega cluster.
	// The pravega-operator will eventually make the Pravega cluster version
	// equal to the expected version.
	//
	// The version must follow the [semver]( http://semver.org) format, for example "3.2.13".
	// Only Pravega released versions are supported: https://github.com/pravega/pravega/releases
	//
	// If version is not set, default is "0.3.2".
	Version string `json:"version"`

	// Bookkeeper configuration
	Bookkeeper *BookkeeperSpec `json:"bookkeeper"`

	// Pravega configuration
	Pravega *PravegaSpec `json:"pravega"`
}

func (s *PravegaClusterSpec) withDefaults(p *PravegaCluster) {
	if s.ZookeeperUri == "" {
		s.ZookeeperUri = DefaultZookeeperUri
	}

	if s.ExternalAccess == nil {
		s.ExternalAccess = &ExternalAccess{}
	}
	s.ExternalAccess.withDefaults()

	if s.Bookkeeper == nil {
		s.Bookkeeper = &BookkeeperSpec{}
	}
	s.Bookkeeper.withDefaults()

	if s.Pravega == nil {
		s.Pravega = &PravegaSpec{}
	}
	s.Pravega.withDefaults()
}

// ExternalAccess defines the configuration of the external access
type ExternalAccess struct {
	// Enabled specifies whether or not external access is enabled
	// By default, external access is not enabled
	Enabled bool `json:"enabled"`

	// Type specifies the service type to achieve external access.
	// Options are "LoadBalancer" and "NodePort".
	// By default, if external access is enabled, it will use "LoadBalancer"
	Type v1.ServiceType `json:"type,omitempty"`
}

func (e *ExternalAccess) withDefaults() {
	if e.Enabled == false {
		e.Type = ""
	} else if e.Enabled == true && e.Type == "" {
		e.Type = DefaultServiceType
	}
}

// ImageSpec defines the fields needed for a Docker repository image
type ImageSpec struct {
	Repository string        `json:"repository"`
	Tag        string        `json:"tag"`
	PullPolicy v1.PullPolicy `json:"pullPolicy"`
}

// PravegaClusterStatus defines the observed state of PravegaCluster
type PravegaClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
}
