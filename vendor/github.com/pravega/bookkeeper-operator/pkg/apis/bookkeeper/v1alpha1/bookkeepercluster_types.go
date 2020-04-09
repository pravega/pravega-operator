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
	"github.com/pravega/pravega-operator/pkg/controller/config"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultZookeeperUri is the default ZooKeeper URI in the form of "hostname:port"
	DefaultZookeeperUri = "zk-client:2181"

	// DefaultServiceType is the default service type for external access
	DefaultServiceType = v1.ServiceTypeLoadBalancer

	// DefaultPravegaVersion is the default tag used for for the Pravega
	// Docker image
	DefaultBookkeeperVersion = "0.6.1"
	// DefaultBookkeeperImageRepository is the default Docker repository for

	// the BookKeeper image
	DefaultBookkeeperImageRepository = "pravega/bookkeeper"

	// DefaultbookkeeperImagePullPolicy is the default image pull policy used
	// for the Bookkeeper Docker image
	DefaultBookkeeperImagePullPolicy = v1.PullAlways

	// DefaultBookkeeperLedgerVolumeSize is the default volume size for the
	// Bookkeeper ledger volume
	DefaultBookkeeperLedgerVolumeSize = "10Gi"

	// DefaultBookkeeperJournalVolumeSize is the default volume size for the
	// Bookkeeper journal volume
	DefaultBookkeeperJournalVolumeSize = "10Gi"

	// DefaultBookkeeperIndexVolumeSize is the default volume size for the
	// Bookkeeper index volume
	DefaultBookkeeperIndexVolumeSize = "10Gi"

	// MinimumBookkeeperReplicas is the minimum number of Bookkeeper replicas
	// accepted
	MinimumBookkeeperReplicas = 3

	// DefaultBookkeeperRequestCPU is the default CPU request for BookKeeper
	DefaultBookkeeperRequestCPU = "500m"

	// DefaultBookkeeperLimitCPU is the default CPU limit for BookKeeper
	DefaultBookkeeperLimitCPU = "1"

	// DefaultBookkeeperRequestMemory is the default memory request for BookKeeper
	DefaultBookkeeperRequestMemory = "1Gi"

	// DefaultBookkeeperLimitMemory is the limit memory limit for BookKeeper
	DefaultBookkeeperLimitMemory = "2Gi"
)

