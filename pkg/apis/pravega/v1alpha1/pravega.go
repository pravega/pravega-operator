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
)

const (
	// DefaultPravegaImageRepository is the default Docker repository for
	// the Pravega image
	DefaultPravegaImageRepository = "pravega/pravega"

	// DefaultPravegaImagePullPolicy is the default image pull policy used
	// for the Pravega Docker image
	DefaultPravegaImagePullPolicy = v1.PullAlways

	// DefaultPravegaCacheVolumeSize is the default volume size for the
	// Pravega SegmentStore cache volume
	DefaultPravegaCacheVolumeSize = "20Gi"

	// DefaultPravegaTier2ClaimName is the default volume claim name used as Tier 2
	DefaultPravegaTier2ClaimName = "pravega-tier2"

	// DefaultControllerReplicas is the default number of replicas for the Pravega
	// Controller component
	DefaultControllerReplicas = 1

	// DefaultSegmentStoreReplicas is the default number of replicas for the Pravega
	// Segment Store component
	DefaultSegmentStoreReplicas = 1

	// DefaultControllerRequestCPU is the default CPU request for Pravega
	DefaultControllerRequestCPU = "250m"

	// DefaultControllerLimitCPU is the default CPU limit for Pravega
	DefaultControllerLimitCPU = "500m"

	// DefaultControllerRequestMemory is the default memory request for Pravega
	DefaultControllerRequestMemory = "512Mi"

	// DefaultControllerLimitMemory is the default memory limit for Pravega
	DefaultControllerLimitMemory = "1Gi"

	// DefaultSegmentStoreRequestCPU is the default CPU request for Pravega
	DefaultSegmentStoreRequestCPU = "500m"

	// DefaultSegmentStoreLimitCPU is the default CPU limit for Pravega
	DefaultSegmentStoreLimitCPU = "1"

	// DefaultSegmentStoreRequestMemory is the default memory request for Pravega
	DefaultSegmentStoreRequestMemory = "1Gi"

	// DefaultSegmentStoreLimitMemory is the default memory limit for Pravega
	DefaultSegmentStoreLimitMemory = "2Gi"
)

// ExternalAccess defines the configuration of the external access
type ExternalAccess struct {
	// Type specifies the service type to achieve external access.
	// Options are "LoadBalancer" and "NodePort".
	// By default, if external access is enabled, it will use "LoadBalancer"
	Type v1.ServiceType `json:"type,omitempty"`

	// Domain Name to be used for External Access
	// If empty no dns name is added for
	DomainName string `json:"domainName,omitempty"`

	// Annotations to be added to the external service
	Annotations map[string]string `json:"annotations"`
}

func (e *ExternalAccess) withDefaults() (changed bool) {
	if e.Type == "" {
		changed = true
		e.Type = DefaultServiceType
	}

	if e.Annotations == nil {
		changed = true
		e.Annotations = map[string]string{}
	}

	return changed
}

type ControllerSpec struct {
	// ControllerReplicas defines the number of Controller replicas.
	// Defaults to 1.
	Replicas int32 `json:"replicas"`

	// Pravega configuration options passed to the Pravega processes
	// a '-D' is prefixed to the provided values to pass them as java system properties
	// See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/config/config.properties
	Options map[string]string `json:"options"`

	// DebugLogging indicates whether or not debug level logging is enabled.
	// Defaults to false.
	DebugLogging bool `json:"debugLogging"`

	// JVM Options for tuning JVM of Controller process.
	// These typically start with '-X' or '-XX'
	// See the following link for a complete list of options
	// https://docs.oracle.com/javase/8/docs/technotes/tools/unix/java.html
	JVMOptions []string `json:"jvmOptions"`

	// ControllerServiceAccountName configures the service account used on controller instances.
	// If not specified, Kubernetes will automatically assign the default service account in the namespace
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// ControllerResources specifies the request and limit of resources that controller can have.
	// ControllerResources includes CPU and memory resources
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Configuration to be used when external access is enabled for this component
	ExternalAccess *ExternalAccess `json:"externalAccess,omitempty"`
}

type SegmentStoreSpec struct {
	// SegmentStoreReplicas defines the number of Segment Store replicas.
	// Defaults to 1.
	Replicas int32 `json:"replicas"`

	// DebugLogging indicates whether or not debug level logging is enabled.
	// Defaults to false.
	DebugLogging bool `json:"debugLogging"`

	// Options is the Pravega configuration that is passed to the Pravega processes
	// as JAVA_OPTS. See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/config/config.properties
	Options map[string]string `json:"options"`

	// JVM Options for tuning JVM of Controller process.
	// These typically start with '-X' or '-XX'
	// See the following link for a complete list of options
	// https://docs.oracle.com/javase/8/docs/technotes/tools/unix/java.html
	JVMOptions []string `json:"jvmOptions"`

	// SegmentStoreServiceAccountName configures the service account used on segment store instances.
	// If not specified, Kubernetes will automatically assign the default service account in the namespace
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// SegmentStoreResources specifies the request and limit of resources that segmentStore can have.
	// SegmentStoreResources includes CPU and memory resources
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Configuration to be used when external access is enabled for this component
	ExternalAccess *ExternalAccess `json:"externalAccess,omitempty"`

	// CacheVolumeClaimTemplate is the spec to describe PVC for the Pravega cache.
	// This field is optional. If no PVC spec, stateful containers will use
	// emptyDir as volume
	CacheVolumeClaimTemplate *v1.PersistentVolumeClaimSpec `json:"cacheVolumeClaimTemplate"`
}

