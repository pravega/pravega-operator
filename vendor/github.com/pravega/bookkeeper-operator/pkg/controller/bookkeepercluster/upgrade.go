/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package bookkeepercluster

import (
	"context"
	"fmt"
	"time"

	bookkeeperv1alpha1 "github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type componentSyncVersionFun struct {
	name string
	fun  func(p *bookkeeperv1alpha1.BookkeeperCluster) (synced bool, err error)
}

// upgrade
func (r *ReconcileBookkeeperCluster) syncClusterVersion(bk *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	defer func() {
		r.client.Status().Update(context.TODO(), bk)
	}()

	// we cannot upgrade if cluster is in UpgradeFailed or Rollback state
	if bk.Status.IsClusterInUpgradeFailedOrRollbackState() {
		return nil
	}

	_, upgradeCondition := bk.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionUpgrading)
	_, readyCondition := bk.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionPodsReady)

	if upgradeCondition == nil {
		// Initially set upgrading condition to false and
		// the current version to the version in the spec
		bk.Status.SetUpgradingConditionFalse()
		bk.Status.CurrentVersion = bk.Spec.Version
		return nil
	}

	if upgradeCondition.Status == corev1.ConditionTrue {
		// Upgrade process already in progress
		if bk.Status.TargetVersion == "" {
			log.Println("syncing to an unknown version: cancelling upgrade process")
			return r.clearUpgradeStatus(bk)
		}

		if bk.Status.TargetVersion == bk.Status.CurrentVersion {
			log.Printf("syncing to version '%s' completed", bk.Status.TargetVersion)
			return r.clearUpgradeStatus(bk)
		}

		//syncCompleted, err := r.syncComponentsVersion(p)
		syncCompleted, err := r.syncBookkeeperVersion(bk)
		if err != nil {
			log.Printf("error syncing cluster version, upgrade failed. %v", err)
			bk.Status.SetErrorConditionTrue("UpgradeFailed", err.Error())
			// emit an event for Upgrade Failure
			message := fmt.Sprintf("Error Upgrading from version %v to %v. %v", bk.Status.CurrentVersion, bk.Status.TargetVersion, err.Error())
			event := util.NewEvent("UPGRADE_ERROR", bk, bookkeeperv1alpha1.UpgradeErrorReason, message, "Error")
			pubErr := r.client.Create(context.TODO(), event)
			if pubErr != nil {
				log.Printf("Error publishing Upgrade Failure event to k8s. %v", pubErr)
			}
			r.clearUpgradeStatus(bk)
			return err
		}

		if syncCompleted {
			// All component versions have been synced
			bk.Status.AddToVersionHistory(bk.Status.TargetVersion)
			bk.Status.CurrentVersion = bk.Status.TargetVersion
			log.Printf("Upgrade completed for all bookkeeper components.")
		}
		return nil
	}

	// No upgrade in progress
	if bk.Spec.Version == bk.Status.CurrentVersion {
		// No intention to upgrade
		return nil
	}

	if !bk.Status.IsClusterInRollbackFailedState() {
		// skip this check when cluster is in RollbackFailed state
		if readyCondition == nil || readyCondition.Status != corev1.ConditionTrue {
			r.clearUpgradeStatus(bk)
			log.Print("cannot trigger upgrade if there are unready pods")
			return nil
		}
	} else {
		// We are upgrading after a rollback failure, reset Error Status
		bk.Status.SetErrorConditionFalse()
	}

	// Need to sync cluster versions
	log.Printf("syncing cluster version from %s to %s", bk.Status.CurrentVersion, bk.Spec.Version)
	// Setting target version and condition.
	// The upgrade process will start on the next reconciliation
	bk.Status.TargetVersion = bk.Spec.Version
	bk.Status.SetUpgradingConditionTrue("", "")

	return nil
}

func (r *ReconcileBookkeeperCluster) clearUpgradeStatus(bk *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	bk.Status.SetUpgradingConditionFalse()
	bk.Status.TargetVersion = ""
	// need to deep copy the status struct, otherwise it will be overwritten
	// when updating the CR below
	status := bk.Status.DeepCopy()

	if err := r.client.Update(context.TODO(), bk); err != nil {
		return err
	}

	bk.Status = *status
	return nil
}

