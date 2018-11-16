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
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PravegaClusterList is the plural form of the Pravega cluster Kubernetes resource
type PravegaClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []PravegaCluster `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PravegaCluster is the type representing the Pravega cluster Kubernetes resource
type PravegaCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PravegaClusterSpec   `json:"spec"`
	Status            PravegaClusterStatus `json:"status,omitempty"`
}

// WithDefaults set default values when not defined in the spec.
func (p *PravegaCluster) WithDefaults() {
	p.Spec.withDefaults(p)
}

// PravegaClusterSpec is the Pravega cluster configuration
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

	// Bookkeeper configuration
	Bookkeeper BookkeeperSpec `json:"bookkeeper"`

	// Pravega configuration
	Pravega PravegaSpec `json:"pravega"`
}

func (s *PravegaClusterSpec) withDefaults(p *PravegaCluster) {
	if s.ZookeeperUri == "" {
		s.ZookeeperUri = DefaultZookeeperUri
	}

	if s.ExternalAccess == nil {
		externalAccess := ExternalAccess{}
		externalAccess.withDefaults()
		s.ExternalAccess = &externalAccess
	}

	s.Bookkeeper.withDefaults()

	s.Pravega.withDefaults()
}

// ExternalAccess defines the configuration of the external access
type ExternalAccess struct {
	// Enabled specifies whether or not external access is enabled
	Enabled bool `json:"enabled"`

	// Type specifies the service type to achieve external access.
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

type PravegaClusterStatus struct {
	// Fill me
}
