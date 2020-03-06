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

func MakeControllerDeployment(p *api.PravegaCluster) *appsv1.Deployment {
	zero := int32(0)
	timeout := int32(600)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.DeploymentNameForController(p.Name),
			Namespace: p.Namespace,
			Labels:    util.LabelsForController(p),
		},
		Spec: appsv1.DeploymentSpec{
			ProgressDeadlineSeconds: &timeout,
			Replicas:                &p.Spec.Pravega.ControllerReplicas,
			RevisionHistoryLimit:    &zero,
			Template:                MakeControllerPodTemplate(p),
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForController(p),
			},
		},
	}
}

func MakeControllerPodTemplate(p *api.PravegaCluster) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      util.LabelsForController(p),
			Annotations: map[string]string{"pravega.version": p.Spec.Version},
		},
		Spec: *makeControllerPodSpec(p),
	}
}

func makeControllerPodSpec(p *api.PravegaCluster) *corev1.PodSpec {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "pravega-controller",
				Image:           util.PravegaImage(p),
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
								Name: util.ConfigMapNameForController(p.Name),
							},
						},
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      heapDumpName,
						MountPath: heapDumpDir,
					},
				},
				Resources: *p.Spec.Pravega.ControllerResources,
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: util.HealthcheckCommand(9090),
						},
					},
					// Controller pods start fast. We give it up to 1 minute to become ready.
					PeriodSeconds:    5,
					FailureThreshold: 12,
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
		Affinity: util.PodAntiAffinity("pravega-controller", p.Name),
		Volumes: []corev1.Volume{
			{
				Name: heapDumpName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
	}

	if p.Spec.Pravega.ControllerServiceAccountName != "" {
		podSpec.ServiceAccountName = p.Spec.Pravega.ControllerServiceAccountName
	}

	configureControllerTLSSecrets(podSpec, p)
	configureAuthSecrets(podSpec, p)
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
			"-XX:+UseCGroupMemoryLimitForHeap",
			"-XX:MaxRAMFraction=2",
		)
	}

	javaOpts = append(javaOpts, util.OverrideDefaultJVMOptions(jvmOpts, p.Spec.Pravega.ControllerJvmOptions)...)

	for name, value := range p.Spec.Pravega.Options {
		javaOpts = append(javaOpts, fmt.Sprintf("-D%v=%v", name, value))
	}

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

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.ConfigMapNameForController(p.Name),
			Labels:    util.LabelsForController(p),
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
			Name:        util.ServiceNameForController(p.Name),
			Namespace:   p.Namespace,
			Labels:      util.LabelsForController(p),
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
			Selector: util.LabelsForController(p),
		},
	}
}

func MakeControllerPodDisruptionBudget(pravegaCluster *api.PravegaCluster) *policyv1beta1.PodDisruptionBudget {
	minAvailable := intstr.FromInt(1)
	return &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.PdbNameForController(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForController(pravegaCluster),
			},
		},
	}
}
