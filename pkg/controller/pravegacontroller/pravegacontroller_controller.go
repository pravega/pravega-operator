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

// Add creates a new PravegaController Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePravegaController{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pravegacontroller-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PravegaController
	err = c.Watch(&source.Kind{Type: &pravegav1beta2.PravegaController{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePravegaController{}

// ReconcilePravegaController reconciles a PravegaController object
type ReconcilePravegaController struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PravegaController object and makes changes based on the state read
// and what is in the PravegaController.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePravegaController) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling PravegaController %s/%s\n", request.Namespace, request.Name)

	// Fetch the PravegaController instance
	PravegaController := &pravegav1beta2.PravegaController{}
	err := r.client.Get(context.TODO(), request.NamespacedName, PravegaController)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Printf("PravegaController %s/%s not found. Ignoring since object must be deleted\n", request.Namespace, request.Name)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Printf("failed to get PravegaController: %v", err)
		return reconcile.Result{}, err
	}

	// Set default configuration for unspecified values
	changed := PravegaController.WithDefaults()
	if changed {
		log.Printf("Setting default settings for pravega-cluster: %s", request.Name)
		if err = r.client.Update(context.TODO(), PravegaController); err != nil {
			log.Printf("Error applying defaults on Pravega Cluster %v", err)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	err = r.run(PravegaController)
	if err != nil {
		log.Printf("failed to reconcile pravega cluster (%s): %v", PravegaController.Name, err)
		return reconcile.Result{}, err
	}
	return reconcile.Result{RequeueAfter: ReconcileTime}, nil
}

func (r *ReconcilePravegaController) run(p *pravegav1beta2.PravegaController) (err error) {

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

func (r *ReconcilePravegaController) reconcileFinalizers(p *pravegav1beta2.PravegaController) (err error) {
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

func (r *ReconcilePravegaController) reconcileConfigMap(p *pravegav1beta2.PravegaController) (err error) {

	currentConfigMap := &corev1.ConfigMap{}
	configMap := pravega.MakeControllerConfigMap(p)
	controllerutil.SetControllerReference(p, configMap, r.scheme)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: p.ConfigMapName(), Namespace: p.Namespace}, currentConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(context.TODO(), configMap)
			if err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
		}
	} else {
		currentConfigMap := &corev1.ConfigMap{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: p.ConfigMapName(), Namespace: p.Namespace}, currentConfigMap)
		eq := util.CompareConfigMap(currentConfigMap, configMap)
		if !eq {
			err := r.client.Update(context.TODO(), configMap)
			if err != nil {
				return err
			}
			//restarting controller pods
			err = r.restartDeploymentPod(p)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ReconcilePravegaController) reconcilePdb(p *pravegav1beta2.PravegaController) (err error) {

	pdb := pravega.MakeControllerPodDisruptionBudget(p)
	controllerutil.SetControllerReference(p, pdb, r.scheme)
	err = r.client.Create(context.TODO(), pdb)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *ReconcilePravegaController) reconcileService(p *pravegav1beta2.PravegaController) (err error) {

	service := pravega.MakeControllerService(p)
	controllerutil.SetControllerReference(p, service, r.scheme)
	err = r.client.Create(context.TODO(), service)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (r *ReconcilePravegaController) cleanUpZookeeperMeta(p *pravegav1beta2.PravegaController) (err error) {
	if err = p.WaitForClusterToTerminate(r.client); err != nil {
		return fmt.Errorf("failed to wait for cluster pods termination (%s): %v", p.Name, err)
	}

	if err = util.DeleteAllZnodes(p.Spec.ZookeeperUri, p.Name); err != nil {
		return fmt.Errorf("failed to delete zookeeper znodes for (%s): %v", p.Name, err)
	}
	return nil
}

func (r *ReconcilePravegaController) deployCluster(p *pravegav1beta2.PravegaController) (err error) {
	err = r.deployController(p)
	if err != nil {
		log.Printf("failed to deploy controller: %v", err)
		return err
	}

	return nil
}

func (r *ReconcilePravegaController) deployController(p *pravegav1beta2.PravegaController) (err error) {

	deployment := pravega.MakeControllerDeployment(p)
	controllerutil.SetControllerReference(p, deployment, r.scheme)
	err = r.client.Create(context.TODO(), deployment)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *ReconcilePravegaController) restartDeploymentPod(p *pravegav1beta2.PravegaController) error {

	currentDeployment := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: p.DeploymentNameForController(), Namespace: p.Namespace}, currentDeployment)
	if err != nil {
		return err
	}
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: currentDeployment.Spec.Template.Labels,
	})
	if err != nil {
		return fmt.Errorf("failed to convert label selector: %v", err)
	}
	podList := &corev1.PodList{}
	podlistOps := &client.ListOptions{
		Namespace:     currentDeployment.Namespace,
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
					return fmt.Errorf("failed to delete controller pod (%s) for 10 mins ", podItem.ObjectMeta.Name)
				}
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			}
			deploy := &appsv1.Deployment{}
			name := p.DeploymentNameForController()
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)
			if err != nil {
				return fmt.Errorf("failed to get deployment (%s): %v", deploy.Name, err)
			}
			start = time.Now()
			for deploy.Status.ReadyReplicas != deploy.Status.Replicas {
				if time.Since(start) > 10*time.Minute {
					return fmt.Errorf("failed to make controller pod ready for 10 mins ")
				}
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)
				if err != nil {
					return fmt.Errorf("failed to get deployment (%s): %v", deploy.Name, err)
				}
			}
		}
	}
	return nil
}

