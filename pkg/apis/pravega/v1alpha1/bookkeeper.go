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
	// DefaultBookkeeperImageRepository is the default Docker repository for
	// the BookKeeper image
	DefaultBookkeeperImageRepository = "pravega/bookkeeper"

	// DefaultBookkeeperLedgerVolumeSize is the default volume size for the
	// Bookkeeper ledger volume
	DefaultBookkeeperLedgerVolumeSize = "10Gi"

	// DefaultBookkeeperJournalVolumeSize is the default volume size for the
	// Bookkeeper journal volume
	DefaultBookkeeperJournalVolumeSize = "10Gi"

	// MinimumBookkeeperReplicas is the minimum number of Bookkeeper replicas
	// accepted
	MinimumBookkeeperReplicas = 3
)

// BookkeeperSpec defines the configuration of BookKeeper
type BookkeeperSpec struct {
	// ImageRepository defines the Docker image repository.
	// By default, "pravega/bookkeeper" will be used.
	ImageRepository string `json:"imageRepository"`

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
}

func (s *BookkeeperSpec) withDefaults() {
	if len(s.ImageRepository) == 0 {
		s.ImageRepository = DefaultBookkeeperImageRepository
	}

	if !config.TestMode && s.Replicas < MinimumBookkeeperReplicas {
		s.Replicas = MinimumBookkeeperReplicas
	}

	if s.Storage == nil {
		s.Storage = &BookkeeperStorageSpec{}
	}
	s.Storage.withDefaults()

	if s.AutoRecovery == nil {
		boolTrue := true
		s.AutoRecovery = &boolTrue
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
