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

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/utils/k8sutil"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LedgerDiskName  = "ledger"
	JournalDiskName = "journal"
)

func deployBookie(pravegaCluster *v1alpha1.PravegaCluster) (err error) {

	err = sdk.Create(makeBookieHeadlessService(pravegaCluster))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = sdk.Create(makeBookieConfigMap(pravegaCluster))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = sdk.Create(makeBookieStatefulSet(pravegaCluster))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func destroyBookieVolumes(metadata metav1.ObjectMeta) {
	logrus.WithFields(logrus.Fields{"name": metadata.Name}).Info("Destroying Bookie volumes")

	err := k8sutil.DeleteCollection("v1", "PersistentVolumeClaim", metadata.Namespace, fmt.Sprintf("app=%v,kind=bookie", metadata.Name))
	if err != nil {
		logrus.Error(err)
	}
}

func makeBookieHeadlessService(pravegaCluster *v1alpha1.PravegaCluster) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.HeadlessServiceNameForBookie(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
			Labels:    k8sutil.LabelsForBookie(pravegaCluster),
			OwnerReferences: []metav1.OwnerReference{
				*k8sutil.AsOwnerRef(pravegaCluster),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "bookie",
					Port: 3181,
				},
			},
			Selector:  k8sutil.LabelsForBookie(pravegaCluster),
			ClusterIP: corev1.ClusterIPNone,
		},
	}
}

func makeBookieStatefulSet(pravegaCluster *v1alpha1.PravegaCluster) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.StatefulSetNameForBookie(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
			Labels:    k8sutil.LabelsForBookie(pravegaCluster),
			OwnerReferences: []metav1.OwnerReference{
				*k8sutil.AsOwnerRef(pravegaCluster),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         k8sutil.HeadlessServiceNameForBookie(pravegaCluster.Name),
			Replicas:            &pravegaCluster.Spec.Bookkeeper.Replicas,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Template:            makeBookieStatefulTemplate(pravegaCluster),
			Selector: &metav1.LabelSelector{
				MatchLabels: k8sutil.LabelsForBookie(pravegaCluster),
			},
			VolumeClaimTemplates: makeBookieVolumeClaimTemplates(&pravegaCluster.Spec.Bookkeeper),
		},
	}
}

func makeBookieStatefulTemplate(pravegaCluster *v1alpha1.PravegaCluster) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: k8sutil.LabelsForBookie(pravegaCluster),
		},
		Spec: *makeBookiePodSpec(pravegaCluster.Name, &pravegaCluster.Spec.Bookkeeper),
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
								Name: k8sutil.ConfigMapNameForBookie(clusterName),
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
			Spec: spec.Storage.JournalVolumeClaimTemplate,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: LedgerDiskName,
			},
			Spec: spec.Storage.LedgerVolumeClaimTemplate,
		},
	}
}

func makeBookieConfigMap(pravegaCluster *v1alpha1.PravegaCluster) *corev1.ConfigMap {
	configData := map[string]string{
		"BK_BOOKIE_EXTRA_OPTS":     "-Xms1g -Xmx1g -XX:MaxDirectMemorySize=1g -XX:+UseG1GC  -XX:MaxGCPauseMillis=10 -XX:+ParallelRefProcEnabled -XX:+UnlockExperimentalVMOptions -XX:+AggressiveOpts -XX:+DoEscapeAnalysis -XX:ParallelGCThreads=32 -XX:ConcGCThreads=32 -XX:G1NewSizePercent=50 -XX:+DisableExplicitGC -XX:-ResizePLAB",
		"ZK_URL":                   pravegaCluster.Spec.ZookeeperUri,
		"BK_useHostNameAsBookieID": "true",
		"PRAVEGA_CLUSTER_NAME":     pravegaCluster.ObjectMeta.Name,
		"WAIT_FOR":                 pravegaCluster.Spec.ZookeeperUri,
	}

	if pravegaCluster.Spec.Bookkeeper.AutoRecovery {
		configData["BK_AUTORECOVERY"] = "true"
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.ConfigMapNameForBookie(pravegaCluster.Name),
			Namespace: pravegaCluster.ObjectMeta.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*k8sutil.AsOwnerRef(pravegaCluster),
			},
		},
		Data: configData,
	}
}
