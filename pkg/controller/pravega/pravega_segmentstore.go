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
	"strings"

	"fmt"

	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	cacheVolumeName       = "cache"
	cacheVolumeMountPoint = "/tmp/pravega/cache"
	tier2FileMountPoint   = "/mnt/tier2"
	tier2VolumeName       = "tier2"
	segmentStoreKind      = "pravega-segmentstore"
)

func MakeSegmentStoreStatefulSet(pravegaCluster *api.PravegaCluster) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.StatefulSetNameForSegmentstore(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         "pravega-segmentstore",
			Replicas:            &pravegaCluster.Spec.Pravega.SegmentStoreReplicas,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: util.LabelsForSegmentStore(pravegaCluster),
				},
				Spec: makeSegmentstorePodSpec(pravegaCluster),
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForSegmentStore(pravegaCluster),
			},
			VolumeClaimTemplates: makeCacheVolumeClaimTemplate(pravegaCluster.Spec.Pravega),
		},
	}
}

func makeSegmentstorePodSpec(pravegaCluster *api.PravegaCluster) corev1.PodSpec {
	environment := []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: util.ConfigMapNameForSegmentstore(pravegaCluster.Name),
				},
			},
		},
	}

	pravegaSpec := pravegaCluster.Spec.Pravega

	environment = configureTier2Secrets(environment, pravegaSpec)

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "pravega-segmentstore",
				Image:           pravegaSpec.Image.String(),
				ImagePullPolicy: pravegaSpec.Image.PullPolicy,
				Args: []string{
					"segmentstore",
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "server",
						ContainerPort: 12345,
					},
				},
				EnvFrom: environment,
				Env:     util.DownwardAPIEnv(),
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      cacheVolumeName,
						MountPath: cacheVolumeMountPoint,
					},
				},
			},
		},
		Affinity: util.PodAntiAffinity("pravega-segmentstore", pravegaCluster.Name),
	}

	if pravegaSpec.SegmentStoreServiceAccountName != "" {
		podSpec.ServiceAccountName = pravegaSpec.SegmentStoreServiceAccountName
	}

	configureTier2Filesystem(&podSpec, pravegaSpec)

	return podSpec
}

func MakeSegmentstoreConfigMap(pravegaCluster *api.PravegaCluster) *corev1.ConfigMap {
	javaOpts := []string{
		"-Dpravegaservice.clusterName=" + pravegaCluster.Name,
	}

	for name, value := range pravegaCluster.Spec.Pravega.Options {
		javaOpts = append(javaOpts, fmt.Sprintf("-D%v=%v", name, value))
	}

	configData := map[string]string{
		"AUTHORIZATION_ENABLED": "false",
		"CLUSTER_NAME":          pravegaCluster.Name,
		"ZK_URL":                pravegaCluster.Spec.ZookeeperUri,
		"JAVA_OPTS":             strings.Join(javaOpts, " "),
		"CONTROLLER_URL":        util.PravegaControllerServiceURL(*pravegaCluster),
		"WAIT_FOR":              pravegaCluster.Spec.ZookeeperUri,
	}

	if pravegaCluster.Spec.ExternalAccess.Enabled {
		configData["K8_EXTERNAL_ACCESS"] = "true"
	}

	if pravegaCluster.Spec.Pravega.DebugLogging {
		configData["log.level"] = "DEBUG"
	}

	for k, v := range getTier2StorageOptions(pravegaCluster.Spec.Pravega) {
		configData[k] = v
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.ConfigMapNameForSegmentstore(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
			Labels:    util.LabelsForSegmentStore(pravegaCluster),
		},
		Data: configData,
	}
}

func makeCacheVolumeClaimTemplate(pravegaSpec *api.PravegaSpec) []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: cacheVolumeName,
			},
			Spec: *pravegaSpec.CacheVolumeClaimTemplate,
		},
	}
}

