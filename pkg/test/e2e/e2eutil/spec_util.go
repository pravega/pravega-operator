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
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"

	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDefaultCluster returns a cluster with some of the fields already filled in spec and the remaining fields
// will be filled with default values
func NewDefaultCluster(namespace string) *api.PravegaCluster {
	return &api.PravegaCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PravegaCluster",
			APIVersion: "pravega.pravega.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pravega",
			Namespace: namespace,
		},
		Spec: api.ClusterSpec{
			Pravega: &api.PravegaSpec{
				Options:                map[string]string{"pravegaservice.cache.size.max": "1610612736"},
				SegmentStoreJVMOptions: []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m"},
				SegmentStoreResources: &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				},
			},
		},
	}
}

func NewClusterWithVersion(namespace, version string) *api.PravegaCluster {
	cluster := NewDefaultCluster(namespace)
	cluster.Spec = api.ClusterSpec{
		Version: version,
		Pravega: &api.PravegaSpec{
			Options:                map[string]string{"pravegaservice.cache.size.max": "1610612736"},
			SegmentStoreJVMOptions: []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m"},
			SegmentStoreResources: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1000m"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2000m"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
			},
		},
	}
	return cluster
}

func NewClusterWithNoSegmentStoreMemoryLimits(namespace string) *api.PravegaCluster {
	return &api.PravegaCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PravegaCluster",
			APIVersion: "pravega.pravega.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pravega",
			Namespace: namespace,
		},
		Spec: api.ClusterSpec{
			Pravega: &api.PravegaSpec{
				Options:                map[string]string{"pravegaservice.cache.size.max": "1610612736"},
				SegmentStoreJVMOptions: []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m"},
				SegmentStoreResources: &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("2000m"),
					},
				},
			},
		},
	}
}

func newTestJob(namespace string, command string) *batchv1.Job {
	deadline := int64(180)
	retries := int32(1)
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-job-",
			Namespace:    namespace,
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds: &deadline,
			BackoffLimit:          &retries,

			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "test-container",
							Image:           "adrianmo/pravega-samples",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/bin/sh", "-c"},
							Args:            []string{command},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
}

// NewTestWriteReadJob returns a Job that can test pravega cluster by running a sample
func NewTestWriteReadJob(namespace string, controllerUri string) *batchv1.Job {
	command := fmt.Sprintf("cd /samples/pravega-client-examples "+
		"&& bin/helloWorldWriter -u tcp://%s:9090 "+
		"&& bin/helloWorldReader -u tcp://%s:9090",
		controllerUri, controllerUri)
	return newTestJob(namespace, command)
}

func NewTier2(namespace string) *corev1.PersistentVolumeClaim {
	storageName := "nfs"
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pravega-tier2",
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageName,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.PersistentVolumeAccessMode(corev1.ReadWriteMany)},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("5Gi"),
				},
			},
		},
	}
}
