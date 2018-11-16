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
	DefaultBookkeeperImageRepository = "pravega/bookkeeper"

	// DefaultBookkeeperImageTag is the default tag used for for the BookKeeper
	// Docker image
	DefaultBookkeeperImageTag = "latest"

	// DefaultBookkeeperImagePullPolicy is the default image pull policy used
	// for the BookKeeper Docker image
	DefaultBookkeeperImagePullPolicy = v1.PullAlways

	// DefaultBookkeeperLedgerVolumeSize is the default volume size for the
	// Bookkeeper ledger volume
	DefaultBookkeeperLedgerVolumeSize = "10Gi"

	// DefaultBookkeeperJournalVolumeSize is the default volume size for the
	// Bookkeeper journal volume
	DefaultBookkeeperJournalVolumeSize = "10Gi"
)

// BookkeeperSpec defines the configuration of BookKeeper
type BookkeeperSpec struct {
	// Image defines the BookKeeper Docker image to use.
	// By default, "pravega/bookkeeper:latest" will be used.
	Image BookkeeperImageSpec `json:"image"`

	// Replicas defines the number of BookKeeper replicas.
	// Minimum is 3. Defaults to 3.
	Replicas int32 `json:"replicas"`

	// Storage configures the storage for BookKeeper
	Storage BookkeeperStorageSpec `json:"storage"`

	// AutoRecovery indicates whether or not BookKeeper auto recovery is enabled.
	// Defaults to false.
	AutoRecovery bool `json:"autoRecovery"`
}

func (s *BookkeeperSpec) withDefaults() {
	if s == nil {
		s = &BookkeeperSpec{}
	}

	s.Image.withDefaults()

	if s.Replicas < 3 {
		s.Replicas = 3
	}

	s.Storage.withDefaults()
}

// BookkeeperImageSpec defines the fields needed for a BookKeeper Docker image
type BookkeeperImageSpec struct {
	ImageSpec
}

// String formats a container image struct as a Docker compatible repository string
func (s *BookkeeperImageSpec) String() string {
	return fmt.Sprintf("%s:%s", s.Repository, s.Tag)
}

func (s *BookkeeperImageSpec) withDefaults() {
	if s == nil {
		s = &BookkeeperImageSpec{ImageSpec{}}
	}

	if s.Repository == "" {
		s.Repository = DefaultBookkeeperImageRepository
	}

	if s.Tag == "" {
		s.Tag = DefaultBookkeeperImageTag
	}

	if s.PullPolicy == "" {
		s.PullPolicy = DefaultBookkeeperImagePullPolicy
	}
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
}

func (s *BookkeeperStorageSpec) withDefaults() {
	if s == nil {
		s = &BookkeeperStorageSpec{}
	}

	if s.LedgerVolumeClaimTemplate == nil {
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
		s.JournalVolumeClaimTemplate = &v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(DefaultBookkeeperJournalVolumeSize),
				},
			},
		}
	}
}
