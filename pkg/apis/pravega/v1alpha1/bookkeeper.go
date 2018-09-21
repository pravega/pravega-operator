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
)

type BookkeeperSpec struct {
	Image        ImageSpec             `json:"image"`
	Replicas     int32                 `json:"replicas"`
	Storage      BookkeeperStorageSpec `json:"storage"`
	AutoRecovery bool                  `json:"autoRecovery"`
}

type BookkeeperStorageSpec struct {
	LedgerVolumeClaimTemplate  v1.PersistentVolumeClaimSpec `json:"ledgerVolumeClaimTemplate"`
	JournalVolumeClaimTemplate v1.PersistentVolumeClaimSpec `json:"journalVolumeClaimTemplate"`
}
