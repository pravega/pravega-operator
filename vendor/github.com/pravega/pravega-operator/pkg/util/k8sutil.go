/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package util

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8s "github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DownwardAPIEnv() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
	}
}

func PodAntiAffinity(component string, clusterName string) *corev1.Affinity {
	return &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "component",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{component},
								},
								{
									Key:      "pravega_cluster",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{clusterName},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}
}

// Wait for pods in cluster to be terminated
func WaitForClusterToTerminate(kubeClient client.Client, p *v1alpha1.PravegaCluster) (err error) {
	listOptions := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(LabelsForPravegaCluster(p)),
	}

	err = wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		podList := &corev1.PodList{}
		err = kubeClient.List(context.TODO(), listOptions, podList)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range podList.Items {
			pod := &podList.Items[i]
			names = append(names, pod.Name)
		}

		if len(names) != 0 {
			return false, nil
		}
		return true, nil
	})

	return err
}

func IsPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsPodFaulty(pod *corev1.Pod) (bool, error) {
	if pod.Status.ContainerStatuses[0].State.Waiting != nil && (pod.Status.ContainerStatuses[0].State.Waiting.Reason == "ImagePullBackOff" ||
		pod.Status.ContainerStatuses[0].State.Waiting.Reason == "CrashLoopBackOff") {
		return true, fmt.Errorf("pod %s update failed because of %s", pod.Name, pod.Status.ContainerStatuses[0].State.Waiting.Reason)
	}
	return false, nil
}

func NewEvent(name string, p *v1alpha1.PravegaCluster, reason string, message string, eventType string) *corev1.Event {
	now := metav1.Now()
	operatorName, _ := k8s.GetOperatorName()
	event := corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: p.Namespace,
			Labels:    LabelsForPravegaCluster(p),
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion:      "pravega.pravega.io/v1alpha1",
			Kind:            "PravegaCluster",
			Name:            p.GetName(),
			Namespace:       p.GetNamespace(),
			ResourceVersion: p.GetResourceVersion(),
			UID:             p.GetUID(),
		},
		Reason:              reason,
		Message:             message,
		FirstTimestamp:      now,
		LastTimestamp:       now,
		Type:                eventType,
		ReportingController: operatorName,
		ReportingInstance:   os.Getenv("POD_NAME"),
	}
	return &event
}
