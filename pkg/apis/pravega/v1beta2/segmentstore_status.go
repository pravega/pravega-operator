/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1beta2

import (
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
)

type SegmentStoreConditionType string

const (
	SegmentStoreConditionPodsReady SegmentStoreConditionType = "PodsReady"
	SegmentStoreConditionUpgrading                           = "Upgrading"
	SegmentStoreConditionRollback                            = "RollbackInProgress"
	SegmentStoreConditionError                               = "Error"

	// Reasons for cluster upgrading condition

	UpdatingSegmentstoreReason = "Updating Segmentstore"
	UpdatingBookkeeperReason   = "Updating Bookkeeper"
	UpgradeErrorReason         = "Upgrade Error"
	RollbackErrorReason        = "Rollback Error"
)

// PravegaSegmentStoreStatus defines the observed state of PravegaCluster
type PravegaSegmentStoreStatus struct {
	// Conditions list all the applied conditions
	Conditions []SegmentStoreCondition `json:"conditions,omitempty"`

	// CurrentVersion is the current cluster version
	CurrentVersion string `json:"currentVersion,omitempty"`

	// TargetVersion is the version the cluster upgrading to.
	// If the cluster is not upgrading, TargetVersion is empty.
	TargetVersion string `json:"targetVersion,omitempty"`

	VersionHistory []string `json:"versionHistory,omitempty"`

	// Replicas is the number of desired replicas in the cluster
	// +optional
	Replicas int32 `json:"replicas"`

	// CurrentReplicas is the number of current replicas in the cluster
	// +optional
	CurrentReplicas int32 `json:"currentReplicas"`

	// ReadyReplicas is the number of ready replicas in the cluster
	// +optional
	ReadyReplicas int32 `json:"readyReplicas"`

	// Members is the Pravega members in the cluster
	// +optional
	Members SegmentStoreMembersStatus `json:"members"`
}

// MembersStatus is the status of the members of the cluster with both
// ready and unready node membership lists
type SegmentStoreMembersStatus struct {
	// +optional
	// +nullable
	Ready []string `json:"ready"`
	// +optional
	// +nullable
	Unready []string `json:"unready"`
}

// SegmentStoreCondition shows the current condition of a Pravega cluster.
// Comply with k8s API conventions
type SegmentStoreCondition struct {
	// Type of Pravega cluster condition.
	// +optional
	Type SegmentStoreConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	// +optional
	Status corev1.ConditionStatus `json:"status"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`

	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
}

func (ps *PravegaSegmentStoreStatus) Init() {
	// Initialise conditions
	conditionTypes := []SegmentStoreConditionType{
		SegmentStoreConditionPodsReady,
		SegmentStoreConditionUpgrading,
		SegmentStoreConditionError,
	}
	for _, conditionType := range conditionTypes {
		if _, condition := ps.GetSegmentStoreCondition(conditionType); condition == nil {
			c := newSegmentStoreCondition(conditionType, corev1.ConditionFalse, "", "")
			ps.setSegmentStoreCondition(*c)
		}
	}

	// Set current cluster version in version history,
	// so if the first upgrade fails we can rollback to this version
	if ps.VersionHistory == nil && ps.CurrentVersion != "" {
		ps.VersionHistory = []string{ps.CurrentVersion}
	}
}

func (ps *PravegaSegmentStoreStatus) SetPodsReadyConditionTrue() {
	c := newSegmentStoreCondition(SegmentStoreConditionPodsReady, corev1.ConditionTrue, "", "")
	ps.setSegmentStoreCondition(*c)
}

func (ps *PravegaSegmentStoreStatus) SetPodsReadyConditionFalse() {
	c := newSegmentStoreCondition(SegmentStoreConditionPodsReady, corev1.ConditionFalse, "", "")
	ps.setSegmentStoreCondition(*c)
}

func (ps *PravegaSegmentStoreStatus) SetUpgradingConditionTrue(reason, message string) {
	c := newSegmentStoreCondition(SegmentStoreConditionUpgrading, corev1.ConditionTrue, reason, message)
	ps.setSegmentStoreCondition(*c)
}

func (ps *PravegaSegmentStoreStatus) SetUpgradingConditionFalse() {
	c := newSegmentStoreCondition(SegmentStoreConditionUpgrading, corev1.ConditionFalse, "", "")
	ps.setSegmentStoreCondition(*c)
}

func (ps *PravegaSegmentStoreStatus) SetErrorConditionTrue(reason, message string) {
	c := newSegmentStoreCondition(SegmentStoreConditionError, corev1.ConditionTrue, reason, message)
	ps.setSegmentStoreCondition(*c)
}

func (ps *PravegaSegmentStoreStatus) SetErrorConditionFalse() {
	c := newSegmentStoreCondition(SegmentStoreConditionError, corev1.ConditionFalse, "", "")
	ps.setSegmentStoreCondition(*c)
}

func (ps *PravegaSegmentStoreStatus) SetRollbackConditionTrue(reason, message string) {
	c := newSegmentStoreCondition(SegmentStoreConditionRollback, corev1.ConditionTrue, reason, message)
	ps.setSegmentStoreCondition(*c)
}
func (ps *PravegaSegmentStoreStatus) SetRollbackConditionFalse() {
	c := newSegmentStoreCondition(SegmentStoreConditionRollback, corev1.ConditionFalse, "", "")
	ps.setSegmentStoreCondition(*c)
}

