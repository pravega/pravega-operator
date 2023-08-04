/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pravegav1beta1 "github.com/pravega/pravega-operator/api/v1beta1"
	"github.com/pravega/pravega-operator/pkg/controller/config"
	"github.com/pravega/pravega-operator/pkg/util"
	log "github.com/sirupsen/logrus"
)

var _ reconcile.Reconciler = &PravegaClusterReconciler{}

// ReconcileTime is the delay between reconciliations
const ReconcileTime = 30 * time.Second

// PravegaClusterReconciler reconciles a PravegaCluster object
type PravegaClusterReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pravegaclusters.pravega.pravega.io,resources=pravegaclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pravegaclusters.pravega.pravega.io,resources=pravegaclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pravegaclusters.pravega.pravega.io,resources=pravegaclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PravegaCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *PravegaClusterReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.Printf("Reconciling PravegaCluster %s/%s\n", request.Namespace, request.Name)

	// Fetch the PravegaCluster instance
	pravegaCluster := &pravegav1beta1.PravegaCluster{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, pravegaCluster)
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
		if err = r.Client.Update(context.TODO(), pravegaCluster); err != nil {
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
	return reconcile.Result{RequeueAfter: ReconcileTime}, nil
}

func (r *PravegaClusterReconciler) run(p *pravegav1beta1.PravegaCluster) (err error) {

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

func (r *PravegaClusterReconciler) reconcileFinalizers(p *pravegav1beta1.PravegaCluster) (err error) {
	currentPravegaCluster := &pravegav1beta1.PravegaCluster{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: p.Name, Namespace: p.Namespace}, currentPravegaCluster)
	if err != nil {
		return fmt.Errorf("failed to get pravega cluster (%s): %v", p.Name, err)
	}
	p.ObjectMeta.ResourceVersion = currentPravegaCluster.ObjectMeta.ResourceVersion
	if p.DeletionTimestamp.IsZero() && !config.DisableFinalizer {
		if !util.ContainsString(p.ObjectMeta.Finalizers, util.ZkFinalizer) {
			p.ObjectMeta.Finalizers = append(p.ObjectMeta.Finalizers, util.ZkFinalizer)
			if err = r.Client.Update(context.TODO(), p); err != nil {
				return fmt.Errorf("failed to add the finalizer (%s): %v", p.Name, err)
			}
		}
	} else {
		if util.ContainsString(p.ObjectMeta.Finalizers, util.ZkFinalizer) {
			p.ObjectMeta.Finalizers = util.RemoveString(p.ObjectMeta.Finalizers, util.ZkFinalizer)
			if err = r.Client.Update(context.TODO(), p); err != nil {
				return fmt.Errorf("failed to update Pravega object (%s): %v", p.Name, err)
			}
			if err = r.cleanUpZookeeperMeta(p); err != nil {
				// emit an event for zk metadata cleanup failure
				message := fmt.Sprintf("failed to cleanup pravega metadata from zookeeper (znode path: /pravega/%s): %v", p.Name, err)
				event := p.NewApplicationEvent("ZKMETA_CLEANUP_ERROR", "ZK Metadata Cleanup Failed", message, "Error")
				pubErr := r.Client.Create(context.TODO(), event)
				if pubErr != nil {
					log.Printf("Error publishing zk metadata cleanup failure event to k8s. %v", pubErr)
				}
				return fmt.Errorf(message)
			}
		}
	}
	return nil
}

func (r *PravegaClusterReconciler) reconcileConfigMap(p *pravegav1beta1.PravegaCluster) (err error) {

	err = r.reconcileControllerConfigMap(p)
	if err != nil {
		return err
	}

	err = r.reconcileSegmentStoreConfigMap(p)
	if err != nil {
		return err
	}

	return nil

}