// PravegaSpec defines the configuration of Pravega
type PravegaSpec struct {
	// Image defines the Pravega Docker image to use.
	// By default, "pravega/pravega" will be used.
	Image *PravegaImageSpec `json:"image"`

	// Controller confugration
	Controller *ControllerSpec `json:"controller"`

	//SegmentStore configuration
	SegmentStore *SegmentStoreSpec `json:"segmentstore"`

	// Tier2 is the configuration of Pravega's tier 2 storage. If no configuration
	// is provided, it will assume that a PersistentVolumeClaim called "pravega-tier2"
	// is present and it will use it as Tier 2
	Tier2 *Tier2Spec `json:"tier2"`
}

func (s *PravegaSpec) withDefaults() (changed bool) {

	if s.Controller == nil {
		changed = true
		s.Controller = &ControllerSpec{}
	}

	if s.Controller.withDefaults() {
		changed = true
	}

	if s.SegmentStore == nil {
		changed = true
		s.SegmentStore = &SegmentStoreSpec{}
	}

	if s.SegmentStore.withDefaults() {
		changed = true
	}

	if s.Image == nil {
		changed = true
		s.Image = &PravegaImageSpec{}
	}

	if s.Image.withDefaults() {
		changed = true
	}

	if s.Tier2 == nil {
		changed = true
		s.Tier2 = &Tier2Spec{}
	}

	if s.Tier2.withDefaults() {
		changed = true
	}

	return changed
}

func (ss *SegmentStoreSpec) withDefaults() (changed bool) {
	if !config.TestMode && ss.Replicas < 1 {
		changed = true
		ss.Replicas = 1
	}

	if ss.Options == nil {
		changed = true
		ss.Options = map[string]string{}
	}

	if ss.JVMOptions == nil {
		changed = true
		ss.JVMOptions = make([]string, 0)
	}

	if ss.Resources == nil {
		changed = true
		ss.Resources = &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(DefaultSegmentStoreRequestCPU),
				v1.ResourceMemory: resource.MustParse(DefaultSegmentStoreRequestMemory),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(DefaultSegmentStoreLimitCPU),
				v1.ResourceMemory: resource.MustParse(DefaultSegmentStoreLimitMemory),
			},
		}
	}

	if ss.ExternalAccess == nil {
		changed = true
		ss.ExternalAccess = &ExternalAccess{}
	}

	if ss.ExternalAccess.withDefaults() {
		changed = true
	}

	if ss.CacheVolumeClaimTemplate == nil {
		changed = true
		ss.CacheVolumeClaimTemplate = &v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(DefaultPravegaCacheVolumeSize),
				},
			},
		}
	}

	return changed
}

func (cs *ControllerSpec) withDefaults() (changed bool) {
	if !config.TestMode && cs.Replicas < 1 {
		changed = true
		cs.Replicas = 1
	}

	if cs.Options == nil {
		changed = true
		cs.Options = map[string]string{}
	}

	if cs.JVMOptions == nil {
		changed = true
		cs.JVMOptions = make([]string, 0)
	}

	if cs.Resources == nil {
		changed = true
		cs.Resources = &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(DefaultControllerRequestCPU),
				v1.ResourceMemory: resource.MustParse(DefaultControllerRequestMemory),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(DefaultControllerLimitCPU),
				v1.ResourceMemory: resource.MustParse(DefaultControllerLimitMemory),
			},
		}
	}

	if cs.ExternalAccess == nil {
		changed = true
		cs.ExternalAccess = &ExternalAccess{}
	}

	if cs.ExternalAccess.withDefaults() {
		changed = true
	}

	return changed
}

// PravegaImageSpec defines the fields needed for a Pravega Docker image
type PravegaImageSpec struct {
	ImageSpec
}

func (s *PravegaImageSpec) withDefaults() (changed bool) {
	if s.Repository == "" {
		changed = true
		s.Repository = DefaultPravegaImageRepository
	}

	s.Tag = ""

	if s.PullPolicy == "" {
		changed = true
		s.PullPolicy = DefaultPravegaImagePullPolicy
	}

	return changed
}

// Tier2Spec configures the Tier 2 storage type to use with Pravega.
// If not specified, Tier 2 will be configured in filesystem mode and will try
// to use a PersistentVolumeClaim with the name "pravega-tier2"
type Tier2Spec struct {
	// FileSystem is used to configure a pre-created Persistent Volume Claim
	// as Tier 2 backend.
	// It is default Tier 2 mode.
	FileSystem *FileSystemSpec `json:"filesystem,omitempty"`

	// Ecs is used to configure a Dell EMC ECS system as a Tier 2 backend
	Ecs *ECSSpec `json:"ecs,omitempty"`

	// Hdfs is used to configure an HDFS system as a Tier 2 backend
	Hdfs *HDFSSpec `json:"hdfs,omitempty"`
}

func (s *Tier2Spec) withDefaults() (changed bool) {
	if s.FileSystem == nil && s.Ecs == nil && s.Hdfs == nil {
		changed = true
		fs := &FileSystemSpec{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: DefaultPravegaTier2ClaimName,
			},
		}
		s.FileSystem = fs
	}

	return changed
}

// FileSystemSpec contains the reference to a PVC.
type FileSystemSpec struct {
	PersistentVolumeClaim *v1.PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim"`
}

// ECSSpec contains the connection details to a Dell EMC ECS system
type ECSSpec struct {
	Uri         string `json:"uri"`
	Bucket      string `json:"bucket"`
	Root        string `json:"root"`
	Namespace   string `json:"namespace"`
	Credentials string `json:"credentials"`
}

// HDFSSpec contains the connection details to an HDFS system
type HDFSSpec struct {
	Uri               string `json:"uri"`
	Root              string `json:"root"`
	ReplicationFactor int32  `json:"replicationFactor"`
}
