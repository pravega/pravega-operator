/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package pravegacontroller

import (
	"context"
	"fmt"
	"sort"
	"time"

	pravegav1beta2 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta2"
	"github.com/pravega/pravega-operator/pkg/controller/pravega"
	"github.com/pravega/pravega-operator/pkg/util"
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
	fun  func(p *pravegav1beta2.PravegaController) (synced bool, err error)
}

// upgrade
func (r *ReconcilePravegaController) syncClusterVersion(p *pravegav1beta2.PravegaController) (err error) {
	defer func() {
		r.client.Status().Update(context.TODO(), p)
	}()

	// we cannot upgrade if cluster is in UpgradeFailed or Rollback state
	if p.Status.IsClusterInUpgradeFailedOrRollbackState() {
		return nil
	}

	_, upgradeCondition := p.Status.GetClusterCondition(pravegav1beta2.ControllerConditionUpgrading)
	_, readyCondition := p.Status.GetClusterCondition(pravegav1beta2.ControllerConditionPodsReady)

	if upgradeCondition == nil {
		// Initially set upgrading condition to false and
		// the current version to the version in the spec
		p.Status.SetUpgradingConditionFalse()
		p.Status.CurrentVersion = p.Spec.Version
		return nil
	}

	if upgradeCondition.Status == corev1.ConditionTrue {
		// Upgrade process already in progress
		if p.Status.TargetVersion == "" {
			log.Println("syncing to an unknown version: cancelling upgrade process")
			return r.clearUpgradeStatus(p)
		}

		if p.Status.TargetVersion == p.Status.CurrentVersion {
			log.Printf("syncing to version '%s' completed", p.Status.TargetVersion)
			return r.clearUpgradeStatus(p)
		}

		syncCompleted, err := r.syncControllerVersion(p)
		if err != nil {
			log.Printf("error syncing cluster version, upgrade failed. %v", err)
			p.Status.SetErrorConditionTrue("UpgradeFailed", err.Error())
			// emit an event for Upgrade Failure
			message := fmt.Sprintf("Error Upgrading from version %v to %v. %v", p.Status.CurrentVersion, p.Status.TargetVersion, err.Error())
			event := p.NewEvent("UPGRADE_ERROR", pravegav1beta2.UpgradeErrorReason, message, "Error")
			pubErr := r.client.Create(context.TODO(), event)
			if pubErr != nil {
				log.Printf("Error publishing Upgrade Failure event to k8s. %v", pubErr)
			}
			r.clearUpgradeStatus(p)
			return err
		}

		if syncCompleted {
			// All component versions have been synced
			p.Status.AddToVersionHistory(p.Status.TargetVersion)
			p.Status.CurrentVersion = p.Status.TargetVersion
			log.Printf("Upgrade completed for all pravega components.")
		}
		return nil
	}

	// No upgrade in progress
	if p.Spec.Version == p.Status.CurrentVersion {
		// No intention to upgrade
		return nil
	}

	if !p.Status.IsClusterInRollbackFailedState() {
		// skip this check when cluster is in RollbackFailed state
		if readyCondition == nil || readyCondition.Status != corev1.ConditionTrue {
			r.clearUpgradeStatus(p)
			log.Print("cannot trigger upgrade if there are unready pods")
			return nil
		}
	} else {
		// We are upgrading after a rollback failure, reset Error Status
		p.Status.SetErrorConditionFalse()
	}

	// Need to sync cluster versions
	log.Printf("syncing cluster version from %s to %s", p.Status.CurrentVersion, p.Spec.Version)
	// Setting target version and condition.
	// The upgrade process will start on the next reconciliation
	p.Status.TargetVersion = p.Spec.Version
	p.Status.SetUpgradingConditionTrue("", "")

	return nil
}

func (r *ReconcilePravegaController) clearUpgradeStatus(p *pravegav1beta2.PravegaController) (err error) {
	p.Status.SetUpgradingConditionFalse()
	p.Status.TargetVersion = ""
	// need to deep copy the status struct, otherwise it will be overwritten
	// when updating the CR below
	status := p.Status.DeepCopy()

	if err := r.client.Update(context.TODO(), p); err != nil {
		return err
	}

	p.Status = *status
	return nil
}

func (r *ReconcilePravegaController) rollbackClusterVersion(p *pravegav1beta2.PravegaController, version string) (err error) {
	defer func() {
		r.client.Status().Update(context.TODO(), p)
	}()
	_, rollbackCondition := p.Status.GetClusterCondition(pravegav1beta2.ControllerConditionRollback)
	if rollbackCondition == nil || rollbackCondition.Status != corev1.ConditionTrue {
		// We're in the first iteration for Rollback
		// Add Rollback Condition to Cluster Status
		log.Printf("Updating Target Version to  %v", version)
		p.Status.TargetVersion = version
		p.Status.SetRollbackConditionTrue("", "")
		updateErr := r.client.Status().Update(context.TODO(), p)
		if updateErr != nil {
			p.Status.SetRollbackConditionFalse()
			log.Printf("Error updating cluster: %v", updateErr.Error())
			return fmt.Errorf("Error updating cluster status. %v", updateErr)
		}
		return nil
	}

	syncCompleted, err := r.syncControllerVersion(p)
	if err != nil {
		// Error rolling back, set appropriate status and ask for manual intervention
		p.Status.SetErrorConditionTrue("RollbackFailed", err.Error())
		// emit an event for Rollback Failure
		message := fmt.Sprintf("Error Rollingback from version %v to %v. %v", p.Status.CurrentVersion, p.Status.TargetVersion, err.Error())
		event := p.NewEvent("ROLLBACK_ERROR", pravegav1beta2.RollbackErrorReason, message, "Error")
		pubErr := r.client.Create(context.TODO(), event)
		if pubErr != nil {
			log.Printf("Error publishing ROLLBACK_ERROR event to k8s. %v", pubErr)
		}
		r.clearRollbackStatus(p)
		log.Printf("Error rolling back to cluster version %v. Reason: %v", version, err)
		//r.client.Status().Update(context.TODO(), p)
		return err
	}

	if syncCompleted {
		// All component versions have been synced
		p.Status.CurrentVersion = p.Status.TargetVersion
		// Set Error/UpgradeFailed Condition to 'false', so rollback is not triggered again
		p.Status.SetErrorConditionFalse()
		r.clearRollbackStatus(p)
		log.Printf("Rollback to version %v completed for all pravega components.", version)
	}
	//r.client.Status().Update(context.TODO(), p)
	return nil
}

