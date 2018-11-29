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

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/utils/k8sutil"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ReconcilePravegaCluster(pravegaCluster *api.PravegaCluster) (err error) {
	err = deployBookie(pravegaCluster)
	if err != nil {
		return err
	}

	err = deployController(pravegaCluster)
	if err != nil {
		return err
	}

	err = deploySegmentStore(pravegaCluster)
	if err != nil {
		return err
	}

	err = syncClusterSize(pravegaCluster)
	if err != nil {
		return err
	}

	return nil
}

func syncClusterSize(pravegaCluster *api.PravegaCluster) (err error) {
	err = syncBookieSize(pravegaCluster)
	if err != nil {
		return err
	}

	err = syncSegmentStoreSize(pravegaCluster)
	if err != nil {
		return err
	}

	err = syncControllerSize(pravegaCluster)
	if err != nil {
		return err
	}

	return nil
}

func syncBookieSize(pravegaCluster *api.PravegaCluster) (err error) {
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.StatefulSetNameForBookie(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
		},
	}

	err = sdk.Get(sts)
	if err != nil {
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}

	if *sts.Spec.Replicas != pravegaCluster.Spec.Bookkeeper.Replicas {
		sts.Spec.Replicas = &(pravegaCluster.Spec.Bookkeeper.Replicas)
		err = sdk.Update(sts)
		if err != nil {
			return fmt.Errorf("failed to update size of stateful-set (%s): %v", sts.Name, err)
		}
	}

	return nil
}

func syncSegmentStoreSize(pravegaCluster *api.PravegaCluster) (err error) {
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.StatefulSetNameForSegmentstore(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
		},
	}

	err = sdk.Get(sts)
	if err != nil {
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}

	if *sts.Spec.Replicas != pravegaCluster.Spec.Pravega.SegmentStoreReplicas {
		sts.Spec.Replicas = &(pravegaCluster.Spec.Pravega.SegmentStoreReplicas)
		err = sdk.Update(sts)
		if err != nil {
			return fmt.Errorf("failed to update size of stateful-set (%s): %v", sts.Name, err)
		}
	}
	return nil
}

func syncControllerSize(pravegaCluster *api.PravegaCluster) (err error) {
	deploy := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.DeploymentNameForController(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
		},
	}

	err = sdk.Get(deploy)
	if err != nil {
		return fmt.Errorf("failed to get deployment (%s): %v", deploy.Name, err)
	}

	if *deploy.Spec.Replicas != pravegaCluster.Spec.Pravega.ControllerReplicas {
		deploy.Spec.Replicas = &(pravegaCluster.Spec.Pravega.ControllerReplicas)
		err = sdk.Update(deploy)
		if err != nil {
			return fmt.Errorf("failed to update size of deployment (%s): %v", deploy.Name, err)
		}
	}
	return nil
}