func (r *ReconcilePravegaController) syncClusterSize(p *pravegav1beta2.PravegaController) (err error) {

	err = r.syncControllerSize(p)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcilePravegaController) syncControllerSize(p *pravegav1beta2.PravegaController) (err error) {
	deploy := &appsv1.Deployment{}
	name := p.DeploymentNameForController()
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)
	if err != nil {
		return fmt.Errorf("failed to get deployment (%s): %v", deploy.Name, err)
	}

	if *deploy.Spec.Replicas != p.Spec.Replicas {
		deploy.Spec.Replicas = &(p.Spec.Replicas)
		err = r.client.Update(context.TODO(), deploy)
		if err != nil {
			return fmt.Errorf("failed to update size of deployment (%s): %v", deploy.Name, err)
		}
	}
	return nil
}

func (r *ReconcilePravegaController) reconcileClusterStatus(p *pravegav1beta2.PravegaController) error {

	p.Status.Init()

	expectedSize := p.GetControllerClusterExpectedSize()
	listOps := &client.ListOptions{
		Namespace:     p.Namespace,
		LabelSelector: labels.SelectorFromSet(p.LabelsForPravegaController()),
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

/*func (r *ReconcilePravegaController) rollbackFailedUpgrade(p *pravegav1beta2.PravegaController) error {
	if r.isRollbackTriggered(p) {
		// start rollback to previous version
		previousVersion := p.Status.GetLastVersion()
		log.Printf("Rolling back to last cluster version  %v", previousVersion)
		//Rollback cluster to previous version
		return r.rollbackClusterVersion(p, previousVersion)
	}
	return nil
}

func (r *ReconcilePravegaController) isRollbackTriggered(p *pravegav1beta2.PravegaController) bool {
	if p.Status.IsClusterInUpgradeFailedState() && p.Spec.Version == p.Status.GetLastVersion() {
		return true
	}
	return false
}*/

//this function will return true only in case of upgrading from a version below 0.7 to pravega version 0.7 or later
func (r *ReconcilePravegaController) IsClusterUpgradingTo07(p *pravegav1beta2.PravegaController) bool {
	if !util.IsVersionBelow07(p.Spec.Version) && util.IsVersionBelow07(p.Status.CurrentVersion) {
		return true
	}
	return false
}
