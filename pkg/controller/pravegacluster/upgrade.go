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
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type componentSyncVersionFun struct {
	name string
	fun  func(p *pravegav1alpha1.PravegaCluster) (synced bool, err error)
}

func (r *ReconcilePravegaCluster) syncClusterVersion(p *pravegav1alpha1.PravegaCluster) (err error) {
	defer func() {
		r.client.Status().Update(context.TODO(), p)
	}()

	_, condition := p.Status.GetClusterCondition(pravegav1alpha1.ClusterConditionUpgrading)

	if condition == nil {
		// Initially set upgrading condition to false and
		// the current version to the version in the spec
		p.Status.SetUpgradingConditionFalse()
		p.Status.CurrentVersion = p.Spec.Version
		return nil
	}

	if condition.Status == corev1.ConditionTrue {
		// Upgrade process already in progress

		if p.Status.TargetVersion == "" {
			log.Println("upgrading to an unknown version: resetting upgrade condition to false")
			p.Status.SetUpgradingConditionFalse()
			return nil
		}

		if p.Status.TargetVersion == p.Status.CurrentVersion {
			log.Printf("upgrade to version '%s' completed", p.Status.TargetVersion)
			return r.clearUpgradeStatus(p)
		}

		if err := r.syncComponentsVersion(p); err != nil {
			log.Printf("error upgrading cluster: aborting upgrade process: %v", err)
			// TODO: set error condition with reason and message
			// TODO: roll back upgrade
			return r.clearUpgradeStatus(p)
		}
		return nil
		//return r.dummyUpgrade(p)
	}

	// No upgrade in progress

	if p.Spec.Version == p.Status.CurrentVersion {
		// No intention to upgrade
		return nil
	}

	// The user wants to upgrade to a different version
	log.Printf("user wants to upgrade from %s to %s", p.Status.CurrentVersion, p.Spec.Version)

	// Setting target version and condition.
	// The upgrade process will start on the next reconciliation
	p.Status.TargetVersion = p.Spec.Version
	p.Status.SetUpgradingConditionTrue()

	return nil
}

func (r *ReconcilePravegaCluster) clearUpgradeStatus(p *pravegav1alpha1.PravegaCluster) (err error) {
	p.Status.SetUpgradingConditionFalse()
	p.Status.TargetVersion = ""
	// need to deep copy the status struct, otherwise it will be overridden
	// when updating the CR below
	status := p.Status.DeepCopy()

	p.Spec.Version = p.Status.CurrentVersion
	if err := r.client.Update(context.TODO(), p); err != nil {
		return err
	}

	p.Status = *status
	return nil
}

func (r *ReconcilePravegaCluster) syncComponentsVersion(p *pravegav1alpha1.PravegaCluster) (err error) {
	var synced bool

	for _, component := range []componentSyncVersionFun{
		componentSyncVersionFun{
			name: "bookkeeper",
			fun:  r.syncBookkeeperVersion,
		},
		componentSyncVersionFun{
			name: "segmentstore",
			fun:  r.syncSegmentStoreVersion,
		},
		componentSyncVersionFun{
			name: "controller",
			fun:  r.syncControllerVersion,
		},
	} {
		synced, err = component.fun(p)
		if err != nil {
			return fmt.Errorf("failed to upgrade %s to %v", component.name, err)
		}

		if synced {
			log.Printf("%s upgrade has been completed", component.name)
		} else {
			// component upgrade is still in progress
			// Do not continue with the next component until this one is done
			return nil
		}
	}

	// All component versions have been synced
	p.Status.CurrentVersion = p.Status.TargetVersion
	return nil
}

