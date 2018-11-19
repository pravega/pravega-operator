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

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/utils/k8sutil"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

func deploySegmentStore(pravegaCluster *api.PravegaCluster) (err error) {
	err = sdk.Create(makeSegmentStoreHeadlessService(pravegaCluster))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	if pravegaCluster.Spec.ExternalAccess.Enabled {
		services := makeSegmentStoreExternalServices(pravegaCluster)
		for _, service := range services {
			err = sdk.Create(service)
			if err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
		}
	}

	err = sdk.Create(makeSegmentstoreConfigMap(pravegaCluster))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = sdk.Create(makeSegmentStoreStatefulSet(pravegaCluster))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func destroySegmentstoreCacheVolumes(metadata metav1.ObjectMeta) {
	logrus.WithFields(logrus.Fields{"name": metadata.Name}).Info("Destroying SegmentStore Cache volumes")

	err := k8sutil.DeleteCollection("v1", "PersistentVolumeClaim", metadata.Namespace, fmt.Sprintf("app=%v,kind=%v", metadata.Name, segmentStoreKind))
	if err != nil {
		logrus.Error(err)
	}
}

func makeSegmentStoreStatefulSet(pravegaCluster *api.PravegaCluster) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.StatefulSetNameForSegmentstore(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*k8sutil.AsOwnerRef(pravegaCluster),
			},
			Annotations: map[string]string{
				"service-per-pod-label": appsv1.StatefulSetPodNameLabel,
			},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         "pravega-segmentstore",
			Replicas:            &pravegaCluster.Spec.Pravega.SegmentStoreReplicas,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: k8sutil.LabelsForSegmentStore(pravegaCluster),
				},
				Spec: makeSegmentstorePodSpec(pravegaCluster),
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: k8sutil.LabelsForSegmentStore(pravegaCluster),
			},
			VolumeClaimTemplates: makeCacheVolumeClaimTemplate(&pravegaCluster.Spec.Pravega),
		},
	}
}

func makeSegmentstorePodSpec(pravegaCluster *api.PravegaCluster) corev1.PodSpec {
	environment := []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: k8sutil.ConfigMapNameForSegmentstore(pravegaCluster.Name),
				},
			},
		},
	}

	pravegaSpec := pravegaCluster.Spec.Pravega

	environment = configureTier2Secrets(environment, &pravegaSpec)

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
				Env:     k8sutil.DownwardAPIEnv(),
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      cacheVolumeName,
						MountPath: cacheVolumeMountPoint,
					},
				},
			},
		},
		Affinity: &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "component",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"pravega-segmentstore"},
								},
								{
									Key:      "pravega_cluster",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{pravegaCluster.Name},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}

	configureTier2Filesystem(&podSpec, &pravegaSpec)

	return podSpec
}

func makeSegmentstoreConfigMap(pravegaCluster *api.PravegaCluster) *corev1.ConfigMap {
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
		"CONTROLLER_URL":        k8sutil.PravegaControllerServiceURL(*pravegaCluster),
		"WAIT_FOR":              pravegaCluster.Spec.ZookeeperUri,
	}

	if pravegaCluster.Spec.ExternalAccess.Enabled {
		configData["K8_EXTERNAL_ACCESS"] = "true"
	}

	if pravegaCluster.Spec.Pravega.DebugLogging {
		configData["log.level"] = "DEBUG"
	}

	for k, v := range getTier2StorageOptions(&pravegaCluster.Spec.Pravega) {
		configData[k] = v
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.ConfigMapNameForSegmentstore(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
			Labels:    k8sutil.LabelsForSegmentStore(pravegaCluster),
			OwnerReferences: []metav1.OwnerReference{
				*k8sutil.AsOwnerRef(pravegaCluster),
			},
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
			Spec: pravegaSpec.CacheVolumeClaimTemplate,
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
				PersistentVolumeClaim: &pravegaSpec.Tier2.FileSystem.PersistentVolumeClaim,
			},
		})
	}
}

func makeSegmentStoreHeadlessService(pravegaCluster *api.PravegaCluster) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.HeadlessServiceNameForSegmentStore(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
			Labels:    k8sutil.LabelsForSegmentStore(pravegaCluster),
			OwnerReferences: []metav1.OwnerReference{
				*k8sutil.AsOwnerRef(pravegaCluster),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "server",
					Port:     12345,
					Protocol: "TCP",
				},
			},
			Selector:  k8sutil.LabelsForSegmentStore(pravegaCluster),
			ClusterIP: corev1.ClusterIPNone,
		},
	}
}

func makeSegmentStoreExternalServices(pravegaCluster *api.PravegaCluster) []*corev1.Service {
	var service *corev1.Service
	services := make([]*corev1.Service, pravegaCluster.Spec.Pravega.SegmentStoreReplicas)

	for i := int32(0); i < pravegaCluster.Spec.Pravega.SegmentStoreReplicas; i++ {
		service = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      k8sutil.ServiceNameForSegmentStore(pravegaCluster.Name, i),
				Namespace: pravegaCluster.Namespace,
				Labels:    k8sutil.LabelsForSegmentStore(pravegaCluster),
				OwnerReferences: []metav1.OwnerReference{
					*k8sutil.AsOwnerRef(pravegaCluster),
				},
			},
			Spec: corev1.ServiceSpec{
				Type:                  pravegaCluster.Spec.ExternalAccess.Type,
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
				Ports: []corev1.ServicePort{
					{
						Name:       "server",
						Port:       12345,
						Protocol:   "TCP",
						TargetPort: intstr.FromInt(12345),
					},
				},
				Selector: map[string]string{
					appsv1.StatefulSetPodNameLabel: fmt.Sprintf("%s-%d", k8sutil.StatefulSetNameForSegmentstore(pravegaCluster.Name), i),
				},
			},
		}
		services[i] = service
	}
	return services
}