func (r *PravegaClusterReconciler) reconcileControllerConfigMap(p *pravegav1beta1.PravegaCluster) (err error) {

	currentConfigMap := &corev1.ConfigMap{}
	configMap := MakeControllerConfigMap(p)
	controllerutil.SetControllerReference(p, configMap, r.Scheme)
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: p.ConfigMapNameForController(), Namespace: p.Namespace}, currentConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), configMap)
			if err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
		}
	} else {
		currentConfigMap := &corev1.ConfigMap{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: p.ConfigMapNameForController(), Namespace: p.Namespace}, currentConfigMap)
		eq := util.CompareConfigMap(currentConfigMap, configMap)
		if !eq {
			configMap.ObjectMeta.ResourceVersion = currentConfigMap.ObjectMeta.ResourceVersion
			err := r.Client.Update(context.TODO(), configMap)
			if err != nil {
				return err
			}
			//restarting controller pods
			if !r.checkVersionUpgradeTriggered(p) {
				err = r.restartDeploymentPod(p)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *PravegaClusterReconciler) reconcileSegmentStoreConfigMap(p *pravegav1beta1.PravegaCluster) (err error) {
	currentConfigMap := &corev1.ConfigMap{}
	segmentStorePortUpdated := false
	configMap := MakeSegmentstoreConfigMap(p)
	controllerutil.SetControllerReference(p, configMap, r.Scheme)
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: p.ConfigMapNameForSegmentstore(), Namespace: p.Namespace}, currentConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), configMap)
			if err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
		}
	} else {
		currentConfigMap := &corev1.ConfigMap{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: p.ConfigMapNameForSegmentstore(), Namespace: p.Namespace}, currentConfigMap)
		eq := util.CompareConfigMap(currentConfigMap, configMap)
		if !eq {
			configMap.ObjectMeta.ResourceVersion = currentConfigMap.ObjectMeta.ResourceVersion
			segmentStorePortUpdated = r.checkSegmentStorePortUpdated(p, currentConfigMap)
			err := r.Client.Update(context.TODO(), configMap)
			if err != nil {
				return err
			}
			//restarting sts pods
			if !r.checkVersionUpgradeTriggered(p) && !segmentStorePortUpdated {
				err = r.restartStsPod(p)

				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func (r *PravegaClusterReconciler) checkSegmentStorePortUpdated(p *pravegav1beta1.PravegaCluster, cm *corev1.ConfigMap) bool {
	if val, ok := p.Spec.Pravega.Options["pravegaservice.service.listener.port"]; ok {
		eq := false
		data := strings.Split(cm.Data["JAVA_OPTS"], " ")
		key := fmt.Sprintf("-Dpravegaservice.service.listener.port=%v", val)
		for _, checkstring := range data {
			if checkstring == key {
				eq = true
			}
		}
		if !eq {
			return true
		}
	}
	return false
}

func (r *PravegaClusterReconciler) checkVersionUpgradeTriggered(p *pravegav1beta1.PravegaCluster) bool {
	currentPravegaCluster := &pravegav1beta1.PravegaCluster{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: p.Name, Namespace: p.Namespace}, currentPravegaCluster)
	if err == nil && currentPravegaCluster.Status.CurrentVersion != p.Spec.Version {
		return true
	}
	return false
}

func (r *PravegaClusterReconciler) reconcilePdb(p *pravegav1beta1.PravegaCluster) (err error) {

	err = r.reconcileControllerPdb(p)
	if err != nil {
		return err
	}

	err = r.reconcileSegmentStorePdb(p)
	if err != nil {
		return err
	}

	return nil

}

func (r *PravegaClusterReconciler) reconcileControllerPdb(p *pravegav1beta1.PravegaCluster) (err error) {

	pdb := MakeControllerPodDisruptionBudget(p)
	controllerutil.SetControllerReference(p, pdb, r.Scheme)
	err = r.Client.Create(context.TODO(), pdb)

	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	currentPdb := &policyv1.PodDisruptionBudget{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: pdb.Name, Namespace: p.Namespace}, currentPdb)
	if err != nil {
		return err
	}
	return r.updatePdb(currentPdb, pdb)
}

