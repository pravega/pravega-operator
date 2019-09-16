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
			Template: MakeBookiePodTemplate(pravegaCluster),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForBookie(pravegaCluster),
			},
			VolumeClaimTemplates: makeBookieVolumeClaimTemplates(pravegaCluster.Spec.Bookkeeper),
		},
	}
}

func MakeBookiePodTemplate(p *v1alpha1.PravegaCluster) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      util.LabelsForBookie(p),
			Annotations: map[string]string{"pravega.version": p.Spec.Version},
		},
		Spec: *makeBookiePodSpec(p),
	}
}

func makeBookiePodSpec(p *v1alpha1.PravegaCluster) *corev1.PodSpec {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "bookie",
				Image:           util.BookkeeperImage(p),
				ImagePullPolicy: p.Spec.Bookkeeper.Image.PullPolicy,
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
						MountPath: "/bk/ledgers",
					},
					{
						Name:      JournalDiskName,
						MountPath: "/bk/journal",
					},
					{
						Name:      IndexDiskName,
						MountPath: "/bk/index",
					},
					{
						Name:      heapDumpName,
						MountPath: heapDumpDir,
					},
				},
				Resources: *p.Spec.Bookkeeper.Resources,
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
		Volumes: []corev1.Volume{
			{
				Name: heapDumpName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
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
	memoryOpts := []string{
		"-Xms1g",
		"-XX:MaxDirectMemorySize=1g",
		"-XX:+ExitOnOutOfMemoryError",
		"-XX:+CrashOnOutOfMemoryError",
		"-XX:+HeapDumpOnOutOfMemoryError",
		"-XX:HeapDumpPath=" + heapDumpDir,
	}

	if match, _ := util.CompareVersions(pravegaCluster.Spec.Version, "0.4.0", ">="); match {
		// Pravega < 0.4 uses a Java version that does not support the options below
		memoryOpts = append(memoryOpts,
			"-XX:+UnlockExperimentalVMOptions",
			"-XX:+UseCGroupMemoryLimitForHeap",
			"-XX:MaxRAMFraction=2",
		)
	}
	memoryOpts = util.OverrideDefaultJVMOptions(memoryOpts, pravegaCluster.Spec.Bookkeeper.BookkeeperJVMOptions.MemoryOpts)

	gcOpts := []string{
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
	gcOpts = util.OverrideDefaultJVMOptions(gcOpts, pravegaCluster.Spec.Bookkeeper.BookkeeperJVMOptions.GcOpts)

	gcLoggingOpts := []string{
		"-XX:+PrintGCDetails",
		"-XX:+PrintGCDateStamps",
		"-XX:+PrintGCApplicationStoppedTime",
		"-XX:+UseGCLogFileRotation",
		"-XX:NumberOfGCLogFiles=5",
		"-XX:GCLogFileSize=64m",
	}
	gcLoggingOpts = util.OverrideDefaultJVMOptions(gcLoggingOpts, pravegaCluster.Spec.Bookkeeper.BookkeeperJVMOptions.GcLoggingOpts)

	extraOpts := []string{}
	if pravegaCluster.Spec.Bookkeeper.BookkeeperJVMOptions.ExtraOpts != nil {
		extraOpts = pravegaCluster.Spec.Bookkeeper.BookkeeperJVMOptions.ExtraOpts
	}

	configData := map[string]string{
		"BOOKIE_MEM_OPTS":          strings.Join(memoryOpts, " "),
		"BOOKIE_GC_OPTS":           strings.Join(gcOpts, " "),
		"BOOKIE_GC_LOGGING_OPTS":   strings.Join(gcLoggingOpts, " "),
		"BOOKIE_EXTRA_OPTS":        strings.Join(extraOpts, " "),
		"ZK_URL":                   pravegaCluster.Spec.ZookeeperUri,
		"BK_useHostNameAsBookieID": "true",
		"PRAVEGA_CLUSTER_NAME":     pravegaCluster.ObjectMeta.Name,
		"WAIT_FOR":                 pravegaCluster.Spec.ZookeeperUri,
	}

	if match, _ := util.CompareVersions(pravegaCluster.Spec.Version, "0.5.0", "<"); match {
		// Pravega < 0.5 uses BookKeeper 4.5, which does not play well
		// with hostnames that resolve to different IP addresses over time
		configData["BK_useHostNameAsBookieID"] = "false"
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
