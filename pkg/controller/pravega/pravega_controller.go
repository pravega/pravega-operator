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
	"github.com/pravega/pravega-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func MakeControllerDeployment(p *api.PravegaCluster) *appsv1.Deployment {
	zero := int32(0)
	timeout := int32(600)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.DeploymentNameForController(),
			Namespace: p.Namespace,
			Labels:    p.LabelsForController(),
		},
		Spec: appsv1.DeploymentSpec{
			ProgressDeadlineSeconds: &timeout,
			Replicas:                &p.Spec.Pravega.ControllerReplicas,
			RevisionHistoryLimit:    &zero,
			Template:                MakeControllerPodTemplate(p),
			Selector: &metav1.LabelSelector{
				MatchLabels: p.LabelsForController(),
			},
		},
	}
}

func MakeControllerPodTemplate(p *api.PravegaCluster) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      p.LabelsForController(),
			Annotations: map[string]string{"pravega.version": p.Spec.Version},
		},
		Spec: *makeControllerPodSpec(p),
	}
}

func makeControllerPodSpec(p *api.PravegaCluster) *corev1.PodSpec {
	// Parse volumes & volumeMounts parameters first
	var hostPathVolumeMounts []string
	var emptyDirVolumeMounts []string
	var configMapVolumeMounts []string
	var ok bool

	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount

	if _, ok = p.Spec.Pravega.Options["hostPathVolumeMounts"]; ok {
		hostPathVolumeMounts = strings.Split(p.Spec.Pravega.Options["hostPathVolumeMounts"], ",")
		for _, vm := range hostPathVolumeMounts {
			s := strings.Split(vm, "=")
			v := corev1.Volume{
				Name: s[0],
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: s[1],
					},
				},
			}
			volumes = append(volumes, v)

			m := corev1.VolumeMount{
				Name:      s[0],
				MountPath: s[1],
			}
			volumeMounts = append(volumeMounts, m)
		}
	}
	if _, ok = p.Spec.Pravega.Options["emptyDirVolumeMounts"]; ok {
		emptyDirVolumeMounts = strings.Split(p.Spec.Pravega.Options["emptyDirVolumeMounts"], ",")
		for _, vm := range emptyDirVolumeMounts {
			s := strings.Split(vm, "=")
			v := corev1.Volume{
				Name: s[0],
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			}
			volumes = append(volumes, v)

			m := corev1.VolumeMount{
				Name:      s[0],
				MountPath: s[1],
			}
			volumeMounts = append(volumeMounts, m)
		}
	} else {
		// if user did not set emptyDirVolumeMounts
		v := corev1.Volume{
			Name: heapDumpName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
		volumes = append(volumes, v)

		m := corev1.VolumeMount{
			Name:      heapDumpName,
			MountPath: heapDumpDir,
		}
		volumeMounts = append(volumeMounts, m)
	}
	if _, ok = p.Spec.Pravega.Options["configMapVolumeMounts"]; ok {
		configMapVolumeMounts = strings.Split(p.Spec.Pravega.Options["configMapVolumeMounts"], ",")
		for _, vm := range configMapVolumeMounts {
			p := strings.Split(vm, "=")
			s := strings.Split(p[0], ":")
			v := corev1.Volume{
				Name: s[0],
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: s[0],
						},
					},
				},
			}
			volumes = append(volumes, v)

			m := corev1.VolumeMount{
				Name:      s[0],
				MountPath: p[1],
				SubPath:   s[1],
			}
			volumeMounts = append(volumeMounts, m)
		}
	}

	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "pravega-controller",
				Image:           p.PravegaImage(),
				ImagePullPolicy: p.Spec.Pravega.Image.PullPolicy,
				Args: []string{
					"controller",
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "rest",
						ContainerPort: 10080,
					},
					{
						Name:          "grpc",
						ContainerPort: 9090,
					},
				},
				EnvFrom: []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: p.ConfigMapNameForController(),
							},
						},
					},
				},
				VolumeMounts: volumeMounts,
				Resources:    *p.Spec.Pravega.ControllerResources,
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: util.ControllerReadinessCheck(10080, p.Spec.Authentication.IsEnabled()),
						},
					},
					// Controller pods start fast. We give it up to 20 seconds to become ready.
					InitialDelaySeconds: 20,
					TimeoutSeconds:      60,
					SuccessThreshold:    3,
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: util.HealthcheckCommand(9090),
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
		Affinity: p.Spec.Pravega.ControllerPodAffinity,
		Volumes:  volumes,
	}
	if p.Spec.Pravega.ControllerServiceAccountName != "" {
		podSpec.ServiceAccountName = p.Spec.Pravega.ControllerServiceAccountName
	}

	if p.Spec.Pravega.ControllerSecurityContext != nil {
		podSpec.SecurityContext = p.Spec.Pravega.ControllerSecurityContext
	}

	configureControllerTLSSecrets(podSpec, p)
	configureAuthSecrets(podSpec, p)
	configureControllerAuthSecrets(podSpec, p)
	return podSpec
}

