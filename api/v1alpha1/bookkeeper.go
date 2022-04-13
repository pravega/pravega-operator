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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// DefaultBookkeeperImageRepository is the default Docker repository for
	// the BookKeeper image
	DefaultBookkeeperImageRepository = "pravega/bookkeeper"

	// DefaultPravegaImagePullPolicy is the default image pull policy used
	// for the Pravega Docker image
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

// BookkeeperSpec defines the configuration of BookKeeper
type BookkeeperSpec struct {
	// Image defines the BookKeeper Docker image to use.
	// By default, "pravega/bookkeeper" will be used.
	// +optional
	Image *BookkeeperImageSpec `json:"image"`

	// Replicas defines the number of BookKeeper replicas.
	// Minimum is 3. Defaults to 3.
	// +optional
	Replicas int32 `json:"replicas"`

	// Storage configures the storage for BookKeeper
	// +optional
	Storage *BookkeeperStorageSpec `json:"storage"`

	// AutoRecovery indicates whether or not BookKeeper auto recovery is enabled.
	// Defaults to true.
	// +optional
	AutoRecovery *bool `json:"autoRecovery"`

	// ServiceAccountName configures the service account used on BookKeeper instances
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// BookieResources specifies the request and limit of resources that bookie can have.
	// BookieResources includes CPU and memory resources
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Options is the Bookkeeper configuration that is to override the bk_server.conf
	// in bookkeeper. Some examples can be found here
	// https://github.com/apache/bookkeeper/blob/master/docker/README.md
	// +optional
	Options map[string]string `json:"options"`

	// JVM is the JVM options for bookkeeper. It will be passed to the JVM for performance tuning.
	// If this field is not specified, the operator will use a set of default
	// options that is good enough for general deployment.
	// +optional
	BookkeeperJVMOptions *BookkeeperJVMOptions `json:"bookkeeperJVMOptions"`
}

func (s *BookkeeperSpec) withDefaults() (changed bool) {
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

	if s.BookkeeperJVMOptions == nil {
		changed = true
		s.BookkeeperJVMOptions = &BookkeeperJVMOptions{}
	}

	if s.BookkeeperJVMOptions.withDefaults() {
		changed = true
	}

	return changed
}

// BookkeeperImageSpec defines the fields needed for a BookKeeper Docker image
type BookkeeperImageSpec struct {
	ImageSpec `json:"imageSpec,omitempty"`
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

type BookkeeperJVMOptions struct {
	// +optional
	MemoryOpts []string `json:"memoryOpts"`
	// +optional
	GcOpts []string `json:"gcOpts"`
	// +optional
	GcLoggingOpts []string `json:"gcLoggingOpts"`
	// +optional
	ExtraOpts []string `json:"extraOpts"`
}

func (s *BookkeeperJVMOptions) withDefaults() (changed bool) {
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
	// +optional
	LedgerVolumeClaimTemplate *v1.PersistentVolumeClaimSpec `json:"ledgerVolumeClaimTemplate"`

	// JournalVolumeClaimTemplate is the spec to describe PVC for the BookKeeper journal
	// This field is optional. If no PVC spec and there is no default storage class,
	// stateful containers will use emptyDir as volume
	// +optional
	JournalVolumeClaimTemplate *v1.PersistentVolumeClaimSpec `json:"journalVolumeClaimTemplate"`

	// IndexVolumeClaimTemplate is the spec to describe PVC for the BookKeeper index
	// This field is optional. If no PVC spec and there is no default storage class,
	// stateful containers will use emptyDir as volume
	// +optional
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