func (r *PravegaClusterReconciler) reconcileSegmentStorePdb(p *pravegav1beta1.PravegaCluster) (err error) {
	pdb := MakeSegmentstorePodDisruptionBudget(p)
	controllerutil.SetControllerReference(p, pdb, r.Scheme)
	err = r.Client.Create(context.TODO(), pdb)

	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	currentPdb := &policyv1.PodDisruptionBudget{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: pdb.Name, Namespace: p.Namespace}, currentPdb)
	if err != nil {
		return err
	}
	return r.updatePdb(currentPdb, pdb)
}

func (r *PravegaClusterReconciler) updatePdb(currentPdb *policyv1.PodDisruptionBudget, newPdb *policyv1.PodDisruptionBudget) (err error) {

	if !reflect.DeepEqual(currentPdb.Spec.MaxUnavailable, newPdb.Spec.MaxUnavailable) {
		currentPdb.Spec.MaxUnavailable = newPdb.Spec.MaxUnavailable
		currentPdb.Spec.MinAvailable = newPdb.Spec.MinAvailable
		err = r.Client.Update(context.TODO(), currentPdb)
		if err != nil {
			return fmt.Errorf("failed to update pdb (%s): %v", currentPdb.Name, err)
		}
	}
	return nil
}

func (r *PravegaClusterReconciler) reconcileService(p *pravegav1beta1.PravegaCluster) (err error) {

	err = r.reconcileControllerService(p)
	if err != nil {
		return err
	}

	err = r.reconcileSegmentStoreService(p)
	if err != nil {
		return err
	}

	return nil

}

func (r *PravegaClusterReconciler) reconcileControllerService(p *pravegav1beta1.PravegaCluster) (err error) {

	service := MakeControllerService(p)
	controllerutil.SetControllerReference(p, service, r.Scheme)
	currentService := &corev1.Service{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, currentService)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), service)
			if err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
		} else {
			return err
		}
	} else {
		currentService.ObjectMeta.Labels = service.ObjectMeta.Labels
		currentService.ObjectMeta.Annotations = service.ObjectMeta.Annotations
		currentService.Spec.Selector = service.Spec.Selector
		err = r.Client.Update(context.TODO(), currentService)
		if err != nil {
			return fmt.Errorf("failed to update  service (%s): %v", service.Name, err)
		}
	}
	return nil
}

