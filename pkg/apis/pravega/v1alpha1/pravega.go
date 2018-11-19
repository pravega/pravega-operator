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
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// DefaultBookkeeperImageRepository is the default Docker repository for
	// the BookKeeper image
	DefaultPravegaImageRepository = "pravega/pravega"

	// DefaultBookkeeperImageTag is the default tag used for for the BookKeeper
	// Docker image
	DefaultPravegaImageTag = "latest"

	// DefaultBookkeeperImagePullPolicy is the default image pull policy used
	// for the BookKeeper Docker image
	DefaultPravegaImagePullPolicy = v1.PullAlways

	// DefaultPravegaCacheVolumeSize is the default volume size for the
	// Pravega SegmentStore cache volume
	DefaultPravegaCacheVolumeSize = "20Gi"

	// DefaultPravegaTier2ClaimName is the default volume claim name used as Tier 2
	DefaultPravegaTier2ClaimName = "pravega-tier2"
)

// PravegaSpec defines the configuration of Pravega
type PravegaSpec struct {
	// ControllerReplicas defines the number of Controller replicas.
	// Defaults to 1.
	ControllerReplicas int32 `json:"controllerReplicas"`

	// SegmentStoreReplicas defines the number of Segment Store replicas.
	// Defaults to 1.
	SegmentStoreReplicas int32 `json:"segmentStoreReplicas"`

	// DebugLogging indicates whether or not debug level logging is enabled.
	// Defaults to false.
	DebugLogging bool `json:"debugLogging"`

	// Image defines the Pravega Docker image to use.
	// By default, "pravega/pravega:latest" will be used.
	Image PravegaImageSpec `json:"image"`

	// Options is the Pravega configuration that is passed to the Pravega processes
	// as JAVA_OPTS. See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/config/config.properties
	Options map[string]string `json:"options"`

	// CacheVolumeClaimTemplate is the spec to describe PVC for the Pravega cache.
	// This field is optional. If no PVC spec, stateful containers will use
	// emptyDir as volume
	CacheVolumeClaimTemplate *v1.PersistentVolumeClaimSpec `json:"cacheVolumeClaimTemplate"`

	// Tier2 is the configuration of Pravega's tier 2 storage. If no configuration
	// is provided, it will assume that a PersistentVolumeClaim called "pravega-tier2"
	// is present and it will use it as Tier 2
	Tier2 Tier2Spec `json:"tier2"`
}

func (s *PravegaSpec) withDefaults() {
	if s == nil {
		s = &PravegaSpec{}
	}

	if s.ControllerReplicas < 1 {
		s.ControllerReplicas = 1
	}

	if s.SegmentStoreReplicas < 1 {
		s.SegmentStoreReplicas = 1
	}

	s.Image.withDefaults()

	if s.Options == nil {
		s.Options = map[string]string{}
	}

	if s.CacheVolumeClaimTemplate == nil {
		s.CacheVolumeClaimTemplate = &v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(DefaultPravegaCacheVolumeSize),
				},
			},
		}
	}

	s.Tier2.withDefaults()
}

// PravegaImageSpec defines the fields needed for a Pravega Docker image
type PravegaImageSpec struct {
	ImageSpec
}

// String formats a container image struct as a Docker compatible repository string
func (s *PravegaImageSpec) String() string {
	return fmt.Sprintf("%s:%s", s.Repository, s.Tag)
}

func (s *PravegaImageSpec) withDefaults() {
	if s == nil {
		s = &PravegaImageSpec{}
	}

	if s.Repository == "" {
		s.Repository = DefaultPravegaImageRepository
	}

	if s.Tag == "" {
		s.Tag = DefaultPravegaImageTag
	}

	if s.PullPolicy == "" {
		s.PullPolicy = DefaultPravegaImagePullPolicy
	}
}

// Tier2Spec configures the Tier 2 storage type to use with Pravega
type Tier2Spec struct {
	FileSystem *FileSystemSpec `json:"filesystem,omitempty"`
	Ecs        *ECSSpec        `json:"ecs,omitempty"`
	Hdfs       *HDFSSpec       `json:"hdfs,omitempty"`
}

func (s *Tier2Spec) withDefaults() {
	if s == nil {
		s = &Tier2Spec{}
		fs := &FileSystemSpec{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: DefaultPravegaTier2ClaimName,
			},
		}
		s.FileSystem = fs
	}
}

type FileSystemSpec struct {
	PersistentVolumeClaim *v1.PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim"`
}

type ECSSpec struct {
	Uri         string `json:"uri"`
	Bucket      string `json:"bucket"`
	Root        string `json:"root"`
	Namespace   string `json:"namespace"`
	Credentials string `json:"credentials"`
}

type HDFSSpec struct {
	Uri               string `json:"uri"`
	Root              string `json:"root"`
	ReplicationFactor int32  `json:"replicationFactor"`
}
