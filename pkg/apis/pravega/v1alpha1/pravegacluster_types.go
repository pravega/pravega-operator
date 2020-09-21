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
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// DefaultZookeeperUri is the default ZooKeeper URI in the form of "hostname:port"
	DefaultZookeeperUri = "zookeeper-client:2181"

	// DefaultServiceType is the default service type for external access
	DefaultServiceType = v1.ServiceTypeLoadBalancer

	// DefaultPravegaVersion is the default tag used for for the Pravega
	// Docker image
	DefaultPravegaVersion = "0.4.0"
)

func (*PravegaCluster) Hub() {}

func (p *PravegaCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	log.Print("SetupWebhookWithManager invoked")
	return ctrl.NewWebhookManagedBy(mgr).
		For(p).
		Complete()
}

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
// +k8s:openapi-gen=true

// Generate CRD using kubebuilder

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=pk
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.currentVersion`,description="The current pravega version"
// +kubebuilder:printcolumn:name="Desired Version",type=string,JSONPath=`.spec.version`,description="The desired pravega version"
// +kubebuilder:printcolumn:name="Desired Members",type=integer,JSONPath=`.status.replicas`,description="The number of desired pravega members"
// +kubebuilder:printcolumn:name="Ready Members",type=integer,JSONPath=`.status.readyReplicas`,description="The number of ready pravega members"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// PravegaCluster is the Schema for the pravegaclusters API
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
	// By default, the value "zookeeper-client:2181" is used, that corresponds to the
	// default Zookeeper service created by the Pravega Zookkeeper operator
	// available at: https://github.com/pravega/zookeeper-operator
	// +optional
	ZookeeperUri string `json:"zookeeperUri"`

	// ExternalAccess specifies whether or not to allow external access
	// to clients and the service type to use to achieve it
	// By default, external access is not enabled
	// +optional
	ExternalAccess *ExternalAccess `json:"externalAccess"`

	// TLS is the Pravega security configuration that is passed to the Pravega processes.
	// See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md
	TLS *TLSPolicy `json:"tls,omitempty"`

	// Authentication can be enabled for authorizing all communication from clients to controller and segment store
	// See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md
	Authentication *AuthenticationParameters `json:"authentication,omitempty"`

	// Version is the expected version of the Pravega cluster.
	// The pravega-operator will eventually make the Pravega cluster version
	// equal to the expected version.
	//
	// The version must follow the [semver]( http://semver.org) format, for example "3.2.13".
	// Only Pravega released versions are supported: https://github.com/pravega/pravega/releases
	//
	// If version is not set, default is "0.4.0".
	// +optional
	Version string `json:"version"`

	// Bookkeeper configuration
	// +optional
	Bookkeeper *BookkeeperSpec `json:"bookkeeper"`

	// Pravega configuration
	// +optional
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

	if s.TLS == nil {
		changed = true
		s.TLS = &TLSPolicy{
			Static: &StaticTLS{},
		}
	}

	if s.Authentication == nil {
		changed = true
		s.Authentication = &AuthenticationParameters{}
	}

	if s.Version == "" {
		s.Version = DefaultPravegaVersion
		changed = true
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
	// +optional
	Enabled bool `json:"enabled"`

	// Type specifies the service type to achieve external access.
	// Options are "LoadBalancer" and "NodePort".
	// By default, if external access is enabled, it will use "LoadBalancer"
	Type v1.ServiceType `json:"type,omitempty"`

	// Domain Name to be used for External Access
	// This value is ignored if External Access is disabled
	DomainName string `json:"domainName,omitempty"`
}

func (e *ExternalAccess) withDefaults() (changed bool) {
	if e.Enabled == false && (e.Type != "" || e.DomainName != "") {
		changed = true
		e.Type = ""
		e.DomainName = ""
	}
	return changed
}

type TLSPolicy struct {
	// Static TLS means keys/certs are generated by the user and passed to an operator.
	Static *StaticTLS `json:"static,omitempty"`
}

type StaticTLS struct {
	ControllerSecret   string `json:"controllerSecret,omitempty"`
	SegmentStoreSecret string `json:"segmentStoreSecret,omitempty"`
	CaBundle           string `json:"caBundle,omitempty"`
}

func (tp *TLSPolicy) IsSecureController() bool {
	if tp == nil || tp.Static == nil {
		return false
	}
	return len(tp.Static.ControllerSecret) != 0
}

func (tp *TLSPolicy) IsSecureSegmentStore() bool {
	if tp == nil || tp.Static == nil {
		return false
	}
	return len(tp.Static.SegmentStoreSecret) != 0
}

func (tp *TLSPolicy) IsCaBundlePresent() bool {
	if tp == nil || tp.Static == nil {
		return false
	}
	return len(tp.Static.CaBundle) != 0
}

type AuthenticationParameters struct {
	// Enabled specifies whether or not authentication is enabled
	// By default, authentication is not enabled
	// +optional
	Enabled bool `json:"enabled"`

	// name of Secret containing Password based Authentication Parameters like username, password and acl
	// optional - used only by PasswordAuthHandler for authentication
	PasswordAuthSecret string `json:"passwordAuthSecret,omitempty"`
}

func (ap *AuthenticationParameters) IsEnabled() bool {
	if ap == nil {
		return false
	}
	return ap.Enabled
}

// ImageSpec defines the fields needed for a Docker repository image
type ImageSpec struct {
	// +optional
	Repository string `json:"repository"`

	// Deprecated: Use `spec.Version` instead
	Tag string `json:"tag,omitempty"`

	// +optional
	PullPolicy v1.PullPolicy `json:"pullPolicy"`
}