func (r *PravegaClusterReconciler) reconcileSegmentStoreService(p *pravegav1beta1.PravegaCluster) (err error) {
	headlessService := MakeSegmentStoreHeadlessService(p)
	controllerutil.SetControllerReference(p, headlessService, r.Scheme)
	currentService := &corev1.Service{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: headlessService.Name, Namespace: p.Namespace}, currentService)

	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), headlessService)
			if err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
		}
	} else {

		if currentService.Spec.Ports[0].Port != headlessService.Spec.Ports[0].Port {
			currentService.Spec.Ports[0].Port = headlessService.Spec.Ports[0].Port
			currentService.Spec.Ports[0].TargetPort = headlessService.Spec.Ports[0].TargetPort
			err = r.Client.Update(context.TODO(), currentService)
			if err != nil {
				return fmt.Errorf("failed to update headless service port (%s): %v", currentService.Name, err)
			}
		}
		if len(currentService.Spec.Ports) == 1 {
			currentService.Spec.Ports = append(currentService.Spec.Ports, headlessService.Spec.Ports[1])
			err = r.Client.Update(context.TODO(), currentService)
			if err != nil {
				return fmt.Errorf("failed to update headless service admin port (%s): %v", currentService.Name, err)
			}
		} else if currentService.Spec.Ports[1].Port != headlessService.Spec.Ports[1].Port {
			currentService.Spec.Ports[1].Port = headlessService.Spec.Ports[1].Port
			currentService.Spec.Ports[1].TargetPort = headlessService.Spec.Ports[1].TargetPort
			err = r.Client.Update(context.TODO(), currentService)
			if err != nil {
				return fmt.Errorf("failed to update headless service admin port (%s): %v", currentService.Name, err)
			}
		}
	}

	if p.Spec.ExternalAccess.Enabled {
		currentservice := &corev1.Service{}
		services := MakeSegmentStoreExternalServices(p)
		for _, service := range services {
			controllerutil.SetControllerReference(p, service, r.Scheme)
			err := r.Client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, currentservice)
			if err != nil {
				if errors.IsNotFound(err) {
					err = r.Client.Create(context.TODO(), service)
					if err != nil && !errors.IsAlreadyExists(err) {
						return err
					}
				}
			} else {
				if service.Spec.Ports[0].Port != currentservice.Spec.Ports[0].Port {
					currentservice.Spec.Ports[0].Port = service.Spec.Ports[0].Port
					currentservice.Spec.Ports[0].TargetPort = service.Spec.Ports[0].TargetPort
					err = r.Client.Update(context.TODO(), currentservice)
					if err != nil {
						return fmt.Errorf("failed to update external service port (%s): %v", currentservice.Name, err)
					}
				}
				if len(currentservice.Spec.Ports) == 1 {
					currentservice.Spec.Ports = append(currentservice.Spec.Ports, service.Spec.Ports[1])
					err = r.Client.Update(context.TODO(), currentservice)
					if err != nil {
						return fmt.Errorf("failed to update external service admin port (%s): %v", currentservice.Name, err)
					}
				} else if service.Spec.Ports[1].Port != currentservice.Spec.Ports[1].Port {
					currentservice.Spec.Ports[1].Port = service.Spec.Ports[1].Port
					currentservice.Spec.Ports[1].TargetPort = service.Spec.Ports[1].TargetPort
					err = r.Client.Update(context.TODO(), currentservice)
					if err != nil {
						return fmt.Errorf("failed to update external service admin port (%s): %v", currentservice.Name, err)
					}
				}

				eq := reflect.DeepEqual(currentservice.Annotations["external-dns.alpha.kubernetes.io/hostname"], service.Annotations["external-dns.alpha.kubernetes.io/hostname"])
				if !eq {
					err := r.Client.Delete(context.TODO(), currentservice)
					if err != nil {
						return err
					}
					err = r.Client.Create(context.TODO(), service)
					if err != nil && !errors.IsAlreadyExists(err) {
						return err
					}
					pod := &corev1.Pod{}
					err = r.Client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, pod)
					if err != nil {
						return err
					}
					err = r.Client.Delete(context.TODO(), pod)
					if err != nil {
						return err
					}
					start := time.Now()
					err = r.Client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, pod)
					for err == nil && util.IsPodReady(pod) {
						if time.Since(start) > 10*time.Minute {
							return fmt.Errorf("failed to delete Segmentstore pod (%s) for 10 mins ", pod.Name)
						}
						err = r.Client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, pod)
						log.Printf("waiting for %v pod to be deleted", pod.Name)
					}
					start = time.Now()
					err = r.Client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, pod)
					for err == nil && !util.IsPodReady(pod) {
						if time.Since(start) > 10*time.Minute {
							return fmt.Errorf("failed to get Segmentstore pod (%s) as ready for 10 mins ", pod.Name)
						}
						err = r.Client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: p.Namespace}, pod)
						log.Printf("waiting for %v pod to be in ready state", pod.Name)
					}
				}
			}
		}

	}
	return nil
}

func (r *PravegaClusterReconciler) cleanUpZookeeperMeta(p *pravegav1beta1.PravegaCluster) (err error) {
	if err = p.WaitForClusterToTerminate(r.Client); err != nil {
		return fmt.Errorf("failed to wait for cluster pods termination (%s): %v", p.Name, err)
	}

	if err = util.DeleteAllZnodes(p.Spec.ZookeeperUri, p.Name); err != nil {
		return fmt.Errorf("failed to delete zookeeper znodes for (%s): %v", p.Name, err)
	}
	return nil
}