func configureControllerTLSSecrets(podSpec *corev1.PodSpec, p *api.PravegaCluster) {
	if p.Spec.TLS.IsSecureController() {
		addSecretVolumeWithMount(podSpec, p, tlsVolumeName, p.Spec.TLS.Static.ControllerSecret, tlsVolumeName, tlsMountDir)
	}
}

func configureAuthSecrets(podSpec *corev1.PodSpec, p *api.PravegaCluster) {
	if p.Spec.Authentication.IsEnabled() && p.Spec.Authentication.PasswordAuthSecret != "" {
		addSecretVolumeWithMount(podSpec, p, authVolumeName, p.Spec.Authentication.PasswordAuthSecret,
			authVolumeName, authMountDir)
	}
}

func configureControllerAuthSecrets(podSpec *corev1.PodSpec, p *api.PravegaCluster) {
	if p.Spec.Authentication.IsEnabled() && p.Spec.Authentication.ControllerTokenSecret != "" {
		addSecretVolumeWithMount(podSpec, p, controllerAuthVolumeName, p.Spec.Authentication.ControllerTokenSecret,
			controllerAuthVolumeName, controllerAuthMountDir)
	}
}

func addSecretVolumeWithMount(podSpec *corev1.PodSpec, p *api.PravegaCluster,
	volumeName string, secretName string,
	mountName string, mountDir string) {
	vol := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
	podSpec.Volumes = append(podSpec.Volumes, vol)

	podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      mountName,
		MountPath: mountDir,
	})
}

func MakeControllerConfigMap(p *api.PravegaCluster) *corev1.ConfigMap {
	javaOpts := []string{
		"-Dpravegaservice.clusterName=" + p.Name,
	}

	jvmOpts := []string{
		"-Xms512m",
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

	javaOpts = append(javaOpts, util.OverrideDefaultJVMOptions(jvmOpts, p.Spec.Pravega.ControllerJvmOptions)...)

	for name, value := range p.Spec.Pravega.Options {
		javaOpts = append(javaOpts, fmt.Sprintf("-D%v=%v", name, value))
	}

	sort.Strings(javaOpts)

	authEnabledStr := fmt.Sprint(p.Spec.Authentication.IsEnabled())
	configData := map[string]string{
		"CLUSTER_NAME":           p.Name,
		"ZK_URL":                 p.Spec.ZookeeperUri,
		"JAVA_OPTS":              strings.Join(javaOpts, " "),
		"REST_SERVER_PORT":       "10080",
		"CONTROLLER_SERVER_PORT": "9090",
		"AUTHORIZATION_ENABLED":  authEnabledStr,
		"TOKEN_SIGNING_KEY":      defaultTokenSigningKey,
		"TLS_ENABLED":            "false",
		"WAIT_FOR":               p.Spec.ZookeeperUri,
	}

	if p.Spec.Pravega.DebugLogging {
		configData["log.level"] = "DEBUG"
	}
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.ConfigMapNameForController(),
			Labels:    p.LabelsForController(),
			Namespace: p.Namespace,
		},
		Data: configData,
	}
	return configMap
}

func getControllerServiceType(pravegaCluster *api.PravegaCluster) (serviceType corev1.ServiceType) {
	if pravegaCluster.Spec.Pravega.ControllerExternalServiceType == "" {
		if pravegaCluster.Spec.ExternalAccess.Type == "" {
			return api.DefaultServiceType
		}
		return pravegaCluster.Spec.ExternalAccess.Type
	}
	return pravegaCluster.Spec.Pravega.ControllerExternalServiceType
}

func MakeControllerService(p *api.PravegaCluster) *corev1.Service {
	serviceType := corev1.ServiceTypeClusterIP
	annotationMap := map[string]string{}
	if p.Spec.ExternalAccess.Enabled {
		serviceType = getControllerServiceType(p)
		for k, v := range p.Spec.Pravega.ControllerServiceAnnotations {
			annotationMap[k] = v
		}
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        p.ServiceNameForController(),
			Namespace:   p.Namespace,
			Labels:      p.LabelsForController(),
			Annotations: annotationMap,
		},
		Spec: corev1.ServiceSpec{
			Type: serviceType,
			Ports: []corev1.ServicePort{
				{
					Name: "rest",
					Port: 10080,
				},
				{
					Name: "grpc",
					Port: 9090,
				},
			},
			Selector: p.LabelsForController(),
		},
	}
}

func MakeControllerPodDisruptionBudget(p *api.PravegaCluster) *policyv1beta1.PodDisruptionBudget {
	minAvailable := intstr.FromInt(int(p.Spec.Pravega.MaxUnavailableControllerReplicas))
	return &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.PdbNameForController(),
			Namespace: p.Namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: p.LabelsForController(),
			},
		},
	}
}
