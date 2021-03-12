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
	"sort"
	"strings"

	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	"github.com/pravega/pravega-operator/pkg/controller/config"
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

func MakeSegmentStoreStatefulSet(p *api.PravegaCluster) *appsv1.StatefulSet {
	statefulSet := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.StatefulSetNameForSegmentstore(),
			Namespace: p.Namespace,
			Labels:    p.LabelsForSegmentStore(),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         "pravega-segmentstore",
			Replicas:            &p.Spec.Pravega.SegmentStoreReplicas,
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.OnDeleteStatefulSetStrategyType,
			},
			Template: MakeSegmentStorePodTemplate(p),
			Selector: &metav1.LabelSelector{
				MatchLabels: p.LabelsForSegmentStore(),
			},
		},
	}
	if util.IsVersionBelow07(p.Spec.Version) {
		statefulSet.Spec.VolumeClaimTemplates = makeCacheVolumeClaimTemplate(p)
	}
	return statefulSet
}

func MakeSegmentStorePodTemplate(p *api.PravegaCluster) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      p.LabelsForSegmentStore(),
			Annotations: map[string]string{"pravega.version": p.Spec.Version},
		},
		Spec: makeSegmentstorePodSpec(p),
	}
}

func makeSegmentstorePodSpec(p *api.PravegaCluster) corev1.PodSpec {
	configMapName := strings.TrimSpace(p.Spec.Pravega.SegmentStoreEnvVars)
	secret := p.Spec.Pravega.SegmentStoreSecret
	environment := []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: p.ConfigMapNameForSegmentstore(),
				},
			},
		},
	}
	if configMapName != "" {
		environment = append(environment, corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		})
	}
	if strings.TrimSpace(secret.Secret) != "" && strings.TrimSpace(secret.MountPath) == "" {
		environment = append(environment, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: strings.TrimSpace(secret.Secret),
				},
			},
		})
	}

	environment = configureTier2Secrets(environment, p.Spec.Pravega)
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "pravega-segmentstore",
				Image:           p.PravegaImage(),
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
				EnvFrom:      environment,
				Env:          util.DownwardAPIEnv(),
				VolumeMounts: MakeSegmentStoreVolumeMount(p),
				Resources:    *p.Spec.Pravega.SegmentStoreResources,
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
				SecurityContext: &corev1.SecurityContext{
					Privileged: &config.TestMode,
				},
			},
		},
		Affinity: p.Spec.Pravega.SegmentStorePodAffinity,
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

	if p.Spec.Pravega.SegmentStoreSecurityContext != nil {
		podSpec.SecurityContext = p.Spec.Pravega.SegmentStoreSecurityContext
	}

	configureSegmentstoreSecret(&podSpec, p)

	configureSegmentstoreTLSSecret(&podSpec, p)

	configureCaBundleSecret(&podSpec, p)

	configureLTSFilesystem(&podSpec, p.Spec.Pravega)

	configureSegmentstoreAuthSecret(&podSpec, p)

	return podSpec
}

func MakeSegmentStoreVolumeMount(p *api.PravegaCluster) []corev1.VolumeMount {
	volumeMount := []corev1.VolumeMount{
		{
			Name:      heapDumpName,
			MountPath: heapDumpDir,
		},
	}
	if util.IsVersionBelow07(p.Spec.Version) {
		volumeMount = append(volumeMount, corev1.VolumeMount{
			Name:      cacheVolumeName,
			MountPath: cacheVolumeMountPoint,
		})
	}
	return volumeMount
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
			"-XX:+UseContainerSupport",
			"-XX:MaxRAMPercentage=50.0",
		)
	}

	javaOpts = append(javaOpts, util.OverrideDefaultJVMOptions(jvmOpts, p.Spec.Pravega.SegmentStoreJVMOptions)...)

	for name, value := range p.Spec.Pravega.Options {
		javaOpts = append(javaOpts, fmt.Sprintf("-D%v=%v", name, value))
	}

	sort.Strings(javaOpts)

	authEnabledStr := fmt.Sprint(p.Spec.Authentication.IsEnabled())
	configData := map[string]string{
		"AUTHORIZATION_ENABLED": authEnabledStr,
		"CLUSTER_NAME":          p.Name,
		"ZK_URL":                p.Spec.ZookeeperUri,
		"JAVA_OPTS":             strings.Join(javaOpts, " "),
		"CONTROLLER_URL":        p.PravegaControllerServiceURL(),
	}

	// Wait for at least 3 Bookies to come up
	configData["WAIT_FOR"] = p.Spec.BookkeeperUri

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
			Name:      p.ConfigMapNameForSegmentstore(),
			Namespace: p.Namespace,
			Labels:    p.LabelsForSegmentStore(),
		},
		Data: configData,
	}
}

func makeCacheVolumeClaimTemplate(p *api.PravegaCluster) []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cacheVolumeName,
				Namespace: p.Namespace,
			},
			Spec: *p.Spec.Pravega.CacheVolumeClaimTemplate,
		},
	}
}

func getTier2StorageOptions(pravegaSpec *api.PravegaSpec) map[string]string {
	if pravegaSpec.LongTermStorage.FileSystem != nil {
		return map[string]string{
			"TIER2_STORAGE": "FILESYSTEM",
			"NFS_MOUNT":     ltsFileMountPoint,
		}
	}

	if pravegaSpec.LongTermStorage.Ecs != nil {
		// EXTENDEDS3_ACCESS_KEY_ID & EXTENDEDS3_SECRET_KEY will come from secret storage
		return map[string]string{
			"TIER2_STORAGE":        "EXTENDEDS3",
			"EXTENDEDS3_CONFIGURI": pravegaSpec.LongTermStorage.Ecs.ConfigUri,
			"EXTENDEDS3_BUCKET":    pravegaSpec.LongTermStorage.Ecs.Bucket,
			"EXTENDEDS3_PREFIX":    pravegaSpec.LongTermStorage.Ecs.Prefix,
		}
	}

	if pravegaSpec.LongTermStorage.Hdfs != nil {
		return map[string]string{
			"TIER2_STORAGE": "HDFS",
			"HDFS_URL":      pravegaSpec.LongTermStorage.Hdfs.Uri,
			"HDFS_ROOT":     pravegaSpec.LongTermStorage.Hdfs.Root,
		}
	}

	return make(map[string]string)
}