func (r *PravegaClusterReconciler) deployCluster(p *pravegav1beta1.PravegaCluster) (err error) {
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

		if !util.IsVersionBelow(p.Spec.Version, "0.7.0") {
			newsts := &appsv1.StatefulSet{}
			name := p.StatefulSetNameForSegmentstoreAbove07()
			err = r.Client.Get(context.TODO(),
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

func (r *PravegaClusterReconciler) deleteSTS(p *pravegav1beta1.PravegaCluster) error {
	// We should be here only in case of Pravega CR version migration
	// to version 0.5.0 from version 0.4.x
	sts := &appsv1.StatefulSet{}
	stsName := p.StatefulSetNameForSegmentstoreBelow07()
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: stsName, Namespace: p.Namespace}, sts)
	if err != nil {
		if errors.IsNotFound(err) {
			// nothing to do since old STS was not found
			return nil
		}
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}
	// delete sts, if found
	r.Client.Delete(context.TODO(), sts)
	log.Printf("Deleted old SegmentStore STS %s", sts.Name)
	return nil
}

func (r *PravegaClusterReconciler) deletePVC(p *pravegav1beta1.PravegaCluster) error {
	numPvcs := int(p.Spec.Pravega.SegmentStoreReplicas)
	for i := 0; i < numPvcs; i++ {
		pvcName := "cache-" + p.StatefulSetNameForSegmentstoreBelow07() + "-" + strconv.Itoa(i)
		pvc := &corev1.PersistentVolumeClaim{}
		err := r.Client.Get(context.TODO(),
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
		err = r.Client.Delete(context.TODO(), pvcDelete)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PravegaClusterReconciler) deleteOldSegmentStoreIfExists(p *pravegav1beta1.PravegaCluster) error {
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
			err := r.Client.Get(context.TODO(), types.NamespacedName{Name: svcName, Namespace: p.Namespace}, extService)
			if err != nil {
				if errors.IsNotFound(err) {
					// nothing to do since old STS was not found
					return nil
				}
				return fmt.Errorf("failed to get external service (%s): %v", svcName, err)
			}
			r.Client.Delete(context.TODO(), extService)
			log.Printf("Deleted old SegmentStore external service %s", extService)
		}
	}
	return nil
}

func (r *PravegaClusterReconciler) deployController(p *pravegav1beta1.PravegaCluster) (err error) {

	deployment := MakeControllerDeployment(p)
	controllerutil.SetControllerReference(p, deployment, r.Scheme)
	err = r.Client.Create(context.TODO(), deployment)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	} else if errors.IsAlreadyExists(err) {
		foundDeploy := &appsv1.Deployment{}
		name := p.DeploymentNameForController()
		err := r.Client.Get(context.TODO(),
			types.NamespacedName{Name: name, Namespace: p.Namespace}, foundDeploy)
		if err != nil {
			return err
		}

		if !r.checkVersionUpgradeTriggered(p) && !r.isRollbackTriggered(p) {
			foundDeploy.Spec.Template = deployment.Spec.Template
			err = r.Client.Update(context.TODO(), foundDeploy)
			if err != nil {
				return fmt.Errorf("failed to update deployment set: %v", err)
			}

		}
	}
	return nil
}

