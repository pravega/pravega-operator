package v1alpha1

import (
	"k8s.io/api/core/v1"
)

type BookkeeperSpec struct {
	Image        ImageSpec             `json:"image"`
	Replicas     int32                 `json:"relicas"`
	Storage      BookkeeperStorageSpec `json:"storage"`
	Options      map[string]string     `json:"options"`
	AutoRecovery bool                  `json:"autoRecovery"`
}

type BookkeeperStorageSpec struct {
	LedgerVolumeClaimTemplate  v1.PersistentVolumeClaimSpec `json:"ledgerVolumeClaimTemplate"`
	JournalVolumeClaimTemplate v1.PersistentVolumeClaimSpec `json:"journalVolumeClaimTemplate"`
}
