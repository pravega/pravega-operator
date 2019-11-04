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

	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	externalDNSAnnotationKey = "external-dns.alpha.kubernetes.io/hostname"
	dot                      = "."
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
			Labels:    util.LabelsForSegmentStore(pravegaCluster),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         "pravega-segmentstore",
			Replicas:            &pravegaCluster.Spec.Pravega.SegmentStoreReplicas,
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
			Template: MakeSegmentStorePodTemplate(pravegaCluster),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForSegmentStore(pravegaCluster),
			},
			VolumeClaimTemplates: makeCacheVolumeClaimTemplate(pravegaCluster.Spec.Pravega),
		},
	}
}

func MakeSegmentStorePodTemplate(p *api.PravegaCluster) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      util.LabelsForSegmentStore(p),
			Annotations: map[string]string{"pravega.version": p.Spec.Version},
		},
		Spec: makeSegmentstorePodSpec(p),
	}
}

func makeSegmentstorePodSpec(p *api.PravegaCluster) corev1.PodSpec {
	environment := []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: util.ConfigMapNameForSegmentstore(p.Name),
				},
			},
		},
	}

	environment = configureTier2Secrets(environment, p.Spec.Pravega)

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "pravega-segmentstore",
				Image:           util.PravegaImage(p),
				ImagePullPolicy: p.Spec.Pravega.Image.PullPolicy,
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
					{
						Name:      heapDumpName,
						MountPath: heapDumpDir,
					},
				},
				Resources: *p.Spec.Pravega.SegmentStoreResources,
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: util.HealthcheckCommand(12345),
						},
					},
					// Segment Stores can take a few minutes to become ready when the cluster
					// is configured with external enabled as they need to wait for the allocation
					// of the external IP address.
					// This config gives it up to 5 minutes to become ready.
					PeriodSeconds:    10,
					FailureThreshold: 30,
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: util.HealthcheckCommand(12345),
						},
					},
					// In the readiness probe we allow the pod to take up to 5 minutes
					// to become ready. Therefore, the liveness probe will give it
					// a 5-minute grace period before starting monitoring the container.
					// If the pod fails the health check during 1 minute, Kubernetes
					// will restart it.
					InitialDelaySeconds: 300,
					PeriodSeconds:       15,
					FailureThreshold:    4,
				},
			},
		},
		Affinity: util.PodAntiAffinity("pravega-segmentstore", p.Name),
		Volumes: []corev1.Volume{
			{
				Name: heapDumpName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
	}

	if p.Spec.Pravega.SegmentStoreServiceAccountName != "" {
		podSpec.ServiceAccountName = p.Spec.Pravega.SegmentStoreServiceAccountName
	}

	configureSegmentstoreTLSSecret(&podSpec, p)

	configureTier2Filesystem(&podSpec, p.Spec.Pravega)

	return podSpec
}

