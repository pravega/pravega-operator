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
	"strconv"
	"time"

	pravegav1beta1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	"github.com/pravega/pravega-operator/pkg/controller/pravega"
	"github.com/pravega/pravega-operator/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	log "github.com/sirupsen/logrus"
)

// ReconcileTime is the delay between reconciliations
const ReconcileTime = 30 * time.Second

// Add creates a new PravegaCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePravegaCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pravegacluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PravegaCluster
	err = c.Watch(&source.Kind{Type: &pravegav1beta1.PravegaCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePravegaCluster{}

// ReconcilePravegaCluster reconciles a PravegaCluster object
type ReconcilePravegaCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PravegaCluster object and makes changes based on the state read
// and what is in the PravegaCluster.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePravegaCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling PravegaCluster %s/%s\n", request.Namespace, request.Name)

	// Fetch the PravegaCluster instance
	pravegaCluster := &pravegav1beta1.PravegaCluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, pravegaCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Printf("PravegaCluster %s/%s not found. Ignoring since object must be deleted\n", request.Namespace, request.Name)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Printf("failed to get PravegaCluster: %v", err)
		return reconcile.Result{}, err
	}

	// Set default configuration for unspecified values
	changed := pravegaCluster.WithDefaults()
	if changed {
		log.Printf("Setting default settings for pravega-cluster: %s", request.Name)
		if err = r.client.Update(context.TODO(), pravegaCluster); err != nil {
			log.Printf("Error applying defaults on Pravega Cluster %v", err)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	err = r.run(pravegaCluster)
	if err != nil {
		log.Printf("failed to reconcile pravega cluster (%s): %v", pravegaCluster.Name, err)
		return reconcile.Result{}, err
	}
	log.Printf("Completed Reconcile of pravega cluster %s/%s", request.Namespace, request.Name)
	return reconcile.Result{RequeueAfter: ReconcileTime}, nil
}

func (r *ReconcilePravegaCluster) run(p *pravegav1beta1.PravegaCluster) (err error) {

	err = r.reconcileFinalizers(p)
	if err != nil {
		return fmt.Errorf("failed to reconcile finalizers %v", err)
	}

	err = r.deployCluster(p)
	if err != nil {
		return fmt.Errorf("failed to deploy cluster: %v", err)
	}

	err = r.syncClusterSize(p)
	if err != nil {
		return fmt.Errorf("failed to sync cluster size: %v", err)
	}

	// Upgrade
	err = r.syncClusterVersion(p)
	if err != nil {
		return fmt.Errorf("failed to sync cluster version: %v", err)
	}

	// Rollback
	err = r.rollbackFailedUpgrade(p)
	if err != nil {
		return fmt.Errorf("Rollback attempt failed: %v", err)
	}

	err = r.reconcileClusterStatus(p)
	if err != nil {
		return fmt.Errorf("failed to reconcile cluster status: %v", err)
	}
	return nil
}

func (r *ReconcilePravegaCluster) reconcileFinalizers(p *pravegav1beta1.PravegaCluster) (err error) {
	zkFinalizer := "cleanUpZookeeper"
	if util.ContainsString(p.ObjectMeta.Finalizers, zkFinalizer) {
		p.ObjectMeta.Finalizers = util.RemoveString(p.ObjectMeta.Finalizers, zkFinalizer)
		if err = r.client.Update(context.TODO(), p); err != nil {
			return fmt.Errorf("failed to remove Zk Finalizer from Pravega object (%s): %v", p.Name, err)
		}
		log.Printf("ZK Finalizer removed from Pravega CR.")
	}
	return nil
}

func (r *ReconcilePravegaCluster) deployCluster(p *pravegav1beta1.PravegaCluster) (err error) {
	err = r.deployController(p)
	if err != nil {
		log.Printf("failed to deploy controller: %v", err)
		return err
	}

	/*this check is to avoid creation of a new segmentstore when the CurrentVersion is below 07 and target version is above 07
	  as we are doing it in the upgrade path*/
	if !r.IsClusterUpgradingTo07(p) && !r.IsClusterRollbackingFrom07(p) {

		err = r.deploySegmentStore(p)
		if err != nil {
			log.Printf("failed to deploy segment store: %v", err)
			return err
		}

		if !util.IsVersionBelow07(p.Spec.Version) {
			newsts := &appsv1.StatefulSet{}
			name := p.StatefulSetNameForSegmentstoreAbove07()
			err = r.client.Get(context.TODO(),
				types.NamespacedName{Name: name, Namespace: p.Namespace}, newsts)
			if err != nil {
				return fmt.Errorf("failed to get stateful-set (%s): %v", newsts.Name, err)
			}
			if newsts.Status.ReadyReplicas > 0 {
				return r.deleteOldSegmentStoreIfExists(p)
			}
		}
	}
	return nil
}

func (r *ReconcilePravegaCluster) deleteSTS(p *pravegav1beta1.PravegaCluster) error {
	// We should be here only in case of Pravega CR version migration
	// to version 0.5.0 from version 0.4.x
	sts := &appsv1.StatefulSet{}
	stsName := p.StatefulSetNameForSegmentstoreBelow07()
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: stsName, Namespace: p.Namespace}, sts)
	if err != nil {
		if errors.IsNotFound(err) {
			// nothing to do since old STS was not found
			return nil
		}
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}
	// delete sts, if found
	r.client.Delete(context.TODO(), sts)
	log.Printf("Deleted old SegmentStore STS %s", sts.Name)
	return nil
}

