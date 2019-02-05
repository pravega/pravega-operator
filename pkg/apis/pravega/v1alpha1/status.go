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
	"time"

	corev1 "k8s.io/api/core/v1"
)

type ClusterConditionType string

const (
	ClusterConditionPodsReady ClusterConditionType = "PodsReady"
)

// ClusterStatus defines the observed state of PravegaCluster
type ClusterStatus struct {
	// Conditions list all the applied conditions
	Conditions []ClusterCondition `json:"conditions,omitempty"`

	// CurrentVersion is the current cluster version
	CurrentVersion string `json:"currentVersion"`

	// TargetVersion is the version the cluster upgrading to.
	// If the cluster is not upgrading, TargetVersion is empty.
	TargetVersion string `json:"targetVersion"`

	// Replicas is the number of number of desired replicas in the cluster
	Replicas int32 `json:"replicas"`

	// ReadyReplicas is the number of number of ready replicas in the cluster
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

func (ps *ClusterStatus) ContainsCondition(condType ClusterConditionType, value corev1.ConditionStatus) bool {
	if _, conditon := ps.getClusterCondition(condType); conditon != nil && conditon.Status == value {
		return true
	}
	return false
}

func (ps *ClusterStatus) SetPodsReadyConditionTrue() {
	c := newClusterCondition(ClusterConditionPodsReady, corev1.ConditionTrue, "", "")
	ps.setClusterCondition(*c)
}

func (ps *ClusterStatus) SetPodsReadyConditionFalse() {
	c := newClusterCondition(ClusterConditionPodsReady, corev1.ConditionFalse, "", "")
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

func (ps *ClusterStatus) getClusterCondition(t ClusterConditionType) (int, *ClusterCondition) {
	for i, c := range ps.Conditions {
		if t == c.Type {
			return i, &c
		}
	}
	return -1, nil
}

func (ps *ClusterStatus) setClusterCondition(newCondition ClusterCondition) {
	now := time.Now().Format(time.RFC3339)
	position, existingCondition := ps.getClusterCondition(newCondition.Type)

	if existingCondition == nil {
		ps.Conditions = append(ps.Conditions, newCondition)
		return
	}

	if existingCondition.Status != newCondition.Status {
		newCondition.LastTransitionTime = now
		newCondition.LastUpdateTime = now
	}

	if existingCondition.Reason != newCondition.Reason || existingCondition.Message != newCondition.Message {
		newCondition.LastUpdateTime = now
	}

	ps.Conditions[position] = newCondition
}