func newSegmentStoreCondition(condType SegmentStoreConditionType, status corev1.ConditionStatus, reason, message string) *SegmentStoreCondition {
	return &SegmentStoreCondition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastUpdateTime:     "",
		LastTransitionTime: "",
	}
}

func (ps *PravegaSegmentStoreStatus) GetSegmentStoreCondition(t SegmentStoreConditionType) (int, *SegmentStoreCondition) {
	for i, c := range ps.Conditions {
		if t == c.Type {
			return i, &c
		}
	}
	return -1, nil
}

func (ps *PravegaSegmentStoreStatus) setSegmentStoreCondition(newCondition SegmentStoreCondition) {
	now := time.Now().Format(time.RFC3339)
	position, existingCondition := ps.GetSegmentStoreCondition(newCondition.Type)

	if existingCondition == nil {
		ps.Conditions = append(ps.Conditions, newCondition)
		return
	}

	if existingCondition.Status != newCondition.Status {
		existingCondition.Status = newCondition.Status
		existingCondition.LastTransitionTime = now
		existingCondition.LastUpdateTime = now
	}

	if existingCondition.Reason != newCondition.Reason || existingCondition.Message != newCondition.Message {
		existingCondition.Reason = newCondition.Reason
		existingCondition.Message = newCondition.Message
		existingCondition.LastUpdateTime = now
	}

	ps.Conditions[position] = *existingCondition
}

func (ps *PravegaSegmentStoreStatus) AddToVersionHistory(version string) {
	lastIndex := len(ps.VersionHistory) - 1
	if version != "" && ps.VersionHistory[lastIndex] != version {
		ps.VersionHistory = append(ps.VersionHistory, version)
		log.Printf("Updating version history adding version %v", version)
	}
}

func (ps *PravegaSegmentStoreStatus) GetLastVersion() (previousVersion string) {
	len := len(ps.VersionHistory)
	return ps.VersionHistory[len-1]
}

func (ps *PravegaSegmentStoreStatus) IsClusterInErrorState() bool {
	_, errorCondition := ps.GetSegmentStoreCondition(SegmentStoreConditionError)
	if errorCondition != nil && errorCondition.Status == corev1.ConditionTrue {
		return true
	}
	return false
}

func (ps *PravegaSegmentStoreStatus) IsClusterInUpgradeFailedState() bool {
	_, errorCondition := ps.GetSegmentStoreCondition(SegmentStoreConditionError)
	if errorCondition == nil {
		return false
	}
	if errorCondition.Status == corev1.ConditionTrue && errorCondition.Reason == "UpgradeFailed" {
		return true
	}
	return false
}

func (ps *PravegaSegmentStoreStatus) IsClusterInUpgradeFailedOrRollbackState() bool {
	if ps.IsClusterInUpgradeFailedState() || ps.IsClusterInRollbackState() {
		return true
	}
	return false
}

func (ps *PravegaSegmentStoreStatus) IsClusterInRollbackState() bool {
	_, rollbackCondition := ps.GetSegmentStoreCondition(SegmentStoreConditionRollback)
	if rollbackCondition == nil {
		return false
	}
	if rollbackCondition.Status == corev1.ConditionTrue {
		return true
	}
	return false
}

func (ps *PravegaSegmentStoreStatus) IsClusterInUpgradingState() bool {
	_, upgradeCondition := ps.GetSegmentStoreCondition(SegmentStoreConditionUpgrading)
	if upgradeCondition == nil {
		return false
	}
	if upgradeCondition.Status == corev1.ConditionTrue {
		return true
	}
	return false
}

func (ps *PravegaSegmentStoreStatus) IsClusterInRollbackFailedState() bool {
	_, errorCondition := ps.GetSegmentStoreCondition(SegmentStoreConditionError)
	if errorCondition == nil {
		return false
	}
	if errorCondition.Status == corev1.ConditionTrue && errorCondition.Reason == "RollbackFailed" {
		return true
	}
	return false
}

func (ps *PravegaSegmentStoreStatus) IsClusterInReadyState() bool {
	_, readyCondition := ps.GetSegmentStoreCondition(SegmentStoreConditionPodsReady)
	if readyCondition != nil && readyCondition.Status == corev1.ConditionTrue {
		return true
	}
	return false
}

func (ps *PravegaSegmentStoreStatus) UpdateProgress(reason, updatedReplicas string) {
	if ps.IsClusterInUpgradingState() {
		// Set the upgrade condition reason to be UpgradingBookkeeperReason, message to be 0
		ps.SetUpgradingConditionTrue(reason, updatedReplicas)
	} else {
		ps.SetRollbackConditionTrue(reason, updatedReplicas)
	}
}

func (ps *PravegaSegmentStoreStatus) GetLastCondition() (lastCondition *SegmentStoreCondition) {
	if ps.IsClusterInUpgradingState() {
		_, lastCondition := ps.GetSegmentStoreCondition(SegmentStoreConditionUpgrading)
		return lastCondition
	} else if ps.IsClusterInRollbackState() {
		_, lastCondition := ps.GetSegmentStoreCondition(SegmentStoreConditionRollback)
		return lastCondition
	}
	// nothing to do if we are neither upgrading nor rolling back,
	return nil
}