func (r *ReconcilePravegaCluster) deletePVC(p *pravegav1beta1.PravegaCluster) error {
	numPvcs := int(p.Spec.Pravega.SegmentStoreReplicas)
	for i := 0; i < numPvcs; i++ {
		pvcName := "cache-" + p.StatefulSetNameForSegmentstoreBelow07() + "-" + strconv.Itoa(i)
		//log.Printf("PVC NAME %s", pvcName)
		pvc := &corev1.PersistentVolumeClaim{}
		err := r.client.Get(context.TODO(),
			types.NamespacedName{Name: pvcName, Namespace: p.Namespace}, pvc)
		if err != nil {
			if errors.IsNotFound(err) {
				// nothing to do since old STS was not found
				continue
			}
			return fmt.Errorf("failed to get pvc (%s): %v", pvcName, err)
		}
		pvcDelete := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvcName,
				Namespace: p.Namespace,
			},
		}
		err = r.client.Delete(context.TODO(), pvcDelete)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcilePravegaCluster) deleteOldSegmentStoreIfExists(p *pravegav1beta1.PravegaCluster) error {
	err := r.deleteSTS(p)
	if err != nil {
		return err
	}
	err = r.deletePVC(p)
	if err != nil {
		return err
	}
	if p.Spec.ExternalAccess.Enabled {
		// delete external Services
		for i := int32(0); i < p.Spec.Pravega.SegmentStoreReplicas; i++ {
			extService := &corev1.Service{}
			svcName := p.ServiceNameForSegmentStoreBelow07(i)
			err := r.client.Get(context.TODO(), types.NamespacedName{Name: svcName, Namespace: p.Namespace}, extService)
			if err != nil {
				if errors.IsNotFound(err) {
					// nothing to do since old STS was not found
					return nil
				}
				return fmt.Errorf("failed to get external service (%s): %v", svcName, err)
			}
			r.client.Delete(context.TODO(), extService)
			log.Printf("Deleted old SegmentStore external service %s", extService)
		}
	}
	return nil
}

