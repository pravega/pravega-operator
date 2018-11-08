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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PravegaClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []PravegaCluster `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PravegaCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PravegaClusterSpec   `json:"spec"`
	Status            PravegaClusterStatus `json:"status,omitempty"`
}

type PravegaClusterSpec struct {
	ZookeeperUri   string         `json:"zookeeperUri"`
	ExternalAccess bool           `json:"externalAccess"`
	Bookkeeper     BookkeeperSpec `json:"bookkeeper"`
	Pravega        PravegaSpec    `json:"pravega"`
}

type ImageSpec struct {
	Repository string        `json:"repository"`
	Tag        string        `json:"tag"`
	PullPolicy v1.PullPolicy `json:"pullPolicy"`
}

func (spec *ImageSpec) String() string {
	return spec.Repository + ":" + spec.Tag
}

type PravegaClusterStatus struct {
	// Fill me
}