func (r *ReconcilePravegaCluster) syncControllerVersion(p *pravegav1alpha1.PravegaCluster) (synced bool, err error) {
	deploy := &appsv1.Deployment{}
	name := util.DeploymentNameForController(p.Name)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)
	if err != nil {
		return false, fmt.Errorf("failed to get deployment (%s): %v", deploy.Name, err)
	}

	targetImage := fmt.Sprintf("%s:%s", p.Spec.Pravega.ImageRepository, p.Status.TargetVersion)

	if deploy.Spec.Template.Spec.Containers[0].Image != targetImage {
		// Need to update pod template
		// This will trigger the rolling upgrade process
		log.Printf("updating deployment (%s) pod template image to '%s'", deploy.Name, targetImage)
		deploy.Spec.Template.Spec.Containers[0].Image = targetImage
		err = r.client.Update(context.TODO(), deploy)
		if err != nil {
			return false, err
		}
		// Updated pod template. Upgrade process has been triggered
		return false, nil
	}

	// Pod template already updated
	log.Printf("deployment (%s) status: %d updated, %d ready, %d target", deploy.Name,
		deploy.Status.UpdatedReplicas, deploy.Status.ReadyReplicas, deploy.Status.Replicas)

	// Check whether the upgrade is in progress or has completed
	if deploy.Status.UpdatedReplicas != deploy.Status.Replicas ||
		deploy.Status.UpdatedReplicas != deploy.Status.ReadyReplicas {
		// Update still in progress
		return false, nil
	}

	// Deployment update completed
	return true, nil
}

func (r *ReconcilePravegaCluster) syncSegmentStoreVersion(p *pravegav1alpha1.PravegaCluster) (synced bool, err error) {

	sts := &appsv1.StatefulSet{}
	name := util.StatefulSetNameForSegmentstore(p.Name)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
	if err != nil {
		return false, fmt.Errorf("failed to get statefulset (%s): %v", sts.Name, err)
	}

	targetImage := fmt.Sprintf("%s:%s", p.Spec.Pravega.ImageRepository, p.Status.TargetVersion)

	if sts.Spec.Template.Spec.Containers[0].Image != targetImage {
		// Need to update pod template
		// This will trigger the rolling upgrade process
		log.Printf("updating statefulset (%s) template image to '%s'", sts.Name, targetImage)
		sts.Spec.Template.Spec.Containers[0].Image = targetImage
		err = r.client.Update(context.TODO(), sts)
		if err != nil {
			return false, err
		}
		// Updated pod template. Upgrade process has been triggered
		return false, nil
	}

	// Pod template already updated

	log.Printf("statefulset (%s) status: %d updated, %d ready, %d target", sts.Name,
		sts.Status.UpdatedReplicas, sts.Status.ReadyReplicas, sts.Status.Replicas)

	// Check whether the upgrade is in progress or has completed
	if sts.Status.UpdatedReplicas != sts.Status.Replicas ||
		sts.Status.UpdatedReplicas != sts.Status.ReadyReplicas {
		// Upgrade still in progress
		return false, nil
	}

	// StatefulSet upgrade completed
	return true, nil
}

func (r *ReconcilePravegaCluster) syncBookkeeperVersion(p *pravegav1alpha1.PravegaCluster) (synced bool, err error) {
	sts := &appsv1.StatefulSet{}
	name := util.StatefulSetNameForBookie(p.Name)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
	if err != nil {
		return false, fmt.Errorf("failed to get statefulset (%s): %v", sts.Name, err)
	}

	targetImage := fmt.Sprintf("%s:%s", p.Spec.Bookkeeper.ImageRepository, p.Status.TargetVersion)

	if sts.Spec.Template.Spec.Containers[0].Image != targetImage {
		// Need to update pod template
		// This will trigger the rolling upgrade process
		log.Printf("updating statefulset (%s) template image to '%s'", sts.Name, targetImage)
		sts.Spec.Template.Spec.Containers[0].Image = targetImage
		err = r.client.Update(context.TODO(), sts)
		if err != nil {
			return false, err
		}
		// Updated pod template. Upgrade process has been triggered
		return false, nil
	}

	// Pod template already updated

	log.Printf("statefulset (%s) status: %d updated, %d ready, %d target", sts.Name,
		sts.Status.UpdatedReplicas, sts.Status.ReadyReplicas, sts.Status.Replicas)

	// Check whether the upgrade is in progress or has completed
	if sts.Status.UpdatedReplicas != sts.Status.Replicas ||
		sts.Status.UpdatedReplicas != sts.Status.ReadyReplicas {
		// Upgrade still in progress
		return false, nil
	}

	// StatefulSet upgrade completed
	return true, nil
}
