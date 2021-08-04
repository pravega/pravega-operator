/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1beta1

import (
	"github.com/pravega/pravega-operator/pkg/controller/config"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
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

	// DefaultPravegaLTSClaimName is the default volume claim name used as Tier 2
	DefaultPravegaLTSClaimName = "pravega-tier2"

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

	// DefaultInfluxDBSecretMountDir is the default mountpath of influxdb secret
	DefaultInfluxDBSecretMountDir = "/etc/influxdb-secret-volume"
)

// PravegaSpec defines the configuration of Pravega
type PravegaSpec struct {
	// ControllerReplicas defines the number of Controller replicas.
	// Defaults to 1.
	// +kubebuilder:validation:Minimum=0
	// +optional
	ControllerReplicas int32 `json:"controllerReplicas"`

	// SegmentStoreReplicas defines the number of Segment Store replicas.
	// Defaults to 1.
	// +kubebuilder:validation:Minimum=0
	// +optional
	SegmentStoreReplicas int32 `json:"segmentStoreReplicas"`

	// MaxUnavailableSegmentStoreReplicas defines the
	// MaxUnavailable SegmentStore Replicas
	// Default is 1.
	// +optional
	MaxUnavailableSegmentStoreReplicas int32 `json:"maxUnavailableSegmentStoreReplicas"`

	// MaxUnavailableControllerReplicas defines the
	// MaxUnavailable Controller Replicas
	// Default is 1.
	// +optional
	MaxUnavailableControllerReplicas int32 `json:"maxUnavailableControllerReplicas"`

	// DebugLogging indicates whether or not debug level logging is enabled.
	// Defaults to false.
	// +optional
	DebugLogging bool `json:"debugLogging"`

	// Image defines the Pravega Docker image to use.
	// By default, "pravega/pravega" will be used.
	// +optional
	Image *ImageSpec `json:"image"`

	// Options is the Pravega configuration that is passed to the Pravega processes
	// as JAVA_OPTS. See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/config/config.properties
	// +optional
	Options map[string]string `json:"options"`

	// ControllerJvmOptions is the JVM options for controller. It will be passed to the JVM
	// for performance tuning. If this field is not specified, the operator will use a set of default
	// options that is good enough for general deployment.
	// +optional
	ControllerJvmOptions []string `json:"controllerjvmOptions"`

	// SegmentStoreJVMOptions is the JVM options for Segmentstore. It will be passed to the JVM
	// for performance tuning. If this field is not specified, the operator will use a set of default
	// options that is good enough for general deployment.
	// +optional
	SegmentStoreJVMOptions []string `json:"segmentStoreJVMOptions"`

	// CacheVolumeClaimTemplate is the spec to describe PVC for the Pravega cache.
	// This field is optional. If no PVC spec, stateful containers will use
	// emptyDir as volume
	// +optional
	CacheVolumeClaimTemplate *v1.PersistentVolumeClaimSpec `json:"cacheVolumeClaimTemplate,omitempty"`

	// LongTermStorage is the configuration of Pravega's tier 2 storage. If no configuration
	// is provided, it will assume that a PersistentVolumeClaim called "pravega-longterm"
	// is present and it will use it as Tier 2
	// +optional
	LongTermStorage *LongTermStorageSpec `json:"longtermStorage"`

	// ControllerServiceAccountName configures the service account used on controller instances.
	// If not specified, Kubernetes will automatically assign the default service account in the namespace
	ControllerServiceAccountName string `json:"controllerServiceAccountName,omitempty"`

	// SegmentStoreServiceAccountName configures the service account used on segment store instances.
	// If not specified, Kubernetes will automatically assign the default service account in the namespace
	SegmentStoreServiceAccountName string `json:"segmentStoreServiceAccountName,omitempty"`

	// ControllerResources specifies the request and limit of resources that controller can have.
	// ControllerResources includes CPU and memory resources
	ControllerResources *v1.ResourceRequirements `json:"controllerResources,omitempty"`

	// SegmentStoreResources specifies the request and limit of resources that segmentStore can have.
	// SegmentStoreResources includes CPU and memory resources
	SegmentStoreResources *v1.ResourceRequirements `json:"segmentStoreResources,omitempty"`

	// Provides the name of the configmap created by the user to provide additional key-value pairs
	// that need to be configured into the ss pod as environmental variables
	SegmentStoreEnvVars string `json:"segmentStoreEnvVars,omitempty"`

	// SegmentStoreSecret specifies whether or not any secret needs to be configured into the ss pod
	// either as an environment variable or by mounting it to a volume
	// +optional
	SegmentStoreSecret *SegmentStoreSecret `json:"segmentStoreSecret"`

	// Type specifies the service type to achieve external access.
	// Options are "LoadBalancer" and "NodePort".
	// By default, if external access is enabled, it will use "LoadBalancer"
	ControllerExternalServiceType v1.ServiceType `json:"controllerExtServiceType,omitempty"`

	// Annotations to be added to the external service
	// +optional
	ControllerServiceAnnotations map[string]string `json:"controllerSvcAnnotations"`
	// Labels to be added to the Controller pods
	// +optional
	ControllerPodLabels map[string]string `json:"controllerPodLabels"`

	// Annotations to be added to controller pods
	// +optonal
	ControllerPodAnnotations map[string]string `json:"controllerPodAnnotations"`
	// Type specifies the service type to achieve external access.
	// Options are "LoadBalancer" and "NodePort".
	// By default, if external access is enabled, it will use "LoadBalancer"
	SegmentStoreExternalServiceType v1.ServiceType `json:"segmentStoreExtServiceType,omitempty"`

	// Annotations to be added to the external service
	// +optional
	SegmentStoreServiceAnnotations map[string]string `json:"segmentStoreSvcAnnotations"`

	// Labels to be added to the SegmentStore pods
	// +optional
	SegmentStorePodLabels map[string]string `json:"segmentStorePodLabels"`

	// Annotations to be added to the SegmentStore pods
	// +optional
	SegmentStorePodAnnotations map[string]string `json:"segmentStorePodAnnotations"`

	// Specifying this IP would ensure we use same IP address for all the ss services
	SegmentStoreLoadBalancerIP string `json:"segmentStoreLoadBalancerIP,omitempty"`

	// SegmentStoreExternalTrafficPolicy defines the ExternalTrafficPolicy it can have cluster or local
	SegmentStoreExternalTrafficPolicy string `json:"segmentStoreExternalTrafficPolicy,omitempty"`

	// SegmentStoreSecurityContext holds security configuration that will be applied to a container
	SegmentStoreSecurityContext *corev1.PodSecurityContext `json:"segmentStoreSecurityContext,omitempty"`

	// ControllerSecurityContext holds security configuration that will be applied to a container
	ControllerSecurityContext *corev1.PodSecurityContext `json:"controllerSecurityContext,omitempty"`

	// The scheduling constraints on Controller pods.
	ControllerPodAffinity *corev1.Affinity `json:"controllerPodAffinity,omitempty"`

	// The scheduling constraints on Segementstore pods.
	SegmentStorePodAffinity *corev1.Affinity `json:"segmentStorePodAffinity,omitempty"`

	// This is used to schedule the timeout value for rollback
	RollbackTimeout int32 `json:"rollbacktimeout,omitempty"`

	// This is used as suffix for controller service name
	ControllerSvcNameSuffix string `json:"controllerSvcNameSuffix,omitempty"`

	// This is used as suffix for segmentstore sts name
	SegmentStoreStsNameSuffix string `json:"segmentStoreStsNameSuffix,omitempty"`

	// This is used as suffix for segmentstore headless service name
	SegmentStoreHeadlessSvcNameSuffix string `json:"segmentStoreHeadlessSvcNameSuffix,omitempty"`

	// InitContainers to be added to controller pods
	ControllerInitContainers []v1.Container `json:"controllerInitContainers,omitempty"`

	// InitContainers to be added to segmentstore pods
	SegmentStoreInitContainers []v1.Container `json:"segmentStoreInitContainers,omitempty"`

	// Details of authplugin to be copied into pravega controller
	AuthImplementations *AuthImplementationSpec `json:"authImplementations,omitempty"`

	// InfluxDB Secret specifies the secret name containing credentials and mount path volume
	// that has to be configured in controller and segmentstore pods
	InfluxDBSecret *InfluxDBSecret `json:"influxDBSecret,omitempty"`
}

