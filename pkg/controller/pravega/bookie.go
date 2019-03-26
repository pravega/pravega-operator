/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package pravega

import (
	"fmt"
	"strings"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	LedgerDiskName  = "ledger"
	JournalDiskName = "journal"
	IndexDiskName   = "index"
)

func MakeBookieHeadlessService(pravegaCluster *v1alpha1.PravegaCluster) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.HeadlessServiceNameForBookie(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
			Labels:    util.LabelsForBookie(pravegaCluster),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "bookie",
					Port: 3181,
				},
			},
			Selector:  util.LabelsForBookie(pravegaCluster),
			ClusterIP: corev1.ClusterIPNone,
		},
	}
}

func MakeBookieStatefulSet(pravegaCluster *v1alpha1.PravegaCluster) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.StatefulSetNameForBookie(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
			Labels:    util.LabelsForBookie(pravegaCluster),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         util.HeadlessServiceNameForBookie(pravegaCluster.Name),
			Replicas:            &pravegaCluster.Spec.Bookkeeper.Replicas,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
			Template: makeBookieStatefulTemplate(pravegaCluster),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForBookie(pravegaCluster),
			},
			VolumeClaimTemplates: makeBookieVolumeClaimTemplates(pravegaCluster.Spec.Bookkeeper),
		},
	}
}

func makeBookieStatefulTemplate(p *v1alpha1.PravegaCluster) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      util.LabelsForBookie(p),
			Annotations: map[string]string{"pravega.version": p.Spec.Version},
		},
		Spec: *makeBookiePodSpec(p),
	}
}

func makeBookiePodSpec(p *v1alpha1.PravegaCluster) *corev1.PodSpec {
	terminationGracePeriodSeconds := int64(0)
	podSpec := &corev1.PodSpec{
		TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
		Containers: []corev1.Container{
			{
				Name: "bookie",
				Image: fmt.Sprintf("%s:%s",
					p.Spec.Bookkeeper.ImageRepository,
					p.Spec.Version),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Ports: []corev1.ContainerPort{
					{
						Name:          "bookie",
						ContainerPort: 3181,
					},
				},
				EnvFrom: []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: util.ConfigMapNameForBookie(p.Name),
							},
						},
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      LedgerDiskName,
						MountPath: "/bk/journal",
					},
					{
						Name:      JournalDiskName,
						MountPath: "/bk/ledgers",
					},
					{
						Name:      IndexDiskName,
						MountPath: "/bk/index",
					},
				},
				Resources: *bookkeeperSpec.Resources,
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/bin/sh", "-c", "/opt/bookkeeper/bin/bookkeeper shell bookiesanity"},
						},
					},
					// Bookie pods should start fast. We give it up to 1.5 minute to become ready.
					InitialDelaySeconds: 20,
					PeriodSeconds:       10,
					FailureThreshold:    9,
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: util.HealthcheckCommand(3181),
						},
					},
					// We start the liveness probe from the maximum time the pod can take
					// before becoming ready.
					// If the pod fails the health check during 1 minute, Kubernetes
					// will restart it.
					InitialDelaySeconds: 60,
					PeriodSeconds:       15,
					FailureThreshold:    4,
				},
			},
		},
		Affinity: util.PodAntiAffinity("bookie", p.Name),
	}

	if p.Spec.Bookkeeper.ServiceAccountName != "" {
		podSpec.ServiceAccountName = p.Spec.Bookkeeper.ServiceAccountName
	}

	return podSpec
}

func makeBookieVolumeClaimTemplates(spec *v1alpha1.BookkeeperSpec) []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: JournalDiskName,
			},
			Spec: *spec.Storage.JournalVolumeClaimTemplate,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: LedgerDiskName,
			},
			Spec: *spec.Storage.LedgerVolumeClaimTemplate,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: IndexDiskName,
			},
			Spec: *spec.Storage.IndexVolumeClaimTemplate,
		},
	}
}

func MakeBookieConfigMap(pravegaCluster *v1alpha1.PravegaCluster) *corev1.ConfigMap {
	javaOpts := []string{
		"-Xms1g",
		"-XX:+UnlockExperimentalVMOptions",
		"-XX:+UseCGroupMemoryLimitForHeap",
		"-XX:MaxRAMFraction=1",
		"-XX:MaxDirectMemorySize=1g",
		"-XX:+UseG1GC",
		"-XX:MaxGCPauseMillis=10",
		"-XX:+ParallelRefProcEnabled",
		"-XX:+AggressiveOpts",
		"-XX:+DoEscapeAnalysis",
		"-XX:ParallelGCThreads=32",
		"-XX:ConcGCThreads=32",
		"-XX:G1NewSizePercent=50",
		"-XX:+DisableExplicitGC",
		"-XX:-ResizePLAB",
	}

	configData := map[string]string{
		"BK_BOOKIE_EXTRA_OPTS": strings.Join(javaOpts, " "),
		"ZK_URL":               pravegaCluster.Spec.ZookeeperUri,
		// Set useHostNameAsBookieID to false until BookKeeper Docker
		// image is updated to 4.7
		// This value can be explicitly overridden when using the operator
		// with images based on BookKeeper 4.7 or newer
		"BK_useHostNameAsBookieID": "false",
		"PRAVEGA_CLUSTER_NAME":     pravegaCluster.ObjectMeta.Name,
		"WAIT_FOR":                 pravegaCluster.Spec.ZookeeperUri,
	}

	if *pravegaCluster.Spec.Bookkeeper.AutoRecovery {
		configData["BK_AUTORECOVERY"] = "true"
		// Wait one minute before starting autorecovery. This will give
		// pods some time to come up after being updated or migrated
		configData["BK_lostBookieRecoveryDelay"] = "60"
	}

	for k, v := range pravegaCluster.Spec.Bookkeeper.Options {
		prefixKey := fmt.Sprintf("BK_%s", k)
		configData[prefixKey] = v
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.ConfigMapNameForBookie(pravegaCluster.Name),
			Namespace: pravegaCluster.ObjectMeta.Namespace,
		},
		Data: configData,
	}
}

func MakeBookiePodDisruptionBudget(pravegaCluster *v1alpha1.PravegaCluster) *policyv1beta1.PodDisruptionBudget {
	maxUnavailable := intstr.FromInt(1)
	return &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.PdbNameForBookie(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForBookie(pravegaCluster),
			},
		},
	}
}
