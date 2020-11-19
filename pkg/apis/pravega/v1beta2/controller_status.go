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

type ControllerClusterConditionType string

const (
	ControllerClusterConditionPodsReady ControllerClusterConditionType = "PodsReady"
	ControllerClusterConditionUpgrading                                = "Upgrading"
	ControllerClusterConditionRollback                                 = "RollbackInProgress"
	ControllerClusterConditionError                                    = "Error"

	// Reasons for cluster upgrading condition
	UpdatingControllerReason      = "Updating Controller"
	ControllerUpgradeErrorReason  = "Upgrade Error"
	ControllerRollbackErrorReason = "Rollback Error"
)

// PravegaControllerStatus defines the observed state of PravegaCluster
type PravegaControllerStatus struct {
	// Conditions list all the applied conditions
	Conditions []ControllerClusterCondition `json:"conditions,omitempty"`

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
	Members ControllerMembersStatus `json:"members"`
}

// MembersStatus is the status of the members of the cluster with both
// ready and unready node membership lists
type ControllerMembersStatus struct {
	// +optional
	// +nullable
	Ready []string `json:"ready"`
	// +optional
	// +nullable
	Unready []string `json:"unready"`
}

// ClusterCondition shows the current condition of a Pravega cluster.
// Comply with k8s API conventions
type ControllerClusterCondition struct {
	// Type of Pravega cluster condition.
	// +optional
	Type ControllerClusterConditionType `json:"type"`

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

func (ps *PravegaControllerStatus) Init() {
	// Initialise conditions
	conditionTypes := []ControllerClusterConditionType{
		ControllerClusterConditionPodsReady,
		ControllerClusterConditionUpgrading,
		ControllerClusterConditionError,
	}
	for _, conditionType := range conditionTypes {
		if _, condition := ps.GetClusterCondition(conditionType); condition == nil {
			c := newControllerClusterCondition(conditionType, corev1.ConditionFalse, "", "")
			ps.setClusterCondition(*c)
		}
	}

	// Set current cluster version in version history,
	// so if the first upgrade fails we can rollback to this version
	if ps.VersionHistory == nil && ps.CurrentVersion != "" {
		ps.VersionHistory = []string{ps.CurrentVersion}
	}
}

func (ps *PravegaControllerStatus) SetPodsReadyConditionTrue() {
	c := newControllerClusterCondition(ControllerClusterConditionPodsReady, corev1.ConditionTrue, "", "")
	ps.setClusterCondition(*c)
}

func (ps *PravegaControllerStatus) SetPodsReadyConditionFalse() {
	c := newControllerClusterCondition(ControllerClusterConditionPodsReady, corev1.ConditionFalse, "", "")
	ps.setClusterCondition(*c)
}

func (ps *PravegaControllerStatus) SetUpgradingConditionTrue(reason, message string) {
	c := newControllerClusterCondition(ControllerClusterConditionUpgrading, corev1.ConditionTrue, reason, message)
	ps.setClusterCondition(*c)
}

func (ps *PravegaControllerStatus) SetUpgradingConditionFalse() {
	c := newControllerClusterCondition(ControllerClusterConditionUpgrading, corev1.ConditionFalse, "", "")
	ps.setClusterCondition(*c)
}

func (ps *PravegaControllerStatus) SetErrorConditionTrue(reason, message string) {
	c := newControllerClusterCondition(ControllerClusterConditionError, corev1.ConditionTrue, reason, message)
	ps.setClusterCondition(*c)
}

func (ps *PravegaControllerStatus) SetErrorConditionFalse() {
	c := newControllerClusterCondition(ControllerClusterConditionError, corev1.ConditionFalse, "", "")
	ps.setClusterCondition(*c)
}

func (ps *PravegaControllerStatus) SetRollbackConditionTrue(reason, message string) {
	c := newControllerClusterCondition(ControllerClusterConditionRollback, corev1.ConditionTrue, reason, message)
	ps.setClusterCondition(*c)
}
func (ps *PravegaControllerStatus) SetRollbackConditionFalse() {
	c := newControllerClusterCondition(ControllerClusterConditionRollback, corev1.ConditionFalse, "", "")
	ps.setClusterCondition(*c)
}

func newControllerClusterCondition(condType ControllerClusterConditionType, status corev1.ConditionStatus, reason, message string) *ControllerClusterCondition {
	return &ControllerClusterCondition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastUpdateTime:     "",
		LastTransitionTime: "",
	}
}
func (ps *PravegaControllerStatus) GetClusterCondition(t ControllerClusterConditionType) (int, *ControllerClusterCondition) {
	for i, c := range ps.Conditions {
		if t == c.Type {
			return i, &c
		}
	}
	return -1, nil
}

