/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package pravegasegmentstore

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	pravegav1beta2 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta2"
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

// Add creates a new PravegaSegmentStore Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePravegaSegmentStore{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("PravegaSegmentStore-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PravegaSegmentStore
	err = c.Watch(&source.Kind{Type: &pravegav1beta2.PravegaSegmentStore{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePravegaSegmentStore{}

// ReconcilePravegaSegmentStore reconciles a PravegaSegmentStore object
type ReconcilePravegaSegmentStore struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PravegaSegmentStore object and makes changes based on the state read
// and what is in the PravegaSegmentStore.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePravegaSegmentStore) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling PravegaSegmentStore %s/%s\n", request.Namespace, request.Name)

	// Fetch the PravegaSegmentStore instance
	PravegaSegmentStore := &pravegav1beta2.PravegaSegmentStore{}
	err := r.client.Get(context.TODO(), request.NamespacedName, PravegaSegmentStore)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Printf("PravegaSegmentStore %s/%s not found. Ignoring since object must be deleted\n", request.Namespace, request.Name)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Printf("failed to get PravegaSegmentStore: %v", err)
		return reconcile.Result{}, err
	}

	// Set default configuration for unspecified values
	changed := PravegaSegmentStore.WithDefaults()
	if changed {
		log.Printf("Setting default settings for pravega-cluster: %s", request.Name)
		if err = r.client.Update(context.TODO(), PravegaSegmentStore); err != nil {
			log.Printf("Error applying defaults on Pravega Cluster %v", err)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	err = r.run(PravegaSegmentStore)
	if err != nil {
		log.Printf("failed to reconcile pravega cluster (%s): %v", PravegaSegmentStore.Name, err)
		return reconcile.Result{}, err
	}
	return reconcile.Result{RequeueAfter: ReconcileTime}, nil
}

func (r *ReconcilePravegaSegmentStore) run(p *pravegav1beta2.PravegaSegmentStore) (err error) {

	err = r.reconcileFinalizers(p)
	if err != nil {
		return fmt.Errorf("failed to reconcile finalizers %v", err)
	}

	err = r.reconcileConfigMap(p)
	if err != nil {
		return fmt.Errorf("failed to reconcile configMap %v", err)
	}

	err = r.reconcilePdb(p)
	if err != nil {
		return fmt.Errorf("failed to reconcile pdb %v", err)
	}

	err = r.reconcileService(p)
	if err != nil {
		return fmt.Errorf("failed to reconcile service %v", err)
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
	/*	err = r.syncClusterVersion(p)
		if err != nil {
			return fmt.Errorf("failed to sync cluster version: %v", err)
		}

		// Rollback
		err = r.rollbackFailedUpgrade(p)
		if err != nil {
			return fmt.Errorf("Rollback attempt failed: %v", err)
		}*/

	err = r.reconcileClusterStatus(p)
	if err != nil {
		return fmt.Errorf("failed to reconcile cluster status: %v", err)
	}
	return nil
}

func (r *ReconcilePravegaSegmentStore) reconcileFinalizers(p *pravegav1beta2.PravegaSegmentStore) (err error) {
	if p.DeletionTimestamp.IsZero() {
		if !util.ContainsString(p.ObjectMeta.Finalizers, util.ZkFinalizer) {
			p.ObjectMeta.Finalizers = append(p.ObjectMeta.Finalizers, util.ZkFinalizer)
			if err = r.client.Update(context.TODO(), p); err != nil {
				return fmt.Errorf("failed to add the finalizer (%s): %v", p.Name, err)
			}
		}
	} else {
		if util.ContainsString(p.ObjectMeta.Finalizers, util.ZkFinalizer) {
			p.ObjectMeta.Finalizers = util.RemoveString(p.ObjectMeta.Finalizers, util.ZkFinalizer)
			if err = r.client.Update(context.TODO(), p); err != nil {
				return fmt.Errorf("failed to update Pravega object (%s): %v", p.Name, err)
			}
			if err = r.cleanUpZookeeperMeta(p); err != nil {
				// emit an event for zk metadata cleanup failure
				message := fmt.Sprintf("failed to cleanup pravega metadata from zookeeper (znode path: /pravega/%s): %v", p.Name, err)
				event := p.NewApplicationEvent("ZKMETA_CLEANUP_ERROR", "ZK Metadata Cleanup Failed", message, "Error")
				pubErr := r.client.Create(context.TODO(), event)
				if pubErr != nil {
					log.Printf("Error publishing zk metadata cleanup failure event to k8s. %v", pubErr)
				}
				return fmt.Errorf(message)
			}
		}
	}
	return nil
}

func (r *ReconcilePravegaSegmentStore) reconcileConfigMap(p *pravegav1beta2.PravegaSegmentStore) (err error) {

	currentConfigMap := &corev1.ConfigMap{}
	configMap := pravega.MakeSegmentstoreConfigMap(p)
	controllerutil.SetControllerReference(p, configMap, r.scheme)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: p.ConfigMapNameForSegmentstore(), Namespace: p.Namespace}, currentConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(context.TODO(), configMap)
			if err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
		}
	} else {
		currentConfigMap := &corev1.ConfigMap{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: p.ConfigMapNameForSegmentstore(), Namespace: p.Namespace}, currentConfigMap)
		eq := util.CompareConfigMap(currentConfigMap, configMap)
		if !eq {
			err := r.client.Update(context.TODO(), configMap)
			if err != nil {
				return err
			}
			//restarting sts pods
			err = r.restartStsPod(p)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ReconcilePravegaSegmentStore) reconcilePdb(p *pravegav1beta2.PravegaSegmentStore) (err error) {

	pdb := pravega.MakeSegmentstorePodDisruptionBudget(p)
	controllerutil.SetControllerReference(p, pdb, r.scheme)
	err = r.client.Create(context.TODO(), pdb)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *ReconcilePravegaSegmentStore) reconcileService(p *pravegav1beta2.PravegaSegmentStore) (err error) {

	headlessService := pravega.MakeSegmentStoreHeadlessService(p)
	controllerutil.SetControllerReference(p, headlessService, r.scheme)
	err = r.client.Create(context.TODO(), headlessService)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	if p.Spec.ExternalAccess.Enabled {
		currentservice := &corev1.Service{}
		services := pravega.MakeSegmentStoreExternalServices(p)
		for _, service := range services {
			controllerutil.SetControllerReference(p, service, r.scheme)
			err := r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, currentservice)
			if err != nil {
				if errors.IsNotFound(err) {
					err = r.client.Create(context.TODO(), service)
					if err != nil && !errors.IsAlreadyExists(err) {
						return err
					}
				}
			} else {
				eq := reflect.DeepEqual(currentservice.Annotations["external-dns.alpha.kubernetes.io/hostname"], service.Annotations["external-dns.alpha.kubernetes.io/hostname"])
				if !eq {
					err := r.client.Delete(context.TODO(), currentservice)
					if err != nil {
						return err
					}
					err = r.client.Create(context.TODO(), service)
					if err != nil && !errors.IsAlreadyExists(err) {
						return err
					}
					pod := &corev1.Pod{}
					err = r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, pod)
					if err != nil {
						return err
					}
					err = r.client.Delete(context.TODO(), pod)
					if err != nil {
						return err
					}
					start := time.Now()
					err = r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, pod)
					for err == nil && util.IsPodReady(pod) {
						if time.Since(start) > 10*time.Minute {
							return fmt.Errorf("failed to delete Segmentstore pod (%s) for 10 mins ", pod.Name)
						}
						err = r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, pod)
						log.Printf("waiting for %v pod to be deleted", pod.Name)
					}
					start = time.Now()
					err = r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, pod)
					for err == nil && !util.IsPodReady(pod) {
						if time.Since(start) > 10*time.Minute {
							return fmt.Errorf("failed to get Segmentstore pod (%s) as ready for 10 mins ", pod.Name)
						}
						err = r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, pod)
						log.Printf("waiting for %v pod to be in ready state", pod.Name)
					}
				}
			}
		}

	}
	return nil
}