func (r *PravegaClusterReconciler) deploySegmentStore(p *pravegav1beta1.PravegaCluster) (err error) {
	statefulSet := MakeSegmentStoreStatefulSet(p)
	controllerutil.SetControllerReference(p, statefulSet, r.Scheme)
	if statefulSet.Spec.VolumeClaimTemplates != nil {
		for i := range statefulSet.Spec.VolumeClaimTemplates {
			controllerutil.SetControllerReference(p, &statefulSet.Spec.VolumeClaimTemplates[i], r.Scheme)
		}
	}

	err = r.Client.Create(context.TODO(), statefulSet)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		} else {
			sts := &appsv1.StatefulSet{}
			name := p.StatefulSetNameForSegmentstore()
			err := r.Client.Get(context.TODO(),
				types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
			if err != nil {
				return err
			}

			if !r.checkVersionUpgradeTriggered(p) && !r.isRollbackTriggered(p) {
				originalsts := sts.DeepCopy()
				sts.Spec.Template = statefulSet.Spec.Template
				err = r.Client.Update(context.TODO(), sts)
				if err != nil {
					return fmt.Errorf("failed to update stateful set: %v", err)
				}

				if !reflect.DeepEqual(originalsts.Spec.Template, sts.Spec.Template) {
					err = r.restartStsPod(p)
					if err != nil {
						return err
					}
				}
			}

			owRefs := sts.GetOwnerReferences()
			if hasOldVersionOwnerReference(owRefs) {
				log.Printf("Deleting SSS STS as it has old version owner ref.")
				err = r.Client.Delete(context.TODO(), sts)
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

func (r *PravegaClusterReconciler) restartStsPod(p *pravegav1beta1.PravegaCluster) error {

	currentSts := &appsv1.StatefulSet{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: p.StatefulSetNameForSegmentstore(), Namespace: p.Namespace}, currentSts)
	if err != nil {
		return err
	}
	labels := p.LabelsForPravegaCluster()
	labels["component"] = "pravega-segmentstore"
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: labels,
	})
	if err != nil {
		return fmt.Errorf("failed to convert label selector: %v", err)
	}
	podList := &corev1.PodList{}
	podlistOps := &client.ListOptions{
		Namespace:     currentSts.Namespace,
		LabelSelector: selector,
	}
	err = r.Client.List(context.TODO(), podList, podlistOps)
	if err != nil {
		return err
	}

	for _, podItem := range podList.Items {
		err := r.Client.Delete(context.TODO(), &podItem)
		if err != nil {
			return err
		} else {
			start := time.Now()
			pod := &corev1.Pod{}
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			for util.IsPodReady(pod) {
				if time.Since(start) > 10*time.Minute {
					return fmt.Errorf("failed to delete Segmentstore pod (%s) for 10 mins ", podItem.ObjectMeta.Name)
				}
				err = r.Client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			}
			start = time.Now()
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			for !util.IsPodReady(pod) {
				if time.Since(start) > 10*time.Minute {
					return fmt.Errorf("failed to get Segmentstore pod (%s) as ready for 10 mins ", podItem.ObjectMeta.Name)
				}
				err = r.Client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			}
		}
	}
	return nil
}

func (r *PravegaClusterReconciler) restartDeploymentPod(p *pravegav1beta1.PravegaCluster) error {

	currentDeployment := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: p.DeploymentNameForController(), Namespace: p.Namespace}, currentDeployment)
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
	err = r.Client.List(context.TODO(), podList, podlistOps)
	if err != nil {
		return err
	}

	for _, podItem := range podList.Items {
		err := r.Client.Delete(context.TODO(), &podItem)
		if err != nil {
			return err
		} else {
			start := time.Now()
			pod := &corev1.Pod{}
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			for util.IsPodReady(pod) {
				if time.Since(start) > 10*time.Minute {
					return fmt.Errorf("failed to delete controller pod (%s) for 10 mins ", podItem.ObjectMeta.Name)
				}
				err = r.Client.Get(context.TODO(), types.NamespacedName{Name: podItem.ObjectMeta.Name, Namespace: podItem.ObjectMeta.Namespace}, pod)
			}
			deploy := &appsv1.Deployment{}
			name := p.DeploymentNameForController()
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)
			if err != nil {
				return fmt.Errorf("failed to get deployment (%s): %v", deploy.Name, err)
			}
			start = time.Now()
			for deploy.Status.ReadyReplicas != deploy.Status.Replicas {
				if time.Since(start) > 10*time.Minute {
					return fmt.Errorf("failed to make controller pod ready for 10 mins ")
				}
				err = r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)
				if err != nil {
					return fmt.Errorf("failed to get deployment (%s): %v", deploy.Name, err)
				}
			}
		}
	}
	return nil
}

