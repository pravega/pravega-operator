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
	"time"

	pravegav1alpha1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

func (r *ReconcilePravegaCluster) syncClusterVersion(p *pravegav1alpha1.PravegaCluster) (err error) {

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
			p.Status.SetUpgradingConditionFalse()
			p.Status.TargetVersion = ""
			p.Spec.Version = p.Status.CurrentVersion

			if err := r.client.Update(context.TODO(), p.DeepCopy()); err != nil {
				return err
			}
			return nil
		}

		return upgrade(p)
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

func upgrade(p *pravegav1alpha1.PravegaCluster) (err error) {
	// TODO: do the actual upgrade
	log.Println("*** UPGRADING *** ")
	_, condition := p.Status.GetClusterCondition(pravegav1alpha1.ClusterConditionUpgrading)

	now := time.Now()
	t, err := time.Parse(time.RFC3339, condition.LastTransitionTime)
	if err != nil {
		return err
	}

	if now.Sub(t) > time.Minute*1 {
		log.Println("*** UPGRADE FINISHED ***")
		p.Status.CurrentVersion = p.Status.TargetVersion
	}

	return nil
}