func (s *PravegaSpec) withDefaults() (changed bool) {
	if !config.TestMode && s.ControllerReplicas < 1 {
		changed = true
		s.ControllerReplicas = 1
	}

	if !config.TestMode && s.RollbackTimeout < 1 {
		changed = true
		s.RollbackTimeout = 10
	}

	if !config.TestMode && s.SegmentStoreReplicas < 1 {
		changed = true
		s.SegmentStoreReplicas = 1
	}

	if !config.TestMode && s.MaxUnavailableControllerReplicas < 1 {
		changed = true
		s.MaxUnavailableControllerReplicas = 1
	}

	if !config.TestMode && s.MaxUnavailableSegmentStoreReplicas < 1 {
		changed = true
		s.MaxUnavailableSegmentStoreReplicas = 1
	}

	if s.Image == nil {
		changed = true
		s.Image = &ImageSpec{}
	}
	if s.Image.withDefaults() {
		changed = true
	}

	if s.Options == nil {
		changed = true
		s.Options = map[string]string{}
	}

	if s.ControllerJvmOptions == nil {
		changed = true
		s.ControllerJvmOptions = []string{}
	}

	if s.SegmentStoreJVMOptions == nil {
		changed = true
		s.SegmentStoreJVMOptions = []string{}
	}

	if s.LongTermStorage == nil {
		changed = true
		s.LongTermStorage = &LongTermStorageSpec{}
	}

	if s.LongTermStorage.withDefaults() {
		changed = true
	}

	if s.ControllerResources == nil {
		changed = true
		s.ControllerResources = &v1.ResourceRequirements{
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

	if s.SegmentStoreResources == nil {
		changed = true
		s.SegmentStoreResources = &v1.ResourceRequirements{
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

	if s.SegmentStoreSecret == nil {
		changed = true
		s.SegmentStoreSecret = &SegmentStoreSecret{}
	}

	if s.SegmentStoreSecret.withDefaults() {
		changed = true
	}

	if s.InfluxDBSecret == nil {
		changed = true
		s.InfluxDBSecret = &InfluxDBSecret{}
	}

	if s.InfluxDBSecret.withDefaults() {
		changed = true
	}
	if s.ControllerServiceAnnotations == nil {
		changed = true
		s.ControllerServiceAnnotations = map[string]string{}
	}
	if s.ControllerPodLabels == nil {
		changed = true
		s.ControllerPodLabels = map[string]string{}
	}

	if s.SegmentStoreServiceAnnotations == nil {
		changed = true
		s.SegmentStoreServiceAnnotations = map[string]string{}
	}

	if s.SegmentStorePodLabels == nil {
		changed = true
		s.SegmentStorePodLabels = map[string]string{}
	}

	if s.ControllerSvcNameSuffix == "" {
		changed = true
		s.ControllerSvcNameSuffix = "pravega-controller"
	}

	if s.SegmentStoreHeadlessSvcNameSuffix == "" {
		changed = true
		s.SegmentStoreHeadlessSvcNameSuffix = "pravega-segmentstore-headless"
	}
	if s.SegmentStoreStsNameSuffix == "" {
		changed = true
		s.SegmentStoreStsNameSuffix = "pravega-segment-store"
	}
	if s.ControllerPodAnnotations == nil {
		changed = true
		s.ControllerPodAnnotations = map[string]string{}
	}
	if s.SegmentStorePodAnnotations == nil {
		changed = true
		s.SegmentStorePodAnnotations = map[string]string{}
	}
	return changed
}

// SegmentStoreSecret defines the configuration of the secret for the Segment Store
type SegmentStoreSecret struct {
	// Secret specifies the name of Secret which needs to be configured
	// +optional
	Secret string `json:"secret"`

	// Path to the volume where the secret will be mounted
	// This value is considered only when the secret is provided
	// If this value is provided, the secret is mounted to a Volume
	// else the secret is exposed as an Environment Variable
	// +optional
	MountPath string `json:"mountPath"`
}

func (s *SegmentStoreSecret) withDefaults() (changed bool) {
	if s.Secret == "" {
		s.MountPath = ""
	}

	return changed
}

// InfluxDBSecret defines the configuration of the secret for controller and
// segmentstore
type InfluxDBSecret struct {
	// Secret specifies the name of Secret which needs to be configured
	Secret string `json:"secret"`

	// Path to the volume where the secret will be mounted
	// This value is considered only when the secret is provided
	// If this value is provided, the secret is mounted to specified Volume
	// else default mount path is used.
	// +optional
	MountPath string `json:"mountPath"`
}

func (s *InfluxDBSecret) withDefaults() (changed bool) {
	if s.Secret != "" && s.MountPath == "" {
		s.MountPath = DefaultInfluxDBSecretMountDir
		changed = true
	}

	return changed
}

func (s *ImageSpec) withDefaults() (changed bool) {
	if s.Repository == "" {
		changed = true
		s.Repository = DefaultPravegaImageRepository
	}

	if s.PullPolicy == "" {
		changed = true
		s.PullPolicy = DefaultPravegaImagePullPolicy
	}

	return changed
}

// LongTermStorageSpec configures the Tier 2 storage type to use with Pravega.
// If not specified, Tier 2 will be configured in filesystem mode and will try
// to use a PersistentVolumeClaim with the name "pravega-longterm"
type LongTermStorageSpec struct {
	// FileSystem is used to configure a pre-created Persistent Volume Claim
	// as Tier 2 backend.
	// It is default Tier 2 mode.
	FileSystem *FileSystemSpec `json:"filesystem,omitempty"`

	// Ecs is used to configure a Dell EMC ECS system as a Tier 2 backend
	Ecs *ECSSpec `json:"ecs,omitempty"`

	// Hdfs is used to configure an HDFS system as a Tier 2 backend
	Hdfs *HDFSSpec `json:"hdfs,omitempty"`
}

// AuthImpemenationSpec helps to inject plugins to contoller pod
type AuthImplementationSpec struct {
	MountPath    string            `json:"mountPath,omitempty"`
	AuthHandlers []AuthHandlerSpec `json:"authHandlers"`
}

// Contains details of auth handler
type AuthHandlerSpec struct {
	Image  string `json:"image"`
	Source string `json:"source"`
}

func (s *LongTermStorageSpec) withDefaults() (changed bool) {
	if s.FileSystem == nil && s.Ecs == nil && s.Hdfs == nil {
		changed = true
		fs := &FileSystemSpec{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: DefaultPravegaLTSClaimName,
			},
		}
		s.FileSystem = fs
	}

	return changed
}

// FileSystemSpec contains the reference to a PVC.
type FileSystemSpec struct {
	// +optional
	PersistentVolumeClaim *v1.PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim"`
}

// ECSSpec contains the connection details to a Dell EMC ECS system
type ECSSpec struct {
	// +optional
	ConfigUri string `json:"configUri"`
	// +optional
	Bucket string `json:"bucket"`
	// +optional
	Prefix string `json:"prefix"`
	// +optional
	Credentials string `json:"credentials"`
}

// HDFSSpec contains the connection details to an HDFS system
type HDFSSpec struct {
	// +optional
	Uri string `json:"uri"`
	// +optional
	Root string `json:"root"`
	// +optional
	ReplicationFactor int32 `json:"replicationFactor"`
}
