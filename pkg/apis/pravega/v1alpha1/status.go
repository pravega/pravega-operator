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
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
)

type ClusterConditionType string

const (
	ClusterConditionPodsReady ClusterConditionType = "PodsReady"
	ClusterConditionUpgrading                      = "Upgrading"
	ClusterConditionRollback                       = "RollbackInProgress"
	ClusterConditionError                          = "Error"
)

// ClusterStatus defines the observed state of PravegaCluster
type ClusterStatus struct {
	// Conditions list all the applied conditions
	Conditions []ClusterCondition `json:"conditions,omitempty"`

	// CurrentVersion is the current cluster version
	CurrentVersion string `json:"currentVersion,omitempty"`

	// TargetVersion is the version the cluster upgrading to.
	// If the cluster is not upgrading, TargetVersion is empty.
	TargetVersion string `json:"targetVersion,omitempty"`

	VersionHistory []string `json:"versionHistory,omitempty"`

	// Replicas is the number of desired replicas in the cluster
	Replicas int32 `json:"replicas"`

	// CurrentReplicas is the number of current replicas in the cluster
	CurrentReplicas int32 `json:"currentReplicas"`

	// ReadyReplicas is the number of ready replicas in the cluster
	ReadyReplicas int32 `json:"readyReplicas"`

	// Members is the Pravega members in the cluster
	Members MembersStatus `json:"members"`
}

// MembersStatus is the status of the members of the cluster with both
// ready and unready node membership lists
type MembersStatus struct {
	Ready   []string `json:"ready"`
	Unready []string `json:"unready"`
}

// ClusterCondition shows the current condition of a Pravega cluster.
// Comply with k8s API conventions
type ClusterCondition struct {
	// Type of Pravega cluster condition.
	Type ClusterConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
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

func (ps *ClusterStatus) Init() {
	// Initialise conditions
	conditionTypes := []ClusterConditionType{
		ClusterConditionPodsReady,
		ClusterConditionUpgrading,
		ClusterConditionError,
	}
	for _, conditionType := range conditionTypes {
		if _, condition := ps.GetClusterCondition(conditionType); condition == nil {
			c := newClusterCondition(conditionType, corev1.ConditionFalse, "", "")
			ps.setClusterCondition(*c)
		}
	}

	// Set current cluster version in version history,
	// so if the first upgrade fails we can rollback to this version
	if ps.VersionHistory == nil && ps.CurrentVersion != "" {
		ps.VersionHistory = []string{ps.CurrentVersion}
	}
}

func (ps *ClusterStatus) SetPodsReadyConditionTrue() {
	c := newClusterCondition(ClusterConditionPodsReady, corev1.ConditionTrue, "", "")
	ps.setClusterCondition(*c)
}

func (ps *ClusterStatus) SetPodsReadyConditionFalse() {
	c := newClusterCondition(ClusterConditionPodsReady, corev1.ConditionFalse, "", "")
	ps.setClusterCondition(*c)
}

func (ps *ClusterStatus) SetUpgradingConditionTrue() {
	c := newClusterCondition(ClusterConditionUpgrading, corev1.ConditionTrue, "", "")
	ps.setClusterCondition(*c)
}

func (ps *ClusterStatus) SetUpgradingConditionFalse() {
	c := newClusterCondition(ClusterConditionUpgrading, corev1.ConditionFalse, "", "")
	ps.setClusterCondition(*c)
}

func (ps *ClusterStatus) SetUpgradedReplicasForComponent(componentName string, updatedReplicas int32, totalReplicas int32) {
	_, upgradeCondition := ps.GetClusterCondition(ClusterConditionUpgrading)
	if upgradeCondition != nil && upgradeCondition.Status == corev1.ConditionTrue {
		reason := fmt.Sprintf("Upgrading component: %s", componentName)
		message := fmt.Sprintf("Upgraded Replicas: %v, Total Replicas: %v", updatedReplicas, totalReplicas)
		c := newClusterCondition(ClusterConditionUpgrading, corev1.ConditionTrue, reason, message)
		ps.setClusterCondition(*c)
	}
}

func (ps *ClusterStatus) SetErrorConditionTrue(reason, message string) {
	c := newClusterCondition(ClusterConditionError, corev1.ConditionTrue, reason, message)
	ps.setClusterCondition(*c)
}

func (ps *ClusterStatus) SetErrorConditionFalse() {
	c := newClusterCondition(ClusterConditionError, corev1.ConditionFalse, "", "")
	ps.setClusterCondition(*c)
}

func (ps *ClusterStatus) SetRollbackConditionTrue() {
	c := newClusterCondition(ClusterConditionRollback, corev1.ConditionTrue, "", "")
	ps.setClusterCondition(*c)
}
func (ps *ClusterStatus) SetRollbackConditionFalse() {
	c := newClusterCondition(ClusterConditionRollback, corev1.ConditionFalse, "", "")
	ps.setClusterCondition(*c)
}

func newClusterCondition(condType ClusterConditionType, status corev1.ConditionStatus, reason, message string) *ClusterCondition {
	return &ClusterCondition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastUpdateTime:     "",
		LastTransitionTime: "",
	}
}

func (ps *ClusterStatus) GetClusterCondition(t ClusterConditionType) (int, *ClusterCondition) {
	for i, c := range ps.Conditions {
		if t == c.Type {
			return i, &c
		}
	}
	return -1, nil
}

func (ps *ClusterStatus) setClusterCondition(newCondition ClusterCondition) {
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

func (ps *ClusterStatus) AddToVersionHistory(version string) {
	lastIndex := len(ps.VersionHistory) - 1
	if version != "" && ps.VersionHistory[lastIndex] != version {
		ps.VersionHistory = append(ps.VersionHistory, version)
		log.Printf("Updating version history adding version %v", version)
	}
}

func (ps *ClusterStatus) GetLastVersion() (previousVersion string, err error) {
	if ps.VersionHistory == nil {
		return "", fmt.Errorf("ERROR: No previous cluster version found")
	}
	len := len(ps.VersionHistory)
	return ps.VersionHistory[len-1], nil
}

func (ps *ClusterStatus) HasUpgradeFailed() bool {
	_, errorCondition := ps.GetClusterCondition(ClusterConditionError)
	if errorCondition == nil {
		return false
	}
	if errorCondition.Status == corev1.ConditionTrue && errorCondition.Reason == "UpgradeFailed" {
		return true
	}
	return false
}
