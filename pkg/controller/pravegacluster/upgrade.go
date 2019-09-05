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
	"github.com/pravega/pravega-operator/pkg/controller/pravega"
	"github.com/pravega/pravega-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type componentSyncVersionFun struct {
	name string
	fun  func(p *pravegav1alpha1.PravegaCluster) (synced bool, err error)
}

func (r *ReconcilePravegaCluster) syncClusterVersion(p *pravegav1alpha1.PravegaCluster) (err error) {
	defer func() {
		r.client.Status().Update(context.TODO(), p)
	}()

	_, upgradeCondition := p.Status.GetClusterCondition(pravegav1alpha1.ClusterConditionUpgrading)
	_, readyCondition := p.Status.GetClusterCondition(pravegav1alpha1.ClusterConditionPodsReady)

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

		if err := r.syncComponentsVersion(p); err != nil {
			log.Printf("error syncing cluster version, need manual intervention. %v", err)
			// TODO: Trigger roll back to previous version
			p.Status.SetErrorConditionTrue("UpgradeFailed", err.Error())
			r.clearUpgradeStatus(p)
		}
		return nil
	}

	// No upgrade in progress

	if p.Spec.Version == p.Status.CurrentVersion {
		// No intention to upgrade
		return nil
	}

	if readyCondition == nil || readyCondition.Status != corev1.ConditionTrue {
		r.clearUpgradeStatus(p)
		log.Print("cannot trigger upgrade if there are unready pods")
		return nil
	}

	// Need to sync cluster versions
	log.Printf("syncing cluster version from %s to %s", p.Status.CurrentVersion, p.Spec.Version)

	// Setting target version and condition.
	// The upgrade process will start on the next reconciliation
	p.Status.TargetVersion = p.Spec.Version
	p.Status.SetUpgradingConditionTrue("", "")

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
			return fmt.Errorf("failed to sync %s version. %s", component.name, err)
		}

		if synced {
			log.Printf("%s version sync has been completed", component.name)
		} else {
			// component version sync is still in progress
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

	targetImage, err := util.PravegaTargetImage(p)
	if err != nil {
		return false, err
	}

	if deploy.Spec.Template.Spec.Containers[0].Image != targetImage {
		// Need to update pod template
		// This will trigger the rolling upgrade process
		log.Printf("updating deployment (%s) pod template image to '%s'", deploy.Name, targetImage)

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
		// Update still in progress, check if there is progress made within a timeout.
		for _, v := range deploy.Status.Conditions {
			if v.Type == appsv1.DeploymentProgressing &&
				v.Status == corev1.ConditionFalse && v.Reason == "ProgressDeadlineExceeded" {
				// upgrade fails
				return false, fmt.Errorf("updating deployment (%s) failed due to %s", deploy.Name, v.Reason)
			}
		}
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

	targetImage, err := util.PravegaTargetImage(p)
	if err != nil {
		return false, err
	}

	if sts.Spec.Template.Spec.Containers[0].Image != targetImage {
		// Need to update pod template
		// This will trigger the rolling upgrade process
		log.Printf("updating statefulset (%s) template image to '%s'", sts.Name, targetImage)

		sts.Spec.Template = pravega.MakeSegmentStorePodTemplate(p)
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
	if sts.Status.UpdatedReplicas == sts.Status.Replicas &&
		sts.Status.UpdatedReplicas == sts.Status.ReadyReplicas {
		// StatefulSet upgrade completed
		return true, nil
	}
	// Upgrade still in progress

	// Check if segmentstore fail to have progress within a timeout
	err = checkUpgradeCondition(p, pravegav1alpha1.UpgradingSegmentstoreReason, sts.Status.UpdatedReplicas)
	if err != nil {
		return false, fmt.Errorf("updating statefulset (%s) failed due to %v", sts.Name, err)
	}

	// If all replicas are ready, upgrade an old pod
	ready, err := r.checkUpdatedPods(sts, p.Status.TargetVersion)
	if err != nil {
		// Abort if there is any errors with the updated pods
		return false, err
	}

	if ready {
		pod, err := r.getOneOutdatedPod(sts, p.Status.TargetVersion)
		if err != nil {
			return false, err
		}

		if pod == nil {
			return false, fmt.Errorf("could not obtain outdated pod")
		}

		log.Infof("upgrading pod: %s", pod.Name)

		err = r.client.Delete(context.TODO(), pod)
		if err != nil {
			return false, err
		}
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

	targetImage, err := util.BookkeeperTargetImage(p)
	if err != nil {
		return false, err
	}

	if sts.Spec.Template.Spec.Containers[0].Image != targetImage {
		// Need to update pod template
		// This will trigger the rolling upgrade process
		log.Printf("updating statefulset (%s) template image to '%s'", sts.Name, targetImage)
		sts.Spec.Template = pravega.MakeBookiePodTemplate(p)
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
		// TODO: wait until there is no under replicated ledger
		// https://bookkeeper.apache.org/docs/4.7.2/reference/cli/#listunderreplicated
		return true, nil
	}

	// Check if bookkeeper fail to have progress
	err = checkUpgradeCondition(p, pravegav1alpha1.UpgradingBookkeeperReason, sts.Status.UpdatedReplicas)
	if err != nil {
		return false, fmt.Errorf("updating statefulset (%s) failed due to %v", sts.Name, err)
	}

	// Upgrade still in progress
	// If all replicas are ready, upgrade an old pod

	ready, err := r.checkUpdatedPods(sts, p.Status.TargetVersion)
	if err != nil {
		// Abort if there is any errors with the updated pods
		return false, err
	}

	if ready {
		pod, err := r.getOneOutdatedPod(sts, p.Status.TargetVersion)
		if err != nil {
			return false, err
		}

		if pod == nil {
			return false, fmt.Errorf("could not obtain outdated pod")
		}

		log.Infof("upgrading pod: %s", pod.Name)

		err = r.client.Delete(context.TODO(), pod)
		if err != nil {
			return false, err
		}
	}

	// wait until the next reconcile iteration
	return false, nil
}

func (r *ReconcilePravegaCluster) checkUpdatedPods(sts *appsv1.StatefulSet, version string) (bool, error) {
	pods, err := r.getPodsWithVersion(sts, version)
	if err != nil {
		return false, err
	}

	for _, pod := range pods {
		//TODO: find out a more reliable way to determine if a pod is having issues
		if pod.Status.ContainerStatuses[0].RestartCount > 1 {
			return false, fmt.Errorf("pod %s is restarting", pod.Name)
		}

		if !util.IsPodReady(pod) {
			// At least one updated pod is still not ready
			if pod.Status.ContainerStatuses[0].State.Waiting != nil && (pod.Status.ContainerStatuses[0].State.Waiting.Reason == "ImagePullBackOff" ||
				pod.Status.ContainerStatuses[0].State.Waiting.Reason == "CrashLoopBackOff") {
				return false, fmt.Errorf("pod %s update failed because of %s", pod.Name, pod.Status.ContainerStatuses[0].State.Waiting.Reason)
			}
			return false, nil
		}
	}
	return true, nil
}

func (r *ReconcilePravegaCluster) getOneOutdatedPod(sts *appsv1.StatefulSet, version string) (*corev1.Pod, error) {
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

func (r *ReconcilePravegaCluster) getPodsWithVersion(sts *appsv1.StatefulSet, version string) ([]*corev1.Pod, error) {
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

	var pods []*corev1.Pod
	for _, podItem := range podList.Items {
		if util.GetPodVersion(&podItem) != version {
			continue
		}
		pods = append(pods, podItem.DeepCopy())
	}
	return pods, nil
}

func checkUpgradeCondition(p *pravegav1alpha1.PravegaCluster, reason string, updatedReplicas int32) error {
	_, lastCondition := p.Status.GetClusterCondition(pravegav1alpha1.ClusterConditionUpgrading)
	if lastCondition.Reason == reason && lastCondition.Message != string(updatedReplicas) {
		// nothing changed, check if reaches timeout
		parsedTime, _ := time.Parse(time.RFC3339, lastCondition.LastUpdateTime)
		if time.Now().After(parsedTime.Add(time.Duration(10 * time.Minute))) {
			// timeout has passed
			return fmt.Errorf("progress deadline exceeded")
		}
	}
	// update the status to the latest, this will also set the transition timestamp to now
	p.Status.SetUpgradingConditionTrue(reason, string(updatedReplicas))
	return nil
}
