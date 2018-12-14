/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package e2eutil

import (
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDefaultCluster returns a cluster with an empty spec, which will be filled
// with default values
func NewDefaultCluster(namespace string) *api.PravegaCluster {
	return &api.PravegaCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PravegaCluster",
			APIVersion: "pravega.pravega.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
	}
}
