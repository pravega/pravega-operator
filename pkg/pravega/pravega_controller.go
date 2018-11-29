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
	"k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deployController(pravegaCluster *api.PravegaCluster) (err error) {
	err = sdk.Create(makeControllerConfigMap(pravegaCluster))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = sdk.Create(makeControllerDeployment(pravegaCluster))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = sdk.Create(makeControllerService(pravegaCluster))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func makeControllerDeployment(pravegaCluster *api.PravegaCluster) *v1beta1.Deployment {
	return &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.DeploymentNameForController(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*k8sutil.AsOwnerRef(pravegaCluster),
			},
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &pravegaCluster.Spec.Pravega.ControllerReplicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: k8sutil.LabelsForController(pravegaCluster),
				},
				Spec: *makeControllerPodSpec(pravegaCluster.Name, &pravegaCluster.Spec.Pravega),
			},
		},
	}
}

func makeControllerPodSpec(name string, pravegaSpec *api.PravegaSpec) *corev1.PodSpec {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "pravega-controller",
				Image:           pravegaSpec.Image.String(),
				ImagePullPolicy: pravegaSpec.Image.PullPolicy,
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
								Name: k8sutil.ConfigMapNameForController(name),
							},
						},
					},
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: k8sutil.HealthcheckCommand(9090),
						},
					},
					// Controller pods start fast. We give it up to 1 minute to become ready.
					PeriodSeconds:    5,
					FailureThreshold: 12,
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: k8sutil.HealthcheckCommand(9090),
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
	}

	if pravegaSpec.ControllerServiceAccountName != "" {
		podSpec.ServiceAccountName = pravegaSpec.ControllerServiceAccountName
	}

	return podSpec
}

func makeControllerConfigMap(pravegaCluster *api.PravegaCluster) *corev1.ConfigMap {
	var javaOpts = []string{
		"-Dpravegaservice.clusterName=" + pravegaCluster.Name,
	}

	for name, value := range pravegaCluster.Spec.Pravega.Options {
		javaOpts = append(javaOpts, fmt.Sprintf("-D%v=%v", name, value))
	}

	configData := map[string]string{
		"CLUSTER_NAME":           pravegaCluster.Name,
		"ZK_URL":                 pravegaCluster.Spec.ZookeeperUri,
		"JAVA_OPTS":              strings.Join(javaOpts, " "),
		"REST_SERVER_PORT":       "10080",
		"CONTROLLER_SERVER_PORT": "9090",
		"AUTHORIZATION_ENABLED":  "false",
		"TOKEN_SIGNING_KEY":      "secret",
		"USER_PASSWORD_FILE":     "/etc/pravega/conf/passwd",
		"TLS_ENABLED":            "false",
		"WAIT_FOR":               pravegaCluster.Spec.ZookeeperUri,
	}

	for name, value := range pravegaCluster.Spec.Pravega.Options {
		configData[name] = value
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.ConfigMapNameForController(pravegaCluster.Name),
			Labels:    k8sutil.LabelsForController(pravegaCluster),
			Namespace: pravegaCluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*k8sutil.AsOwnerRef(pravegaCluster),
			},
		},
		Data: configData,
	}
}

func makeControllerService(pravegaCluster *api.PravegaCluster) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.ServiceNameForController(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
			Labels:    k8sutil.LabelsForController(pravegaCluster),
			OwnerReferences: []metav1.OwnerReference{
				*k8sutil.AsOwnerRef(pravegaCluster),
			},
		},
		Spec: corev1.ServiceSpec{
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
			Selector: k8sutil.LabelsForController(pravegaCluster),
		},
	}
}
