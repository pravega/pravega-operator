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
	"time"

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
		// Initially set upgrading condition to false
		p.Status.SetUpgradingConditionFalse()
		return nil
	}

	if p.Status.CurrentVersion == "" {
		// Initially set the current version to the version in the spec
		p.Status.CurrentVersion = p.Spec.Version
	}

	if condition.Status == corev1.ConditionTrue {
		// upgrade process already in progress

		if p.Status.TargetVersion == "" {
			log.Println("upgrading to an unknown version. resetting upgrade condition to false")
			p.Status.SetUpgradingConditionFalse()
			return nil
		}

		if p.Status.TargetVersion == p.Status.CurrentVersion {
			log.Printf("upgrade to version '%s' completed", p.Status.TargetVersion)
			return r.clearUpgradeStatus(p)
		}

		if err := r.syncComponentsVersion(p); err != nil {
			log.Printf("error upgrading cluster, aborting upgrade process")
			// TODO: set condition reason and message
			return r.clearUpgradeStatus(p)
		}
		return nil
		//return r.dummyUpgrade(p)
	}

	// no upgrade in progress

	if p.Spec.Version == p.Status.CurrentVersion {
		// No need to sync versions
		return nil
	}

	// The user wants to upgrade to a different version
	log.Printf("user wants to upgrade from %s to %s", p.Status.CurrentVersion, p.Spec.Version)

	// setting target version and condition.
	// will start the upgrade process in the next reconciliation
	p.Status.TargetVersion = p.Spec.Version
	p.Status.SetUpgradingConditionTrue()

	return nil
}

func (r *ReconcilePravegaCluster) clearUpgradeStatus(p *pravegav1alpha1.PravegaCluster) (err error) {
	p.Status.SetUpgradingConditionFalse()
	p.Status.TargetVersion = ""
	p.Spec.Version = p.Status.CurrentVersion

	if err := r.client.Update(context.TODO(), p.DeepCopy()); err != nil {
		return err
	}
	return nil
}

func (r *ReconcilePravegaCluster) dummyUpgrade(p *pravegav1alpha1.PravegaCluster) (err error) {
	log.Println("*** UPGRADING *** ")
	_, condition := p.Status.GetClusterCondition(pravegav1alpha1.ClusterConditionUpgrading)

	now := time.Now()
	t, err := time.Parse(time.RFC3339, condition.LastTransitionTime)
	if err != nil {
		return err
	}

	if now.Sub(t) > time.Minute*2 {
		log.Println("*** UPGRADE FINISHED ***")
		p.Status.CurrentVersion = p.Status.TargetVersion
	}

	return nil
}

func (r *ReconcilePravegaCluster) syncComponentsVersion(p *pravegav1alpha1.PravegaCluster) (err error) {
	var synced bool

	for _, component := range []componentSyncVersionFun{
		componentSyncVersionFun{
			name: "controller",
			fun:  r.syncControllerVersion,
		},
		componentSyncVersionFun{
			name: "segmentstore",
			fun:  r.syncSegmentStoreVersion,
		},
		componentSyncVersionFun{
			name: "bookkeeper",
			fun:  r.syncBookkeeperVersion,
		},
	} {
		synced, err = component.fun(p)
		if err != nil {
			return fmt.Errorf("failed to sync %s version: %v", component.name, err)
		}

		if synced {
			fmt.Printf("%s version already synced", component.name)
		} else {
			fmt.Printf("%s version syncing in progress", component.name)
			// Do not continue with the next component until this one is synced
			return nil
		}
	}

	return nil
}

func (r *ReconcilePravegaCluster) syncControllerVersion(p *pravegav1alpha1.PravegaCluster) (synced bool, err error) {
	deploy := &appsv1.Deployment{}
	name := util.DeploymentNameForController(p.Name)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)
	if err != nil {
		return false, fmt.Errorf("failed to get deployment (%s): %v", deploy.Name, err)
	}

	// check if upgrade already in progress (upgrade condition)
	//   yes: check updatedReplicas vs readyReplicas ?
	//   no:  create patch with image tag change
	//        apply patch

	targetImage := fmt.Sprintf("%s:%s", p.Spec.Pravega.ImageRepository, p.Status.TargetVersion)

	if deploy.Spec.Template.Spec.Containers[0].Image == targetImage {
		// Pod template already updated

	} else {
		// Need to update pod template
		deploy.Spec.Template.Spec.Containers[0].Image = targetImage
		err = r.client.Update(context.TODO(), deploy)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func (r *ReconcilePravegaCluster) syncSegmentStoreVersion(p *pravegav1alpha1.PravegaCluster) (synced bool, err error) {
	// TODO: implement
	return true, nil
}

func (r *ReconcilePravegaCluster) syncBookkeeperVersion(p *pravegav1alpha1.PravegaCluster) (synced bool, err error) {
	// TODO: implement
	return true, nil
}