func (r *PravegaClusterReconciler) syncClusterSize(p *pravegav1beta1.PravegaCluster) (err error) {
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

func (r *PravegaClusterReconciler) syncSegmentStoreSize(p *pravegav1beta1.PravegaCluster) (err error) {
	sts := &appsv1.StatefulSet{}
	name := p.StatefulSetNameForSegmentstore()
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
	if err != nil {
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}

	if *sts.Spec.Replicas != p.Spec.Pravega.SegmentStoreReplicas {
		scaleDown := int32(0)
		if p.Spec.Pravega.SegmentStoreReplicas < *sts.Spec.Replicas {
			scaleDown = *sts.Spec.Replicas - p.Spec.Pravega.SegmentStoreReplicas
		}
		sts.Spec.Replicas = &(p.Spec.Pravega.SegmentStoreReplicas)
		err = r.Client.Update(context.TODO(), sts)
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

func (r *PravegaClusterReconciler) syncControllerSize(p *pravegav1beta1.PravegaCluster) (err error) {
	deploy := &appsv1.Deployment{}
	name := p.DeploymentNameForController()
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)
	if err != nil {
		return fmt.Errorf("failed to get deployment (%s): %v", deploy.Name, err)
	}

	if *deploy.Spec.Replicas != p.Spec.Pravega.ControllerReplicas {
		deploy.Spec.Replicas = &(p.Spec.Pravega.ControllerReplicas)
		err = r.Client.Update(context.TODO(), deploy)
		if err != nil {
			return fmt.Errorf("failed to update size of deployment (%s): %v", deploy.Name, err)
		}
	}
	return nil
}

func (r *PravegaClusterReconciler) syncStatefulSetPvc(sts *appsv1.StatefulSet) error {
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
	err = r.Client.List(context.TODO(), pvcList, pvclistOps)
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

			err = r.Client.Delete(context.TODO(), pvcDelete)
			if err != nil {
				return fmt.Errorf("failed to delete pvc: %v", err)
			}
		}
	}
	return nil
}

func (r *PravegaClusterReconciler) syncStatefulSetExternalServices(sts *appsv1.StatefulSet) error {
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
	err = r.Client.List(context.TODO(), serviceList, servicelistOps)
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

			err = r.Client.Delete(context.TODO(), svcDelete)
			if err != nil {
				return fmt.Errorf("failed to delete svc: %v", err)
			}
		}
	}
	return nil
}

func (r *PravegaClusterReconciler) reconcileClusterStatus(p *pravegav1beta1.PravegaCluster) error {

	p.Status.Init()

	expectedSize := p.GetClusterExpectedSize()
	listOps := &client.ListOptions{
		Namespace:     p.Namespace,
		LabelSelector: labels.SelectorFromSet(p.LabelsForPravegaCluster()),
	}
	podList := &corev1.PodList{}
	err := r.Client.List(context.TODO(), podList, listOps)
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

	err = r.Client.Status().Update(context.TODO(), p)
	if err != nil {
		return fmt.Errorf("failed to update cluster status: %v", err)
	}
	return nil
}

func (r *PravegaClusterReconciler) rollbackFailedUpgrade(p *pravegav1beta1.PravegaCluster) error {
	if r.isRollbackTriggered(p) {
		// start rollback to previous version
		previousVersion := p.Status.GetLastVersion()
		log.Printf("Rolling back to last cluster version  %v", previousVersion)
		//Rollback cluster to previous version
		return r.rollbackClusterVersion(p, previousVersion)
	}
	return nil
}

func (r *PravegaClusterReconciler) isRollbackTriggered(p *pravegav1beta1.PravegaCluster) bool {
	if p.Status.IsClusterInUpgradeFailedState() && p.Spec.Version == p.Status.GetLastVersion() {
		return true
	}
	return false
}

// this function will return true only in case of upgrading from a version below 0.7 to pravega version 0.7 or later
func (r *PravegaClusterReconciler) IsClusterUpgradingTo07(p *pravegav1beta1.PravegaCluster) bool {
	if !util.IsVersionBelow(p.Spec.Version, "0.7.0") && util.IsVersionBelow(p.Status.CurrentVersion, "0.7.0") {
		return true
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *PravegaClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pravegav1beta1.PravegaCluster{}).
		Complete(r)
}