func (r *ReconcilePravegaCluster) deployController(p *pravegav1beta1.PravegaCluster) (err error) {

	pdb := pravega.MakeControllerPodDisruptionBudget(p)
	controllerutil.SetControllerReference(p, pdb, r.scheme)
	err = r.client.Create(context.TODO(), pdb)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	configMap := pravega.MakeControllerConfigMap(p)
	controllerutil.SetControllerReference(p, configMap, r.scheme)
	err = r.client.Create(context.TODO(), configMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	deployment := pravega.MakeControllerDeployment(p)
	controllerutil.SetControllerReference(p, deployment, r.scheme)
	err = r.client.Create(context.TODO(), deployment)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	service := pravega.MakeControllerService(p)
	controllerutil.SetControllerReference(p, service, r.scheme)
	err = r.client.Create(context.TODO(), service)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (r *ReconcilePravegaCluster) deploySegmentStore(p *pravegav1beta1.PravegaCluster) (err error) {

	headlessService := pravega.MakeSegmentStoreHeadlessService(p)
	controllerutil.SetControllerReference(p, headlessService, r.scheme)
	err = r.client.Create(context.TODO(), headlessService)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	if p.Spec.ExternalAccess.Enabled {
		services := pravega.MakeSegmentStoreExternalServices(p)
		for _, service := range services {
			controllerutil.SetControllerReference(p, service, r.scheme)
			err = r.client.Create(context.TODO(), service)
			if err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
		}
	}

	pdb := pravega.MakeSegmentstorePodDisruptionBudget(p)
	controllerutil.SetControllerReference(p, pdb, r.scheme)
	err = r.client.Create(context.TODO(), pdb)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	configMap := pravega.MakeSegmentstoreConfigMap(p)
	controllerutil.SetControllerReference(p, configMap, r.scheme)
	err = r.client.Create(context.TODO(), configMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	statefulSet := pravega.MakeSegmentStoreStatefulSet(p)
	controllerutil.SetControllerReference(p, statefulSet, r.scheme)
	if statefulSet.Spec.VolumeClaimTemplates != nil {
		for i := range statefulSet.Spec.VolumeClaimTemplates {
			controllerutil.SetControllerReference(p, &statefulSet.Spec.VolumeClaimTemplates[i], r.scheme)
		}
	}

	err = r.client.Create(context.TODO(), statefulSet)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		} else {
			sts := &appsv1.StatefulSet{}
			name := p.StatefulSetNameForSegmentstore()
			err := r.client.Get(context.TODO(),
				types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
			if err != nil {
				return err
			}
			owRefs := sts.GetOwnerReferences()
			if hasOldVersionOwnerReference(owRefs) {
				log.Printf("Deleting SSS STS as it has old version owner ref.")
				err = r.client.Delete(context.TODO(), sts)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func hasOldVersionOwnerReference(ownerreference []metav1.OwnerReference) bool {
	for _, value := range ownerreference {
		if value.Kind == "PravegaCluster" && value.APIVersion == "pravega.pravega.io/v1alpha1" {
			return true
		}
	}
	return false
}

func (r *ReconcilePravegaCluster) syncClusterSize(p *pravegav1beta1.PravegaCluster) (err error) {
	/*We skip calling syncSegmentStoreSize() during upgrade/rollback from version 07*/
	if !r.IsClusterUpgradingTo07(p) && !r.IsClusterRollbackingFrom07(p) {
		err = r.syncSegmentStoreSize(p)
		if err != nil {
			return err
		}
	}

	err = r.syncControllerSize(p)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcilePravegaCluster) syncSegmentStoreSize(p *pravegav1beta1.PravegaCluster) (err error) {
	sts := &appsv1.StatefulSet{}
	name := p.StatefulSetNameForSegmentstore()
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
	if err != nil {
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}

	if *sts.Spec.Replicas != p.Spec.Pravega.SegmentStoreReplicas {
		scaleDown := int32(0)
		if p.Spec.Pravega.SegmentStoreReplicas < *sts.Spec.Replicas {
			scaleDown = *sts.Spec.Replicas - p.Spec.Pravega.SegmentStoreReplicas
		}
		sts.Spec.Replicas = &(p.Spec.Pravega.SegmentStoreReplicas)
		err = r.client.Update(context.TODO(), sts)
		if err != nil {
			return fmt.Errorf("failed to update size of stateful-set (%s): %v", sts.Name, err)
		}

		/*We skip calling syncStatefulSetPvc() during upgrade/rollback from version 07*/
		if !r.IsClusterUpgradingTo07(p) && !r.IsClusterRollbackingFrom07(p) {
			err = r.syncStatefulSetPvc(sts)
			if err != nil {
				return fmt.Errorf("failed to sync pvcs of stateful-set (%s): %v", sts.Name, err)
			}
		}

		if p.Spec.ExternalAccess.Enabled && scaleDown > 0 {
			err = r.syncStatefulSetExternalServices(sts)
			if err != nil {
				return fmt.Errorf("failed to sync external svcs of stateful-set (%s): %v", sts.Name, err)
			}
		}
	}
	return nil
}

func (r *ReconcilePravegaCluster) syncControllerSize(p *pravegav1beta1.PravegaCluster) (err error) {
	deploy := &appsv1.Deployment{}
	name := p.DeploymentNameForController()
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)
	if err != nil {
		return fmt.Errorf("failed to get deployment (%s): %v", deploy.Name, err)
	}

	if *deploy.Spec.Replicas != p.Spec.Pravega.ControllerReplicas {
		deploy.Spec.Replicas = &(p.Spec.Pravega.ControllerReplicas)
		err = r.client.Update(context.TODO(), deploy)
		if err != nil {
			return fmt.Errorf("failed to update size of deployment (%s): %v", deploy.Name, err)
		}
	}
	return nil
}

func (r *ReconcilePravegaCluster) syncStatefulSetPvc(sts *appsv1.StatefulSet) error {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: sts.Spec.Template.Labels,
	})
	if err != nil {
		return fmt.Errorf("failed to convert label selector: %v", err)
	}

	pvcList := &corev1.PersistentVolumeClaimList{}
	pvclistOps := &client.ListOptions{
		Namespace:     sts.Namespace,
		LabelSelector: selector,
	}
	err = r.client.List(context.TODO(), pvcList, pvclistOps)
	if err != nil {
		return err
	}

	for _, pvcItem := range pvcList.Items {
		if util.IsOrphan(pvcItem.Name, *sts.Spec.Replicas) {
			pvcDelete := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvcItem.Name,
					Namespace: pvcItem.Namespace,
				},
			}

			err = r.client.Delete(context.TODO(), pvcDelete)
			if err != nil {
				return fmt.Errorf("failed to delete pvc: %v", err)
			}
		}
	}
	return nil
}

func (r *ReconcilePravegaCluster) syncStatefulSetExternalServices(sts *appsv1.StatefulSet) error {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: sts.Spec.Template.Labels,
	})
	if err != nil {
		return fmt.Errorf("failed to convert label selector: %v", err)
	}

	serviceList := &corev1.ServiceList{}
	servicelistOps := &client.ListOptions{
		Namespace:     sts.Namespace,
		LabelSelector: selector,
	}
	err = r.client.List(context.TODO(), serviceList, servicelistOps)
	if err != nil {
		return err
	}

	for _, svcItem := range serviceList.Items {
		if util.IsOrphan(svcItem.Name, *sts.Spec.Replicas) {
			svcDelete := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      svcItem.Name,
					Namespace: svcItem.Namespace,
				},
			}

			err = r.client.Delete(context.TODO(), svcDelete)
			if err != nil {
				return fmt.Errorf("failed to delete svc: %v", err)
			}
		}
	}
	return nil
}

