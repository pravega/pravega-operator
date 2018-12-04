/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package util

import (
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

func AsOwnerRef(pravegaCluster *v1alpha1.PravegaCluster) *metav1.OwnerReference {
	boolTrue := true
	return &metav1.OwnerReference{
		APIVersion: v1alpha1.APIVERSION,
		Kind:       v1alpha1.KIND,
		Name:       pravegaCluster.Name,
		UID:        pravegaCluster.UID,
		Controller: &boolTrue,
		BlockOwnerDeletion: &boolTrue,
	}
}

// GetWatchNamespaceAllowBlank returns the namespace the operator should be watching for changes
func GetWatchNamespaceAllowBlank() (string, error) {
	ns, found := os.LookupEnv(k8sutil.WatchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", k8sutil.WatchNamespaceEnvVar)
	}
	return ns, nil
}

func DownwardAPIEnv() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
	}
}