func getTier2StorageOptions(pravegaSpec *api.PravegaSpec) map[string]string {
	if pravegaSpec.Tier2.FileSystem != nil {
		return map[string]string{
			"TIER2_STORAGE": "FILESYSTEM",
			"NFS_MOUNT":     tier2FileMountPoint,
		}
	}

	if pravegaSpec.Tier2.Ecs != nil {
		// EXTENDEDS3_ACCESS_KEY_ID & EXTENDEDS3_SECRET_KEY will come from secret storage
		return map[string]string{
			"TIER2_STORAGE":        "EXTENDEDS3",
			"EXTENDEDS3_BUCKET":    pravegaSpec.Tier2.Ecs.Bucket,
			"EXTENDEDS3_URI":       pravegaSpec.Tier2.Ecs.Uri,
			"EXTENDEDS3_ROOT":      pravegaSpec.Tier2.Ecs.Root,
			"EXTENDEDS3_NAMESPACE": pravegaSpec.Tier2.Ecs.Namespace,
		}
	}

	if pravegaSpec.Tier2.Hdfs != nil {
		return map[string]string{
			"TIER2_STORAGE": "HDFS",
			"HDFS_URL":      pravegaSpec.Tier2.Hdfs.Uri,
			"HDFS_ROOT":     pravegaSpec.Tier2.Hdfs.Root,
		}
	}

	return make(map[string]string)
}

func configureTier2Secrets(environment []corev1.EnvFromSource, pravegaSpec *api.PravegaSpec) []corev1.EnvFromSource {
	if pravegaSpec.Tier2.Ecs != nil {
		return append(environment, corev1.EnvFromSource{
			Prefix: "EXTENDEDS3_",
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: pravegaSpec.Tier2.Ecs.Credentials,
				},
			},
		})
	}

	return environment
}

func configureTier2Filesystem(podSpec *corev1.PodSpec, pravegaSpec *api.PravegaSpec) {

	if pravegaSpec.Tier2.FileSystem != nil {
		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      tier2VolumeName,
			MountPath: tier2FileMountPoint,
		})

		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name: tier2VolumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: pravegaSpec.Tier2.FileSystem.PersistentVolumeClaim,
			},
		})
	}
}

func MakeSegmentStoreHeadlessService(pravegaCluster *api.PravegaCluster) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.HeadlessServiceNameForSegmentStore(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
			Labels:    util.LabelsForSegmentStore(pravegaCluster),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "server",
					Port:     12345,
					Protocol: "TCP",
				},
			},
			Selector:  util.LabelsForSegmentStore(pravegaCluster),
			ClusterIP: corev1.ClusterIPNone,
		},
	}
}

func MakeSegmentStoreExternalServices(pravegaCluster *api.PravegaCluster) []*corev1.Service {
	var service *corev1.Service
	services := make([]*corev1.Service, pravegaCluster.Spec.Pravega.SegmentStoreReplicas)

	for i := int32(0); i < pravegaCluster.Spec.Pravega.SegmentStoreReplicas; i++ {
		service = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      util.ServiceNameForSegmentStore(pravegaCluster.Name, i),
				Namespace: pravegaCluster.Namespace,
				Labels:    util.LabelsForSegmentStore(pravegaCluster),
			},
			Spec: corev1.ServiceSpec{
				Type: pravegaCluster.Spec.ExternalAccess.Type,
				Ports: []corev1.ServicePort{
					{
						Name:       "server",
						Port:       12345,
						Protocol:   "TCP",
						TargetPort: intstr.FromInt(12345),
					},
				},
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
				Selector: map[string]string{
					appsv1.StatefulSetPodNameLabel: fmt.Sprintf("%s-%d", util.StatefulSetNameForSegmentstore(pravegaCluster.Name), i),
				},
			},
		}
		services[i] = service
	}
	return services
}

func MakeSegmentstorePodDisruptionBudget(pravegaCluster *api.PravegaCluster) *policyv1beta1.PodDisruptionBudget {
	var maxUnavailable intstr.IntOrString

	if pravegaCluster.Spec.Pravega.SegmentStoreReplicas == int32(1) {
		maxUnavailable = intstr.FromInt(0)
	} else {
		maxUnavailable = intstr.FromInt(1)
	}

	return &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.PdbNameForSegmentstore(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForSegmentStore(pravegaCluster),
			},
		},
	}
}