func configureTier2Secrets(environment []corev1.EnvFromSource, pravegaSpec *api.PravegaSpec) []corev1.EnvFromSource {
	if pravegaSpec.LongTermStorage.Ecs != nil {
		return append(environment, corev1.EnvFromSource{
			Prefix: "EXTENDEDS3_",
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: pravegaSpec.LongTermStorage.Ecs.Credentials,
				},
			},
		})
	}

	return environment
}

func configureLTSFilesystem(podSpec *corev1.PodSpec, pravegaSpec *api.PravegaSpec) {

	if pravegaSpec.LongTermStorage.FileSystem != nil {
		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      ltsVolumeName,
			MountPath: ltsFileMountPoint,
		})

		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name: ltsVolumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: pravegaSpec.LongTermStorage.FileSystem.PersistentVolumeClaim,
			},
		})
	}
}

func configureSegmentstoreSecret(podSpec *corev1.PodSpec, p *api.PravegaCluster) {
	secret := p.Spec.Pravega.SegmentStoreSecret
	if strings.TrimSpace(secret.Secret) != "" && strings.TrimSpace(secret.MountPath) != "" {
		vol := corev1.Volume{
			Name: ssSecretVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: strings.TrimSpace(secret.Secret),
				},
			},
		}
		podSpec.Volumes = append(podSpec.Volumes, vol)

		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      ssSecretVolumeName,
			MountPath: strings.TrimSpace(secret.MountPath),
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

func configureSegmentstoreAuthSecret(podSpec *corev1.PodSpec, p *api.PravegaCluster) {
	if p.Spec.Authentication.SegmentStoreTokenSecret != "" {
		vol := corev1.Volume{
			Name: ssAuthVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: p.Spec.Authentication.SegmentStoreTokenSecret,
				},
			},
		}
		podSpec.Volumes = append(podSpec.Volumes, vol)

		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      ssAuthVolumeName,
			MountPath: ssAuthMountDir,
		})
	}
}

func configureCaBundleSecret(podSpec *corev1.PodSpec, p *api.PravegaCluster) {
	if p.Spec.TLS.IsCaBundlePresent() {
		vol := corev1.Volume{
			Name: caBundleVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: p.Spec.TLS.Static.CaBundle,
				},
			},
		}
		podSpec.Volumes = append(podSpec.Volumes, vol)

		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      caBundleVolumeName,
			MountPath: caBundleMountDir,
		})
	}
}

func MakeSegmentStoreHeadlessService(p *api.PravegaCluster) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.HeadlessServiceNameForSegmentStore(),
			Namespace: p.Namespace,
			Labels:    p.LabelsForSegmentStore(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "server",
					Port:     12345,
					Protocol: "TCP",
				},
			},
			Selector:  p.LabelsForSegmentStore(),
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

func MakeSegmentStoreExternalServices(p *api.PravegaCluster) []*corev1.Service {
	var service *corev1.Service
	serviceType := getSSServiceType(p)
	services := make([]*corev1.Service, p.Spec.Pravega.SegmentStoreReplicas)
	for i := int32(0); i < p.Spec.Pravega.SegmentStoreReplicas; i++ {
		ssPodName := p.ServiceNameForSegmentStore(i)
		annotationMap := p.Spec.Pravega.SegmentStoreServiceAnnotations
		annotationValue := generateDNSAnnotationForSvc(p.Spec.ExternalAccess.DomainName, ssPodName)
		if annotationValue != "" {
			annotationMap = cloneMap(p.Spec.Pravega.SegmentStoreServiceAnnotations)
			annotationMap[externalDNSAnnotationKey] = annotationValue
		}
		service = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        ssPodName,
				Namespace:   p.Namespace,
				Labels:      p.LabelsForSegmentStore(),
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
					appsv1.StatefulSetPodNameLabel: fmt.Sprintf("%s-%d", p.StatefulSetNameForSegmentstore(), i),
				},
			},
		}
		if strings.EqualFold(p.Spec.Pravega.SegmentStoreExternalTrafficPolicy, "Cluster") == true {
			service.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeCluster
		} else {
			service.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
		}
		if p.Spec.Pravega.SegmentStoreLoadBalancerIP != "" {
			service.Spec.Ports[0].Port = 12345 + i
			service.Spec.LoadBalancerIP = p.Spec.Pravega.SegmentStoreLoadBalancerIP
		}
		services[i] = service
	}
	return services
}

func MakeSegmentstorePodDisruptionBudget(p *api.PravegaCluster) *policyv1beta1.PodDisruptionBudget {
	var maxUnavailable intstr.IntOrString

	if p.Spec.Pravega.SegmentStoreReplicas == int32(1) {
		maxUnavailable = intstr.FromInt(0)
	} else {
		maxUnavailable = intstr.FromInt(int(p.Spec.Pravega.MaxUnavailableSegmentStoreReplicas))
	}

	return &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.PdbNameForSegmentstore(),
			Namespace: p.Namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: p.LabelsForSegmentStore(),
			},
		},
	}
}