func (r *ReconcileBookkeeperCluster) rollbackClusterVersion(bk *bookkeeperv1alpha1.BookkeeperCluster, version string) (err error) {
	defer func() {
		r.client.Status().Update(context.TODO(), bk)
	}()
	_, rollbackCondition := bk.Status.GetClusterCondition(bookkeeperv1alpha1.ClusterConditionRollback)
	if rollbackCondition == nil || rollbackCondition.Status != corev1.ConditionTrue {
		// We're in the first iteration for Rollback
		// Add Rollback Condition to Cluster Status
		log.Printf("Updating Target Version to  %v", version)
		bk.Status.TargetVersion = version
		bk.Status.SetRollbackConditionTrue("", "")
		updateErr := r.client.Status().Update(context.TODO(), bk)
		if updateErr != nil {
			bk.Status.SetRollbackConditionFalse()
			log.Printf("Error updating cluster: %v", updateErr.Error())
			return fmt.Errorf("Error updating cluster status. %v", updateErr)
		}
		return nil
	}

	syncCompleted, err := r.syncBookkeeperVersion(bk)
	if err != nil {
		// Error rolling back, set appropriate status and ask for manual intervention
		bk.Status.SetErrorConditionTrue("RollbackFailed", err.Error())
		// emit an event for Rollback Failure
		message := fmt.Sprintf("Error Rollingback from version %v to %v. %v", bk.Status.CurrentVersion, bk.Status.TargetVersion, err.Error())
		event := util.NewEvent("ROLLBACK_ERROR", bk, bookkeeperv1alpha1.RollbackErrorReason, message, "Error")
		pubErr := r.client.Create(context.TODO(), event)
		if pubErr != nil {
			log.Printf("Error publishing ROLLBACK_ERROR event to k8s. %v", pubErr)
		}
		r.clearRollbackStatus(bk)
		log.Printf("Error rolling back to cluster version %v. Reason: %v", version, err)
		//r.client.Status().Update(context.TODO(), p)
		return err
	}

	if syncCompleted {
		// All component versions have been synced
		bk.Status.CurrentVersion = bk.Status.TargetVersion
		// Set Error/UpgradeFailed Condition to 'false', so rollback is not triggered again
		bk.Status.SetErrorConditionFalse()
		r.clearRollbackStatus(bk)
		log.Printf("Rollback to version %v completed for all bookkeeper components.", version)
	}
	//r.client.Status().Update(context.TODO(), p)
	return nil
}

func (r *ReconcileBookkeeperCluster) clearRollbackStatus(bk *bookkeeperv1alpha1.BookkeeperCluster) (err error) {
	log.Printf("clearRollbackStatus")
	bk.Status.SetRollbackConditionFalse()
	bk.Status.TargetVersion = ""
	// need to deep copy the status struct, otherwise it will be overwritten
	// when updating the CR below
	status := bk.Status.DeepCopy()

	if err := r.client.Update(context.TODO(), bk); err != nil {
		return err
	}

	bk.Status = *status
	return nil
}

