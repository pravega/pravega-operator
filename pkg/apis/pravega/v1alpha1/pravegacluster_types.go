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
	DefaultPravegaVersion = "0.4.0"
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

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// WithDefaults set default values when not defined in the spec.
func (p *PravegaCluster) WithDefaults() (changed bool) {
	changed = p.Spec.withDefaults()

	return changed
}

// ClusterSpec defines the desired state of PravegaCluster
type ClusterSpec struct {
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
	// If version is not set, default is "0.4.0".
	Version string `json:"version"`

	// Bookkeeper configuration
	Bookkeeper *BookkeeperSpec `json:"bookkeeper"`

	// Pravega configuration
	Pravega *PravegaSpec `json:"pravega"`
}

func (s *ClusterSpec) withDefaults() (changed bool) {
	if s.ZookeeperUri == "" {
		changed = true
		s.ZookeeperUri = DefaultZookeeperUri
	}

	if s.ExternalAccess == nil {
		changed = true
		s.ExternalAccess = &ExternalAccess{}
	}
	if s.ExternalAccess.withDefaults() {
		changed = true
	}

	if s.Version == "" {
		changed = true
		if s.Pravega != nil && s.Pravega.Image != nil && s.Pravega.Image.Tag != "" {
			s.Version = s.Pravega.Image.Tag
		} else {
			s.Version = DefaultPravegaVersion
		}
	}

	if s.Bookkeeper == nil {
		changed = true
		s.Bookkeeper = &BookkeeperSpec{}
	}
	if s.Bookkeeper.withDefaults() {
		changed = true
	}

	if s.Pravega == nil {
		changed = true
		s.Pravega = &PravegaSpec{}
	}
	if s.Pravega.withDefaults() {
		changed = true
	}

	return changed
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

func (e *ExternalAccess) withDefaults() (changed bool) {
	if e.Enabled == false && e.Type != "" {
		changed = true
		e.Type = ""
	} else if e.Enabled == true && e.Type == "" {
		changed = true
		e.Type = DefaultServiceType
	}

	return changed
}

// ImageSpec defines the fields needed for a Docker repository image
type ImageSpec struct {
	Repository string `json:"repository"`

	// Deprecated: Use `spec.Version` instead
	Tag string `json:"tag,omitempty"`

	PullPolicy v1.PullPolicy `json:"pullPolicy"`
}
