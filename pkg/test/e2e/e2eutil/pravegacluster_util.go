/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package e2eutil

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	util "github.com/pravega/pravega-operator/pkg/util"
)

var (
	RetryInterval        = time.Second * 5
	Timeout              = time.Second * 60
	CleanupRetryInterval = time.Second * 1
	CleanupTimeout       = time.Second * 5
)

// CreateCluster creates a PravegaCluster CR with the desired spec
func CreateCluster(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, p *api.PravegaCluster) (*api.PravegaCluster, error) {
	t.Logf("creating pravega cluster: %s", p.Name)
	err := f.Client.Create(goctx.TODO(), p, &framework.CleanupOptions{TestContext: ctx, Timeout: CleanupTimeout, RetryInterval: CleanupRetryInterval})
	if err != nil {
		return nil, fmt.Errorf("failed to create CR: %v", err)
	}

	pravega := &api.PravegaCluster{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: p.Name}, pravega)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}
	t.Logf("created pravega cluster: %s", pravega.Name)
	return pravega, nil
}

// DeleteCluster deletes the PravegaCluster CR specified by cluster spec
func DeleteCluster(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, p *api.PravegaCluster) error {
	t.Logf("deleting pravega cluster: %s", p.Name)
	err := f.Client.Delete(goctx.TODO(), p)
	if err != nil {
		return fmt.Errorf("failed to delete CR: %v", err)
	}

	t.Logf("deleted pravega cluster: %s", p.Name)
	return nil
}

// UpdateCluster updates the PravegaCluster CR
func UpdateCluster(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, p *api.PravegaCluster) error {
	t.Logf("updating pravega cluster: %s", p.Name)
	err := f.Client.Update(goctx.TODO(), p)
	if err != nil {
		return fmt.Errorf("failed to update CR: %v", err)
	}

	t.Logf("updated pravega cluster: %s", p.Name)
	return nil
}

func isPodReady(pod *v1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// WaitForClusterToStart will wait until all cluster pods are ready
func WaitForClusterToStart(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, p *api.PravegaCluster, size int) error {
	t.Logf("waiting for pravega cluster to become ready: %s", p.Name)
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(util.LabelsForPravegaCluster(p)).String(),
	}

	err := wait.Poll(RetryInterval, 5*time.Minute, func() (done bool, err error) {
		podList, err := f.KubeClient.Core().Pods(p.Namespace).List(listOptions)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range podList.Items {
			pod := &podList.Items[i]

			if !isPodReady(pod) {
				continue
			}
			names = append(names, pod.Name)
		}
		t.Logf("waiting for pods to become ready (%d/%d), pods (%v)", len(names), size, names)
		if len(names) != int(size) {
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	t.Logf("pravega cluster ready: %s", p.Name)
	return nil
}

// WaitForClusterToTerminate will wait until all cluster pods are terminated
func WaitForClusterToTerminate(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, p *api.PravegaCluster) error {
	t.Logf("waiting for pravega cluster to terminate: %s", p.Name)
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(util.LabelsForPravegaCluster(p)).String(),
	}

	err := wait.Poll(RetryInterval, 2*time.Minute, func() (done bool, err error) {
		podList, err := f.KubeClient.Core().Pods(p.Namespace).List(listOptions)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range podList.Items {
			pod := &podList.Items[i]
			names = append(names, pod.Name)
		}
		t.Logf("waiting for pods to terminate, running pods (%v)", names)
		if len(names) != 0 {
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	t.Logf("pravega cluster terminated: %s", p.Name)
	return nil
}

// WriteAndReadData writes sample data and reads it back from the given Pravega cluster
func WriteAndReadData(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, p *api.PravegaCluster) error {
	t.Logf("writing and reading data from pravega cluster: %s", p.Name)
	testJob := NewTestWriteReadJob(p.Namespace, util.ServiceNameForController(p.Name))
	err := f.Client.Create(goctx.TODO(), testJob, &framework.CleanupOptions{TestContext: ctx, Timeout: CleanupTimeout, RetryInterval: CleanupRetryInterval})
	if err != nil {
		return fmt.Errorf("failed to create job: %s", err)
	}

	err = wait.Poll(RetryInterval, 3*time.Minute, func() (done bool, err error) {
		job, err := f.KubeClient.BatchV1().Jobs(p.Namespace).Get(testJob.Name, metav1.GetOptions{IncludeUninitialized: false})
		if err != nil {
			return false, err
		}
		if job.Status.CompletionTime.IsZero() {
			return false, nil
		}
		if job.Status.Failed > 0 {
			return false, fmt.Errorf("failed to write and read data from cluster")
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	t.Logf("pravega cluster validated: %s", p.Name)
	return nil
}

func CheckPvcSanity(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, p *api.PravegaCluster) error {
	t.Logf("checking pvc sanity: %s", p.Name)
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(util.LabelsForBookie(p)).String(),
	}
	pvcList, err := f.KubeClient.CoreV1().PersistentVolumeClaims(p.Namespace).List(listOptions)
	if err != nil {
		return err
	}

	for _, pvc := range pvcList.Items {
		if pvc.Status.Phase != v1.ClaimBound {
			continue
		}
		if util.PvcIsOrphan(pvc.Name, p.Spec.Bookkeeper.Replicas) {
			return fmt.Errorf("bookie pvc is illegal")
		}

	}

	listOptions = metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(util.LabelsForSegmentStore(p)).String(),
	}
	pvcList, err = f.KubeClient.CoreV1().PersistentVolumeClaims(p.Namespace).List(listOptions)
	if err != nil {
		return err
	}

	for _, pvc := range pvcList.Items {
		if pvc.Status.Phase != v1.ClaimBound {
			continue
		}
		if util.PvcIsOrphan(pvc.Name, p.Spec.Pravega.SegmentStoreReplicas) {
			return fmt.Errorf("segment store pvc is illegal")
		}

	}

	t.Logf("pvc validated: %s", p.Name)
	return nil
}
