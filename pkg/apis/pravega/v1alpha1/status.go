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
	corev1 "k8s.io/api/core/v1"
	"time"
)

type PravegaClusterConditionType string

const (
	// cluster conditions
	PravegaClusterConditionReady   PravegaClusterConditionType = "Ready"
	PravegaClusterConditionScaling                             = "Scaling"
	PravegaClusterConditionError                               = "Error"

	// ready condition reasons
	PodsNotReady  = "PodsNotReady"
	PodsUnhealthy = "PodsUnhealthy"

	// error condition reasons
	OperatorError = "OperatorError"
)

// PravegaClusterStatus defines the observed state of PravegaCluster
type PravegaClusterStatus struct {
	// Conditions list all the applied conditions
	Conditions []PravegaClusterCondition `json:"conditions,omitempty"`

	// CurrentVersion is the current cluster version
	CurrentVersion string `json:"currentVersion"`

	// TargetVersion is the version the cluster upgrading to.
	// If the cluster is not upgrading, TargetVersion is empty.
	TargetVersion string `json:"targetVersion"`
}

// PravegaClusterCondition shows the current condition of a Pravega cluster.
// Comply with k8s API conventions
type PravegaClusterCondition struct {
	// Type of Pravega cluster condition.
	Type PravegaClusterConditionType `json:"type"`

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

func (ps *PravegaClusterStatus) withDefaults() {
	ps.SetReadyConditionFalse(PodsNotReady, "")
	ps.SetScalingConditionFalse()
	ps.SetErrorConditionFalse()
}

func (ps *PravegaClusterStatus) ContainsCondition(condType PravegaClusterConditionType) bool {
	if _, conditon := ps.getClusterCondition(condType); conditon != nil {
		return true
	}
	return false
}

func (ps *PravegaClusterStatus) SetReadyConditionTrue() {
	c := newClusterCondition(PravegaClusterConditionReady, corev1.ConditionTrue, "", "")
	ps.setClusterCondition(*c)
}

func (ps *PravegaClusterStatus) SetScalingConditionTrue() {
	c := newClusterCondition(PravegaClusterConditionScaling, corev1.ConditionTrue, "", "")
	ps.setClusterCondition(*c)
}

func (ps *PravegaClusterStatus) SetErrorConditionTrue(reason, message string) {
	c := newClusterCondition(PravegaClusterConditionScaling, corev1.ConditionTrue, reason, message)
	ps.setClusterCondition(*c)
}

func (ps *PravegaClusterStatus) SetReadyConditionFalse(reason, message string) {
	c := newClusterCondition(PravegaClusterConditionReady, corev1.ConditionFalse, reason, message)
	ps.setClusterCondition(*c)
}

func (ps *PravegaClusterStatus) SetScalingConditionFalse() {
	c := newClusterCondition(PravegaClusterConditionScaling, corev1.ConditionFalse, "", "")
	ps.setClusterCondition(*c)
}

func (ps *PravegaClusterStatus) SetErrorConditionFalse() {
	c := newClusterCondition(PravegaClusterConditionScaling, corev1.ConditionFalse, "", "")
	ps.setClusterCondition(*c)
}

func (ps *PravegaClusterStatus) ClearReadyCondition() {
	ps.deleteClusterCondition(PravegaClusterConditionReady)
}

func (ps *PravegaClusterStatus) ClearScalingCondition() {
	ps.deleteClusterCondition(PravegaClusterConditionScaling)
}

func (ps *PravegaClusterStatus) ClearErrorCondition() {
	ps.deleteClusterCondition(PravegaClusterConditionError)
}

func newClusterCondition(condType PravegaClusterConditionType, status corev1.ConditionStatus, reason, message string) *PravegaClusterCondition {
	now := time.Now().Format(time.RFC3339)
	return &PravegaClusterCondition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastUpdateTime:     now,
		LastTransitionTime: "",
	}
}

func (ps *PravegaClusterStatus) getClusterCondition(t PravegaClusterConditionType) (int, *PravegaClusterCondition) {
	for i, c := range ps.Conditions {
		if t == c.Type {
			return i, &c
		}
	}
	return -1, nil
}

func (ps *PravegaClusterStatus) setClusterCondition(c PravegaClusterCondition) {
	position, condition := ps.getClusterCondition(c.Type)
	if condition != nil && condition.Status == c.Status && condition.Reason == c.Reason && condition.Message == c.Message {
		condition.LastUpdateTime = time.Now().Format(time.RFC3339)
		return
	}

	if condition != nil {
		c.LastTransitionTime = time.Now().Format(time.RFC3339)
		ps.Conditions[position] = c
	} else {
		ps.Conditions = append(ps.Conditions, c)
	}
}

func (ps *PravegaClusterStatus) deleteClusterCondition(t PravegaClusterConditionType) {
	pos, _ := ps.getClusterCondition(t)
	if pos == -1 {
		return
	}
	ps.Conditions = append(ps.Conditions[:pos], ps.Conditions[pos+1:]...)
}
