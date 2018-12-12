/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package pravegacluster

import (
	"context"
	"fmt"
	pravegav1alpha1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/util"
	"k8s.io/apimachinery/pkg/labels"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes/scheme"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ZKMETAFINALIZER = "zk.meta.finalizer.pravega.io"
	ZKUSERFINALIZER = "zk.user.finalizer.pravega.io"
)

func (r *ReconcilePravegaCluster) cleanUpMetaInZk(p *pravegav1alpha1.PravegaCluster) (err error) {
	// Find a zookeeper node by using the selector in zookeeper service
	zkService := &corev1.Service{}
	buf := strings.Split(p.Spec.ZookeeperUri, ":")
	zkServiceName := buf[0]
	err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: zkServiceName}, zkService)
	if err != nil {
		return fmt.Errorf("failed to get zookeeper service (%s): %v", zkServiceName, err)
	}

	zkNodeLabels := util.LabelsForZookeeperNode(zkService.Spec.Selector)
	zkNodeSelector := labels.SelectorFromSet(zkNodeLabels)
	zkNodeListOps := &client.ListOptions{Namespace: p.Namespace, LabelSelector: zkNodeSelector}

	zkNodeList := &corev1.PodList{}
	err = r.client.List(context.TODO(), zkNodeListOps, zkNodeList)
	if err != nil {
		return fmt.Errorf("failed to list zookeeper pods: %v", err)
	}

	if len(zkNodeList.Items) == 0 {
		return fmt.Errorf("failed to find zookeeper node")
	}

	// Choose one of the zookeeper nodes
	zkNode := zkNodeList.Items[0]
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to get REST config from kube config: %v", err)
	}

	coreclient, err := corev1client.NewForConfig(restconfig)
	if err != nil {
		return fmt.Errorf("failed to create client from REST config: %v", err)
	}

	req := coreclient.RESTClient().
		Post().
		Namespace(zkNode.Namespace).
		Resource("pods").
		Name(zkNode.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: zkNode.Spec.Containers[0].Name,
			Command:   []string{"./bin/zkCli.sh", "deleteall", "/pravega"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restconfig, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %v", err)
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    true,
	})
	if err != nil {
		return fmt.Errorf("zookeeper is still in use, wait and retry: %v", err)
	}
	return nil
}

func (r *ReconcilePravegaCluster) cleanUpZkUser(p *pravegav1alpha1.PravegaCluster) (err error) {
	bookie := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.StatefulSetNameForBookie(p.Name),
			Namespace: p.Namespace,
		},
	}
	err = r.client.Delete(context.TODO(), bookie)
	if err != nil {
		return fmt.Errorf("failed to delete bookie: %v", err)
	}

	segmentstore := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.StatefulSetNameForSegmentstore(p.Name),
			Namespace: p.Namespace,
		},
	}
	err = r.client.Delete(context.TODO(), segmentstore)
	if err != nil {
		return fmt.Errorf("failed to delete segmentstore: %v", err)
	}

	controller := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.DeploymentNameForController(p.Name),
			Namespace: p.Namespace,
		},
	}
	err = r.client.Delete(context.TODO(), controller)
	if err != nil {
		return fmt.Errorf("failed to delete controller: %v", err)
	}
	return nil
}

func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func removeString(slice []string, str string) (result []string) {
	for _, item := range slice {
		if item == str {
			continue
		}
		result = append(result, item)
	}
	return result
}