func MakeSegmentstoreConfigMap(p *api.PravegaCluster) *corev1.ConfigMap {
	javaOpts := []string{
		"-Dpravegaservice.clusterName=" + p.Name,
	}

	jvmOpts := []string{
		"-Xms1g",
		"-XX:+ExitOnOutOfMemoryError",
		"-XX:+CrashOnOutOfMemoryError",
		"-XX:+HeapDumpOnOutOfMemoryError",
		"-XX:HeapDumpPath=" + heapDumpDir,
		"-Dpravegaservice.clusterName=" + p.Name,
	}

	if match, _ := util.CompareVersions(p.Spec.Version, "0.4.0", ">="); match {
		// Pravega < 0.4 uses a Java version that does not support the options below
		jvmOpts = append(jvmOpts,
			"-XX:+UnlockExperimentalVMOptions",
			"-XX:+UseCGroupMemoryLimitForHeap",
			"-XX:MaxRAMFraction=2",
		)
	}

	javaOpts = append(javaOpts, util.OverrideDefaultJVMOptions(jvmOpts, p.Spec.Pravega.SegmentStoreJVMOptions)...)

	for name, value := range p.Spec.Pravega.Options {
		javaOpts = append(javaOpts, fmt.Sprintf("-D%v=%v", name, value))
	}
	authEnabledStr := fmt.Sprint(p.Spec.Authentication.IsEnabled())
	configData := map[string]string{
		"AUTHORIZATION_ENABLED": authEnabledStr,
		"CLUSTER_NAME":          p.Name,
		"ZK_URL":                p.Spec.ZookeeperUri,
		"JAVA_OPTS":             strings.Join(javaOpts, " "),
		"CONTROLLER_URL":        util.PravegaControllerServiceURL(*p),
	}

	// Wait for at least 3 Bookies to come up
	var waitFor []string
	for i := int32(0); i < util.Min(3, p.Spec.Bookkeeper.Replicas); i++ {
		waitFor = append(waitFor,
			fmt.Sprintf("%s-%d.%s.%s:3181",
				util.StatefulSetNameForBookie(p.Name),
				i,
				util.HeadlessServiceNameForBookie(p.Name),
				p.Namespace))
	}
	configData["WAIT_FOR"] = strings.Join(waitFor, ",")

	if p.Spec.ExternalAccess.Enabled {
		configData["K8_EXTERNAL_ACCESS"] = "true"
	}

	if p.Spec.Pravega.DebugLogging {
		configData["log.level"] = "DEBUG"
	}

	for k, v := range getTier2StorageOptions(p.Spec.Pravega) {
		configData[k] = v
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.ConfigMapNameForSegmentstore(p.Name),
			Namespace: p.Namespace,
			Labels:    util.LabelsForSegmentStore(p),
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

func configureSegmentstoreTLSSecret(podSpec *corev1.PodSpec, p *api.PravegaCluster) {
	if p.Spec.TLS.IsSecureSegmentStore() {
		vol := corev1.Volume{
			Name: tlsVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: p.Spec.TLS.Static.SegmentStoreSecret,
				},
			},
		}
		podSpec.Volumes = append(podSpec.Volumes, vol)

		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      tlsVolumeName,
			MountPath: tlsMountDir,
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

func getSSServiceType(pravegaCluster *api.PravegaCluster) (serviceType corev1.ServiceType) {
	if pravegaCluster.Spec.Pravega.SegmentStoreExternalServiceType == "" {
		if pravegaCluster.Spec.ExternalAccess.Type == "" {
			return api.DefaultServiceType
		}
		return pravegaCluster.Spec.ExternalAccess.Type
	}
	return pravegaCluster.Spec.Pravega.SegmentStoreExternalServiceType
}

func cloneMap(sourceMap map[string]string) (annotationMap map[string]string) {
	if len(sourceMap) == 0 {
		return map[string]string{}
	}
	annotationMap = make(map[string]string, len(sourceMap)+1)
	for key, value := range sourceMap {
		annotationMap[key] = value
	}
	return annotationMap
}

func generateDNSAnnotationForSvc(domainName string, podName string) (dnsAnnotationValue string) {
	var ssFQDN string
	if domainName != "" {
		domain := strings.TrimSpace(domainName)
		if strings.HasSuffix(domain, dot) {
			ssFQDN = podName + dot + domain
		} else {
			ssFQDN = podName + dot + domain + dot
		}
	}
	return ssFQDN
}

func MakeSegmentStoreExternalServices(pravegaCluster *api.PravegaCluster) []*corev1.Service {
	var service *corev1.Service

	serviceType := getSSServiceType(pravegaCluster)
	services := make([]*corev1.Service, pravegaCluster.Spec.Pravega.SegmentStoreReplicas)

	for i := int32(0); i < pravegaCluster.Spec.Pravega.SegmentStoreReplicas; i++ {
		ssPodName := util.ServiceNameForSegmentStore(pravegaCluster.Name, i)
		annotationMap := pravegaCluster.Spec.Pravega.SegmentStoreServiceAnnotations
		annotationValue := generateDNSAnnotationForSvc(pravegaCluster.Spec.ExternalAccess.DomainName, ssPodName)

		if annotationValue != "" {
			annotationMap = cloneMap(pravegaCluster.Spec.Pravega.SegmentStoreServiceAnnotations)
			annotationMap[externalDNSAnnotationKey] = annotationValue
		}

		service = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        ssPodName,
				Namespace:   pravegaCluster.Namespace,
				Labels:      util.LabelsForSegmentStore(pravegaCluster),
				Annotations: annotationMap,
			},
			Spec: corev1.ServiceSpec{
				Type: serviceType,
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