func (r *ReconcileBookkeeperCluster) syncBookkeeperVersion(bk *bookkeeperv1alpha1.BookkeeperCluster) (synced bool, err error) {
	sts := &appsv1.StatefulSet{}
	name := util.StatefulSetNameForBookie(bk.Name)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: bk.Namespace}, sts)
	if err != nil {
		return false, fmt.Errorf("failed to get statefulset (%s): %v", sts.Name, err)
	}

	targetImage, err := util.BookkeeperTargetImage(bk)
	if err != nil {
		return false, err
	}

	if sts.Spec.Template.Spec.Containers[0].Image != targetImage {
		bk.Status.UpdateProgress(bookkeeperv1alpha1.UpdatingBookkeeperReason, "0")
		// Need to update pod template
		// This will trigger the rolling upgrade process
		log.Printf("updating statefulset (%s) template image to '%s'", sts.Name, targetImage)

		configMap := MakeBookieConfigMap(bk)
		controllerutil.SetControllerReference(bk, configMap, r.scheme)
		err = r.client.Update(context.TODO(), configMap)
		if err != nil {
			return false, err
		}

		sts.Spec.Template = MakeBookiePodTemplate(bk)
		err = r.client.Update(context.TODO(), sts)
		if err != nil {
			return false, err
		}

		// Updated pod template
		return false, nil
	}

	// Pod template already updated
	log.Printf("statefulset (%s) status: %d updated, %d ready, %d target", sts.Name,
		sts.Status.UpdatedReplicas, sts.Status.ReadyReplicas, sts.Status.Replicas)

	// Check whether the upgrade is in progress or has completed
	if sts.Status.UpdatedReplicas == sts.Status.Replicas &&
		sts.Status.UpdatedReplicas == sts.Status.ReadyReplicas {
		// StatefulSet upgrade completed
		return true, nil
	}

	// Upgrade still in progress
	// Check if bookkeeper fail to have progress
	err = checkSyncTimeout(bk, bookkeeperv1alpha1.UpdatingBookkeeperReason, sts.Status.UpdatedReplicas)
	if err != nil {
		return false, fmt.Errorf("updating statefulset (%s) failed due to %v", sts.Name, err)
	}

	// If all replicas are ready, upgrade an old pod
	pods, err := r.getStsPodsWithVersion(sts, bk.Status.TargetVersion)
	if err != nil {
		return false, err
	}
	ready, err := r.checkUpdatedPods(pods, bk.Status.TargetVersion)
	if err != nil {
		// Abort if there is any errors with the updated pods
		return false, err
	}

	if ready {
		pod, err := r.getOneOutdatedPod(sts, bk.Status.TargetVersion)
		if err != nil {
			return false, err
		}

		if pod == nil {
			return false, fmt.Errorf("could not obtain outdated pod")
		}

		log.Infof("updating pod: %s", pod.Name)

		err = r.client.Delete(context.TODO(), pod)
		if err != nil {
			return false, err
		}
	}
	// wait until the next reconcile iteration
	return false, nil
}

func (r *ReconcileBookkeeperCluster) checkUpdatedPods(pods []*corev1.Pod, version string) (bool, error) {
	for _, pod := range pods {
		if !util.IsPodReady(pod) {
			// At least one updated pod is still not ready, check if it is faulty.
			if faulty, err := util.IsPodFaulty(pod); faulty {
				return false, err
			}
			return false, nil
		}
	}
	return true, nil
}

func (r *ReconcileBookkeeperCluster) getOneOutdatedPod(sts *appsv1.StatefulSet, version string) (*corev1.Pod, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: sts.Spec.Template.Labels,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert label selector: %v", err)
	}

	podList := &corev1.PodList{}
	podlistOps := &client.ListOptions{
		Namespace:     sts.Namespace,
		LabelSelector: selector,
	}
	err = r.client.List(context.TODO(), podlistOps, podList)
	if err != nil {
		return nil, err
	}

	for _, podItem := range podList.Items {
		if util.GetPodVersion(&podItem) == version {
			continue
		}
		return &podItem, nil
	}
	return nil, nil
}

func (r *ReconcileBookkeeperCluster) getStsPodsWithVersion(sts *appsv1.StatefulSet, version string) ([]*corev1.Pod, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: sts.Spec.Template.Labels,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert label selector: %v", err)
	}

	return r.getPodsWithVersion(selector, sts.Namespace, version)
}

func (r *ReconcileBookkeeperCluster) getPodsWithVersion(selector labels.Selector, namespace string, version string) ([]*corev1.Pod, error) {
	podList := &corev1.PodList{}
	podlistOps := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: selector,
	}
	err := r.client.List(context.TODO(), podlistOps, podList)
	if err != nil {
		return nil, err
	}

	var pods []*corev1.Pod
	for _, podItem := range podList.Items {
		if util.GetPodVersion(&podItem) != version {
			continue
		}
		pods = append(pods, podItem.DeepCopy())
	}
	return pods, nil
}

func checkSyncTimeout(bk *bookkeeperv1alpha1.BookkeeperCluster, reason string, updatedReplicas int32) error {
	lastCondition := bk.Status.GetLastCondition()
	if lastCondition == nil {
		return nil
	}
	if lastCondition.Reason == reason && lastCondition.Message == fmt.Sprint(updatedReplicas) {
		// if reason and message are the same as before, which means there is no progress since the last reconciling,
		// then check if it reaches the timeout.
		parsedTime, _ := time.Parse(time.RFC3339, lastCondition.LastUpdateTime)
		if time.Now().After(parsedTime.Add(time.Duration(10 * time.Minute))) {
			// timeout
			return fmt.Errorf("progress deadline exceeded")
		}
		// it hasn't reached timeout
		return nil
	}
	bk.Status.UpdateProgress(reason, fmt.Sprint(updatedReplicas))
	return nil
}
