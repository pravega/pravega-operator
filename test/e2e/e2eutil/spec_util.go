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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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

// Returns a Job that can test pravega cluster by running a pravega example
func NewTestJob(namespace string) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "test-container",
							Image:           "openjdk:8",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/bin/sh", "-c"},
							Args:            []string{"/scripts/testScript.sh"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "test",
									MountPath: "/scripts",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "test-configmap",
									},
									DefaultMode: func(i int32) *int32 { return &i }(0777),
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
}

// Returns a configmap that stores the script for Job to use
func NewTestConfigMap(namespace string, controllerIp string) *corev1.ConfigMap {
	script := `
#!/bin/bash
git clone https://github.com/pravega/pravega-samples.git
cd pravega-samples/
./gradlew clean installDist
cd pravega-client-examples/build/install/pravega-client-examples/
./bin/helloWorldWriter -u tcp://` + controllerIp + `:9090
./bin/helloWorldReader -u tcp://` + controllerIp + `:9090
`
	configData := map[string]string{
		"testScript.sh": script,
	}
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "core/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: namespace,
		},
		Data: configData,
	}
}