func (ps *PravegaControllerStatus) setClusterCondition(newCondition ControllerClusterCondition) {
	now := time.Now().Format(time.RFC3339)
	position, existingCondition := ps.GetClusterCondition(newCondition.Type)

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

func (ps *PravegaControllerStatus) AddToVersionHistory(version string) {
	lastIndex := len(ps.VersionHistory) - 1
	if version != "" && ps.VersionHistory[lastIndex] != version {
		ps.VersionHistory = append(ps.VersionHistory, version)
		log.Printf("Updating version history adding version %v", version)
	}
}

func (ps *PravegaControllerStatus) GetLastVersion() (previousVersion string) {
	len := len(ps.VersionHistory)
	return ps.VersionHistory[len-1]
}

func (ps *PravegaControllerStatus) IsClusterInErrorState() bool {
	_, errorCondition := ps.GetClusterCondition(ClusterConditionError)
	if errorCondition != nil && errorCondition.Status == corev1.ConditionTrue {
		return true
	}
	return false
}

func (ps *PravegaControllerStatus) IsClusterInUpgradeFailedState() bool {
	_, errorCondition := ps.GetClusterCondition(ClusterConditionError)
	if errorCondition == nil {
		return false
	}
	if errorCondition.Status == corev1.ConditionTrue && errorCondition.Reason == "UpgradeFailed" {
		return true
	}
	return false
}

func (ps *PravegaControllerStatus) IsClusterInUpgradeFailedOrRollbackState() bool {
	if ps.IsClusterInUpgradeFailedState() || ps.IsClusterInRollbackState() {
		return true
	}
	return false
}

func (ps *PravegaControllerStatus) IsClusterInRollbackState() bool {
	_, rollbackCondition := ps.GetClusterCondition(ClusterConditionRollback)
	if rollbackCondition == nil {
		return false
	}
	if rollbackCondition.Status == corev1.ConditionTrue {
		return true
	}
	return false
}

func (ps *PravegaControllerStatus) IsClusterInUpgradingState() bool {
	_, upgradeCondition := ps.GetClusterCondition(ClusterConditionUpgrading)
	if upgradeCondition == nil {
		return false
	}
	if upgradeCondition.Status == corev1.ConditionTrue {
		return true
	}
	return false
}

func (ps *PravegaControllerStatus) IsClusterInRollbackFailedState() bool {
	_, errorCondition := ps.GetClusterCondition(ClusterConditionError)
	if errorCondition == nil {
		return false
	}
	if errorCondition.Status == corev1.ConditionTrue && errorCondition.Reason == "RollbackFailed" {
		return true
	}
	return false
}

func (ps *PravegaControllerStatus) IsClusterInReadyState() bool {
	_, readyCondition := ps.GetClusterCondition(ControllerClusterConditionPodsReady)
	if readyCondition != nil && readyCondition.Status == corev1.ConditionTrue {
		return true
	}
	return false
}

func (ps *PravegaControllerStatus) UpdateProgress(reason, updatedReplicas string) {
	if ps.IsClusterInUpgradingState() {
		// Set the upgrade condition reason to be UpgradingBookkeeperReason, message to be 0
		ps.SetUpgradingConditionTrue(reason, updatedReplicas)
	} else {
		ps.SetRollbackConditionTrue(reason, updatedReplicas)
	}
}

func (ps *PravegaControllerStatus) GetLastCondition() (lastCondition *ControllerClusterCondition) {
	if ps.IsClusterInUpgradingState() {
		_, lastCondition := ps.GetClusterCondition(ControllerClusterConditionUpgrading)
		return lastCondition
	} else if ps.IsClusterInRollbackState() {
		_, lastCondition := ps.GetClusterCondition(ControllerClusterConditionRollback)
		return lastCondition
	}
	// nothing to do if we are neither upgrading nor rolling back,
	return nil
}