func (r *ReconcilePravegaCluster) reconcileClusterStatus(p *pravegav1beta1.PravegaCluster) error {

	p.Status.Init()

	expectedSize := p.GetClusterExpectedSize()
	listOps := &client.ListOptions{
		Namespace:     p.Namespace,
		LabelSelector: labels.SelectorFromSet(p.LabelsForPravegaCluster()),
	}
	podList := &corev1.PodList{}
	err := r.client.List(context.TODO(), podList, listOps)
	if err != nil {
		return err
	}

	var (
		readyMembers   []string
		unreadyMembers []string
	)

	for _, p := range podList.Items {
		if util.IsPodReady(&p) {
			readyMembers = append(readyMembers, p.Name)
		} else {
			unreadyMembers = append(unreadyMembers, p.Name)
		}
	}

	if len(readyMembers) == expectedSize {
		p.Status.SetPodsReadyConditionTrue()
	} else {
		p.Status.SetPodsReadyConditionFalse()
	}

	p.Status.Replicas = int32(expectedSize)
	p.Status.CurrentReplicas = int32(len(podList.Items))
	p.Status.ReadyReplicas = int32(len(readyMembers))
	p.Status.Members.Ready = readyMembers
	p.Status.Members.Unready = unreadyMembers

	err = r.client.Status().Update(context.TODO(), p)
	if err != nil {
		return fmt.Errorf("failed to update cluster status: %v", err)
	}
	return nil
}

func (r *ReconcilePravegaCluster) rollbackFailedUpgrade(p *pravegav1beta1.PravegaCluster) error {
	if r.isRollbackTriggered(p) {
		// start rollback to previous version
		previousVersion := p.Status.GetLastVersion()
		log.Printf("Rolling back to last cluster version  %v", previousVersion)
		//Rollback cluster to previous version
		return r.rollbackClusterVersion(p, previousVersion)
	}
	return nil
}

func (r *ReconcilePravegaCluster) isRollbackTriggered(p *pravegav1beta1.PravegaCluster) bool {
	if p.Status.IsClusterInUpgradeFailedState() && p.Spec.Version == p.Status.GetLastVersion() {
		return true
	}
	return false
}

//this function will return true only in case of upgrading from a version below 0.7 to pravega version 0.7 or later
func (r *ReconcilePravegaCluster) IsClusterUpgradingTo07(p *pravegav1beta1.PravegaCluster) bool {
	if !util.IsVersionBelow07(p.Spec.Version) && util.IsVersionBelow07(p.Status.CurrentVersion) {
		return true
	}
	return false
}