func (r *ReconcilePravegaController) clearRollbackStatus(p *pravegav1beta2.PravegaController) (err error) {
	log.Printf("clearRollbackStatus")
	p.Status.SetRollbackConditionFalse()
	p.Status.TargetVersion = ""
	// need to deep copy the status struct, otherwise it will be overwritten
	// when updating the CR below
	status := p.Status.DeepCopy()

	if err := r.client.Update(context.TODO(), p); err != nil {
		return err
	}

	p.Status = *status
	return nil
}

func (r *ReconcilePravegaController) syncControllerVersion(p *pravegav1beta2.PravegaController) (synced bool, err error) {
	deploy := &appsv1.Deployment{}
	name := p.DeploymentNameForController()
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)
	if err != nil {
		return false, fmt.Errorf("failed to get deployment (%s): %v", deploy.Name, err)
	}

	targetImage, err := p.PravegaControllerTargetImage()
	if err != nil {
		return false, err
	}

	if deploy.Spec.Template.Spec.Containers[0].Image != targetImage {
		p.Status.UpdateProgress(pravegav1beta2.UpdatingControllerReason, "0")

		// Need to update pod template
		// This will trigger the rolling upgrade process
		log.Printf("updating deployment (%s) pod template image to '%s'", deploy.Name, targetImage)

		configMap := pravega.MakeControllerConfigMap(p)
		controllerutil.SetControllerReference(p, configMap, r.scheme)
		err = r.client.Update(context.TODO(), configMap)
		if err != nil {
			return false, err
		}

		deploy.Spec.Template = pravega.MakeControllerPodTemplate(p)
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
		// Update still in progress, check if there is progress made within the timeout.
		for _, v := range deploy.Status.Conditions {
			if v.Type == appsv1.DeploymentProgressing &&
				v.Status == corev1.ConditionFalse && v.Reason == "ProgressDeadlineExceeded" {
				// upgrade fails
				return false, fmt.Errorf("updating deployment (%s) failed due to %s", deploy.Name, v.Reason)
			}
		}
		// Check if the updated pod has error. If so, return error and fail fast
		pods, err := r.getDeployPodsWithVersion(deploy, p.Status.TargetVersion)
		if err != nil {
			return false, err
		}
		_, err = r.checkUpdatedPods(pods, p.Status.TargetVersion)
		if err != nil {
			// Abort if there is any errors with the updated pods
			return false, err
		}
		// Wait until next reconcile iteration
		return false, nil
	}

	// Deployment update completed
	return true, nil
}

func (r *ReconcilePravegaController) checkUpdatedPods(pods []*corev1.Pod, version string) (bool, error) {
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

func (r *ReconcilePravegaController) getOneOutdatedPod(sts *appsv1.StatefulSet, version string) (*corev1.Pod, error) {
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
	err = r.client.List(context.TODO(), podList, podlistOps)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(podList.Items, func(i int, j int) bool {
		return podList.Items[i].Name < podList.Items[j].Name
	})

	for _, podItem := range podList.Items {
		if util.GetPodVersion(&podItem) == version {
			continue
		}
		return &podItem, nil
	}
	return nil, nil
}

func (r *ReconcilePravegaController) getStsPodsWithVersion(sts *appsv1.StatefulSet, version string) ([]*corev1.Pod, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: sts.Spec.Template.Labels,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert label selector: %v", err)
	}

	return r.getPodsWithVersion(selector, sts.Namespace, version)
}

func (r *ReconcilePravegaController) getDeployPodsWithVersion(deploy *appsv1.Deployment, version string) ([]*corev1.Pod, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: deploy.Spec.Template.Labels,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert label selector: %v", err)
	}

	return r.getPodsWithVersion(selector, deploy.Namespace, version)
}

func (r *ReconcilePravegaController) getPodsWithVersion(selector labels.Selector, namespace string, version string) ([]*corev1.Pod, error) {
	podList := &corev1.PodList{}
	podlistOps := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: selector,
	}
	err := r.client.List(context.TODO(), podList, podlistOps)
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

func checkSyncTimeout(p *pravegav1beta2.PravegaController, reason string, updatedReplicas int32) error {
	lastCondition := p.Status.GetLastCondition()
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
	p.Status.UpdateProgress(reason, fmt.Sprint(updatedReplicas))
	return nil
}