func (r *ReconcilePravegaSegmentStore) cleanUpZookeeperMeta(p *pravegav1beta2.PravegaSegmentStore) (err error) {
	if err = p.WaitForClusterToTerminate(r.client); err != nil {
		return fmt.Errorf("failed to wait for cluster pods termination (%s): %v", p.Name, err)
	}

	if err = util.DeleteAllZnodes(p.Spec.ZookeeperUri, p.Name); err != nil {
		return fmt.Errorf("failed to delete zookeeper znodes for (%s): %v", p.Name, err)
	}
	return nil
}

func (r *ReconcilePravegaSegmentStore) deployCluster(p *pravegav1beta2.PravegaSegmentStore) (err error) {

	/*this check is to avoid creation of a new segmentstore when the CurrentVersion is below 07 and target version is above 07
	  as we are doing it in the upgrade path*/
	//if !r.IsClusterUpgradingTo07(p) && !r.IsClusterRollbackingFrom07(p) {

	err = r.deploySegmentStore(p)
	if err != nil {
		log.Printf("failed to deploy segment stor e: %v", err)
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
	//}
	return nil
}

func (r *ReconcilePravegaSegmentStore) deleteSTS(p *pravegav1beta2.PravegaSegmentStore) error {
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

func (r *ReconcilePravegaSegmentStore) deletePVC(p *pravegav1beta2.PravegaSegmentStore) error {
	numPvcs := int(p.Spec.Replicas)
	for i := 0; i < numPvcs; i++ {
		pvcName := "cache-" + p.StatefulSetNameForSegmentstoreBelow07() + "-" + strconv.Itoa(i)
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

func (r *ReconcilePravegaSegmentStore) deleteOldSegmentStoreIfExists(p *pravegav1beta2.PravegaSegmentStore) error {
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
		for i := int32(0); i < p.Spec.Replicas; i++ {
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

func (r *ReconcilePravegaSegmentStore) deploySegmentStore(p *pravegav1beta2.PravegaSegmentStore) (err error) {
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
		if value.Kind == "PravegaSegmentStore" && value.APIVersion == "pravega.pravega.io/v1alpha1" {
			return true
		}
	}
	return false
}

func (r *ReconcilePravegaSegmentStore) restartStsPod(p *pravegav1beta2.PravegaSegmentStore) error {

	currentSts := &appsv1.StatefulSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: p.StatefulSetNameForSegmentstore(), Namespace: p.Namespace}, currentSts)
	if err != nil {
		return err
	}
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: currentSts.Spec.Template.Labels,
	})
	if err != nil {
		return fmt.Errorf("failed to convert label selector: %v", err)
	}
	podList := &corev1.PodList{}
	podlistOps := &client.ListOptions{
		Namespace:     currentSts.Namespace,
		LabelSelector: selector,
	}
	err = r.client.List(context.TODO(), podList, podlistOps)
	if err != nil {
		return err
	}

	for _, podItem := range podList.Items {
		err := r.client.Delete(context.TODO(), &podItem)
		if err != nil {
			return err
		} else {
			start := time.Now()
			pod := &corev1.Pod{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			for util.IsPodReady(pod) {
				if time.Since(start) > 10*time.Minute {
					return fmt.Errorf("failed to delete Segmentstore pod (%s) for 10 mins ", podItem.ObjectMeta.Name)
				}
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			}
			start = time.Now()
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			for !util.IsPodReady(pod) {
				if time.Since(start) > 10*time.Minute {
					return fmt.Errorf("failed to get Segmentstore pod (%s) as ready for 10 mins ", podItem.ObjectMeta.Name)
				}
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			}
		}
	}
	return nil
}

func (r *ReconcilePravegaSegmentStore) syncClusterSize(p *pravegav1beta2.PravegaSegmentStore) (err error) {
	/*We skip calling syncSegmentStoreSize() during upgrade/rollback from version 07*/
	//if !r.IsClusterUpgradingTo07(p) && !r.IsClusterRollbackingFrom07(p) {
	err = r.syncSegmentStoreSize(p)
	if err != nil {
		return err
	}
	//}

	return nil
}

func (r *ReconcilePravegaSegmentStore) syncSegmentStoreSize(p *pravegav1beta2.PravegaSegmentStore) (err error) {
	sts := &appsv1.StatefulSet{}
	name := p.StatefulSetNameForSegmentstore()
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
	if err != nil {
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}

	if *sts.Spec.Replicas != p.Spec.Replicas {
		scaleDown := int32(0)
		if p.Spec.Replicas < *sts.Spec.Replicas {
			scaleDown = *sts.Spec.Replicas - p.Spec.Replicas
		}
		sts.Spec.Replicas = &(p.Spec.Replicas)
		err = r.client.Update(context.TODO(), sts)
		if err != nil {
			return fmt.Errorf("failed to update size of stateful-set (%s): %v", sts.Name, err)
		}

		/*We skip calling syncStatefulSetPvc() during upgrade/rollback from version 07*/
		//if !r.IsClusterUpgradingTo07(p) && !r.IsClusterRollbackingFrom07(p) {
		err = r.syncStatefulSetPvc(sts)
		if err != nil {
			return fmt.Errorf("failed to sync pvcs of stateful-set (%s): %v", sts.Name, err)
		}
		//}

		if p.Spec.ExternalAccess.Enabled && scaleDown > 0 {
			err = r.syncStatefulSetExternalServices(sts)
			if err != nil {
				return fmt.Errorf("failed to sync external svcs of stateful-set (%s): %v", sts.Name, err)
			}
		}
	}
	return nil
}

func (r *ReconcilePravegaSegmentStore) syncStatefulSetPvc(sts *appsv1.StatefulSet) error {
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

func (r *ReconcilePravegaSegmentStore) syncStatefulSetExternalServices(sts *appsv1.StatefulSet) error {
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

func (r *ReconcilePravegaSegmentStore) reconcileClusterStatus(p *pravegav1beta2.PravegaSegmentStore) error {

	p.Status.Init()

	expectedSize := p.Spec.Replicas
	listOps := &client.ListOptions{
		Namespace:     p.Namespace,
		LabelSelector: labels.SelectorFromSet(p.LabelsForPravegaSegmentStore()),
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

	if int32(len(readyMembers)) == expectedSize {
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

/*func (r *ReconcilePravegaSegmentStore) rollbackFailedUpgrade(p *pravegav1beta2.PravegaSegmentStore) error {
	if r.isRollbackTriggered(p) {
		// start rollback to previous version
		previousVersion := p.Status.GetLastVersion()
		log.Printf("Rolling back to last cluster version  %v", previousVersion)
		//Rollback cluster to previous version
		return r.rollbackClusterVersion(p, previousVersion)
	}
	return nil
}*/

func (r *ReconcilePravegaSegmentStore) isRollbackTriggered(p *pravegav1beta2.PravegaSegmentStore) bool {
	if p.Status.IsClusterInUpgradeFailedState() && p.Spec.Version == p.Status.GetLastVersion() {
		return true
	}
	return false
}

//this function will return true only in case of upgrading from a version below 0.7 to pravega version 0.7 or later
func (r *ReconcilePravegaSegmentStore) IsClusterUpgradingTo07(p *pravegav1beta2.PravegaSegmentStore) bool {
	if !util.IsVersionBelow07(p.Spec.Version) && util.IsVersionBelow07(p.Status.CurrentVersion) {
		return true
	}
	return false
}
