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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeControllerDeployment(p *api.PravegaCluster) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.DeploymentNameForController(p.Name),
			Namespace: p.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &p.Spec.Pravega.ControllerReplicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: util.LabelsForController(p),
				},
				Spec: *makeControllerPodSpec(p.Name, p.Spec.Pravega),
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: util.LabelsForController(p),
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
								Name: util.ConfigMapNameForController(name),
							},
						},
					},
				},
			},
		},
	}

	if pravegaSpec.ControllerServiceAccountName != "" {
		podSpec.ServiceAccountName = pravegaSpec.ControllerServiceAccountName
	}

	return podSpec
}

func MakeControllerConfigMap(p *api.PravegaCluster) *corev1.ConfigMap {
	var javaOpts = []string{
		"-Dpravegaservice.clusterName=" + p.Name,
	}

	for name, value := range p.Spec.Pravega.Options {
		javaOpts = append(javaOpts, fmt.Sprintf("-D%v=%v", name, value))
	}

	configData := map[string]string{
		"CLUSTER_NAME":           p.Name,
		"ZK_URL":                 p.Spec.ZookeeperUri,
		"JAVA_OPTS":              strings.Join(javaOpts, " "),
		"REST_SERVER_PORT":       "10080",
		"CONTROLLER_SERVER_PORT": "9090",
		"AUTHORIZATION_ENABLED":  "false",
		"TOKEN_SIGNING_KEY":      "secret",
		"USER_PASSWORD_FILE":     "/etc/pravega/conf/passwd",
		"TLS_ENABLED":            "false",
		"WAIT_FOR":               p.Spec.ZookeeperUri,
	}

	for name, value := range p.Spec.Pravega.Options {
		configData[name] = value
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

func MakeControllerService(p *api.PravegaCluster) *corev1.Service {
	serviceType := corev1.ServiceTypeClusterIP
	if p.Spec.ExternalAccess.Enabled {
		serviceType = p.Spec.ExternalAccess.Type
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.ServiceNameForController(p.Name),
			Namespace: p.Namespace,
			Labels:    util.LabelsForController(p),
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