func init() {
	SchemeBuilder.Register(&BookkeeperCluster{}, &BookkeeperClusterList{})
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BookkeeperClusterList contains a list of BookkeeperCluster
type BookkeeperClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BookkeeperCluster `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BookkeeperCluster is the Schema for the BookkeeperClusters API
// +k8s:openapi-gen=true
type BookkeeperCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BookkeeperClusterSpec   `json:"spec,omitempty"`
	Status BookkeeperClusterStatus `json:"status,omitempty"`
}

// WithDefaults set default values when not defined in the spec.
func (bk *BookkeeperCluster) WithDefaults() (changed bool) {
	changed = bk.Spec.withDefaults()
	return changed
}

// ClusterSpec defines the desired state of BookkeeperCluster
type BookkeeperClusterSpec struct {
	// ZookeeperUri specifies the hostname/IP address and port in the format
	// "hostname:port".
	// By default, the value "zk-client:2181" is used, that corresponds to the
	// default Zookeeper service created by the Pravega Zookkeeper operator
	// available at: https://github.com/pravega/zookeeper-operator
	ZookeeperUri string `json:"zookeeperUri"`

	// Image defines the BookKeeper Docker image to use.
	// By default, "pravega/bookkeeper" will be used.
	Image *BookkeeperImageSpec `json:"image"`

	// Replicas defines the number of BookKeeper replicas.
	// Minimum is 3. Defaults to 3.
	Replicas int32 `json:"replicas"`

	// Storage configures the storage for BookKeeper
	Storage *BookkeeperStorageSpec `json:"storage"`

	// AutoRecovery indicates whether or not BookKeeper auto recovery is enabled.
	// Defaults to true.
	AutoRecovery *bool `json:"autoRecovery"`

	// ServiceAccountName configures the service account used on BookKeeper instances
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// BookieResources specifies the request and limit of resources that bookie can have.
	// BookieResources includes CPU and memory resources
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Options is the Bookkeeper configuration that is to override the bk_server.conf
	// in bookkeeper. Some examples can be found here
	// https://github.com/apache/bookkeeper/blob/master/docker/README.md
	Options map[string]string `json:"options"`

	// JVM is the JVM options for bookkeeper. It will be passed to the JVM for performance tuning.
	// If this field is not specified, the operator will use a set of default
	// options that is good enough for general deployment.
	JVMOptions *JVMOptions `json:"jvmOptions"`

	// Provides the name of the configmap created by the user to provide additional key-value pairs
	// that need to be configured into the bookie pods as environmental variables
	EnvVars string `json:"envVars,omitempty"`

	// Version is the expected version of the Pravega cluster.
	// The pravega-operator will eventually make the Pravega cluster version
	// equal to the expected version.
	//
	// The version must follow the [semver]( http://semver.org) format, for example "3.2.13".
	// Only Pravega released versions are supported: https://github.com/pravega/pravega/releases
	//
	// If version is not set, default is "0.4.0".
	Version string `json:"version"`
}

// BookkeeperImageSpec defines the fields needed for a BookKeeper Docker image
type BookkeeperImageSpec struct {
	ImageSpec
}

func (s *BookkeeperImageSpec) withDefaults() (changed bool) {
	if s.Repository == "" {
		changed = true
		s.Repository = DefaultBookkeeperImageRepository
	}

	s.Tag = ""

	if s.PullPolicy == "" {
		changed = true
		s.PullPolicy = DefaultBookkeeperImagePullPolicy
	}

	return changed
}

type JVMOptions struct {
	MemoryOpts    []string `json:"memoryOpts"`
	GcOpts        []string `json:"gcOpts"`
	GcLoggingOpts []string `json:"gcLoggingOpts"`
	ExtraOpts     []string `json:"extraOpts"`
}

func (s *JVMOptions) withDefaults() (changed bool) {
	if s.MemoryOpts == nil {
		changed = true
		s.MemoryOpts = []string{}
	}

	if s.GcOpts == nil {
		changed = true
		s.GcOpts = []string{}
	}

	if s.GcLoggingOpts == nil {
		changed = true
		s.GcLoggingOpts = []string{}
	}

	if s.ExtraOpts == nil {
		changed = true
		s.ExtraOpts = []string{}
	}

	return changed
}

// BookkeeperStorageSpec is the configuration of the volumes used in BookKeeper
type BookkeeperStorageSpec struct {
	// LedgerVolumeClaimTemplate is the spec to describe PVC for the BookKeeper ledger
	// This field is optional. If no PVC spec and there is no default storage class,
	// stateful containers will use emptyDir as volume
	LedgerVolumeClaimTemplate *v1.PersistentVolumeClaimSpec `json:"ledgerVolumeClaimTemplate"`

	// JournalVolumeClaimTemplate is the spec to describe PVC for the BookKeeper journal
	// This field is optional. If no PVC spec and there is no default storage class,
	// stateful containers will use emptyDir as volume
	JournalVolumeClaimTemplate *v1.PersistentVolumeClaimSpec `json:"journalVolumeClaimTemplate"`

	// IndexVolumeClaimTemplate is the spec to describe PVC for the BookKeeper index
	// This field is optional. If no PVC spec and there is no default storage class,
	// stateful containers will use emptyDir as volume
	IndexVolumeClaimTemplate *v1.PersistentVolumeClaimSpec `json:"indexVolumeClaimTemplate"`
}

func (s *BookkeeperStorageSpec) withDefaults() (changed bool) {
	if s.LedgerVolumeClaimTemplate == nil {
		changed = true
		s.LedgerVolumeClaimTemplate = &v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(DefaultBookkeeperLedgerVolumeSize),
				},
			},
		}
	}

	if s.JournalVolumeClaimTemplate == nil {
		changed = true
		s.JournalVolumeClaimTemplate = &v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(DefaultBookkeeperJournalVolumeSize),
				},
			},
		}
	}

	if s.IndexVolumeClaimTemplate == nil {
		changed = true
		s.IndexVolumeClaimTemplate = &v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(DefaultBookkeeperIndexVolumeSize),
				},
			},
		}
	}
	return changed
}

func (s *BookkeeperClusterSpec) withDefaults() (changed bool) {
	if s.ZookeeperUri == "" {
		changed = true
		s.ZookeeperUri = DefaultZookeeperUri
	}

	if s.Image == nil {
		changed = true
		s.Image = &BookkeeperImageSpec{}
	}
	if s.Image.withDefaults() {
		changed = true
	}

	if !config.TestMode && s.Replicas < MinimumBookkeeperReplicas {
		changed = true
		s.Replicas = MinimumBookkeeperReplicas
	}

	if s.Storage == nil {
		changed = true
		s.Storage = &BookkeeperStorageSpec{}
	}
	if s.Storage.withDefaults() {
		changed = true
	}

	if s.AutoRecovery == nil {
		changed = true
		boolTrue := true
		s.AutoRecovery = &boolTrue
	}

	if s.Resources == nil {
		changed = true
		s.Resources = &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(DefaultBookkeeperRequestCPU),
				v1.ResourceMemory: resource.MustParse(DefaultBookkeeperRequestMemory),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(DefaultBookkeeperLimitCPU),
				v1.ResourceMemory: resource.MustParse(DefaultBookkeeperLimitMemory),
			},
		}
	}

	if s.Options == nil {
		s.Options = map[string]string{}
	}

	if s.JVMOptions == nil {
		changed = true
		s.JVMOptions = &JVMOptions{}
	}

	if s.JVMOptions.withDefaults() {
		changed = true
	}

	if s.Version == "" {
		s.Version = DefaultBookkeeperVersion
		changed = true
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
