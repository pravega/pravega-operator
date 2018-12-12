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
			Template:            makeBookieStatefulTemplate(pravegaCluster),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForBookie(pravegaCluster),
			},
			VolumeClaimTemplates: makeBookieVolumeClaimTemplates(pravegaCluster.Spec.Bookkeeper),
		},
	}
}

func makeBookieStatefulTemplate(pravegaCluster *v1alpha1.PravegaCluster) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: util.LabelsForBookie(pravegaCluster),
		},
		Spec: *makeBookiePodSpec(pravegaCluster.Name, pravegaCluster.Spec.Bookkeeper),
	}
}

func makeBookiePodSpec(clusterName string, bookkeeperSpec *v1alpha1.BookkeeperSpec) *corev1.PodSpec {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "bookie",
				Image:           bookkeeperSpec.Image.String(),
				ImagePullPolicy: bookkeeperSpec.Image.PullPolicy,
				Command: []string{
					"/bin/bash", "/opt/bookkeeper/entrypoint.sh",
				},
				Args: []string{
					"/opt/bookkeeper/bin/bookkeeper", "bookie",
				},
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
								Name: util.ConfigMapNameForBookie(clusterName),
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
				},
			},
		},
		Affinity: util.PodAntiAffinity("bookie", clusterName),
	}

	if bookkeeperSpec.ServiceAccountName != "" {
		podSpec.ServiceAccountName = bookkeeperSpec.ServiceAccountName
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
	}
}

func MakeBookieConfigMap(pravegaCluster *v1alpha1.PravegaCluster) *corev1.ConfigMap {
	configData := map[string]string{
		"BK_BOOKIE_EXTRA_OPTS":     "-Xms1g -Xmx1g -XX:MaxDirectMemorySize=1g -XX:+UseG1GC  -XX:MaxGCPauseMillis=10 -XX:+ParallelRefProcEnabled -XX:+UnlockExperimentalVMOptions -XX:+AggressiveOpts -XX:+DoEscapeAnalysis -XX:ParallelGCThreads=32 -XX:ConcGCThreads=32 -XX:G1NewSizePercent=50 -XX:+DisableExplicitGC -XX:-ResizePLAB",
		"ZK_URL":                   pravegaCluster.Spec.ZookeeperUri,
		"BK_useHostNameAsBookieID": "true",
		"PRAVEGA_CLUSTER_NAME":     pravegaCluster.ObjectMeta.Name,
		"WAIT_FOR":                 pravegaCluster.Spec.ZookeeperUri,
	}

	if *pravegaCluster.Spec.Bookkeeper.AutoRecovery {
		configData["BK_AUTORECOVERY"] = "true"
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
