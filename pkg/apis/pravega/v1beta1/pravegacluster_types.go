/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1beta1

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	k8s "github.com/operator-framework/operator-sdk/pkg/k8sutil"
	bkapi "github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var Mgr manager.Manager

const (
	// DefaultZookeeperUri is the default ZooKeeper URI in the form of "hostname:port"
	DefaultZookeeperUri = "zookeeper-client:2181"

	// DefaultBookkeeperUri is the default ZooKeeper URI in the form of "hostname:port"
	DefaultBookkeeperUri = "bookkeeper-bookie-0.bookkeeper-bookie-headless.default.svc.cluster.local:3181,bookkeeper-bookie-1.bookkeeper-bookie-headless.default.svc.cluster.local:3181,bookkeeper-bookie-2.bookkeeper-bookie-headless.default.svc.cluster.local:3181"

	// DefaultServiceType is the default service type for external access
	DefaultServiceType = corev1.ServiceTypeLoadBalancer

	// DefaultPravegaVersion is the default tag used for for the Pravega
	// Docker image
	DefaultPravegaVersion = "0.9.0"
)

func init() {
	SchemeBuilder.Register(&PravegaCluster{}, &PravegaClusterList{})
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PravegaClusterList contains a list of PravegaCluster
type PravegaClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PravegaCluster `json:"items"`
}

// Generate CRD using kubebuilder
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=pk
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.currentVersion`,description="The current pravega version"
// +kubebuilder:printcolumn:name="Desired Version",type=string,JSONPath=`.spec.version`,description="The desired pravega version"
// +kubebuilder:printcolumn:name="Desired Members",type=integer,JSONPath=`.status.replicas`,description="The number of desired pravega members"
// +kubebuilder:printcolumn:name="Ready Members",type=integer,JSONPath=`.status.readyReplicas`,description="The number of ready pravega members"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// PravegaCluster is the Schema for the pravegaclusters API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type PravegaCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// WithDefaults set default values when not defined in the spec.
func (p *PravegaCluster) WithDefaults() (changed bool) {
	changed = p.Spec.withDefaults(p)
	return changed
}

// ClusterSpec defines the desired state of PravegaCluster
type ClusterSpec struct {
	// ZookeeperUri specifies the hostname/IP address and port in the format
	// "hostname:port".
	// By default, the value "zookeeper-client:2181" is used, that corresponds to the
	// default Zookeeper service created by the Pravega Zookkeeper operator
	// available at: https://github.com/pravega/zookeeper-operator
	// +optional
	ZookeeperUri string `json:"zookeeperUri"`

	// ExternalAccess specifies whether or not to allow external access
	// to clients and the service type to use to achieve it
	// By default, external access is not enabled
	// +optional
	ExternalAccess *ExternalAccess `json:"externalAccess"`

	// TLS is the Pravega security configuration that is passed to the Pravega processes.
	// See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md
	// +optional
	TLS *TLSPolicy `json:"tls,omitempty"`

	// Authentication can be enabled for authorizing all communication from clients to controller and segment store
	// See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md
	Authentication *AuthenticationParameters `json:"authentication,omitempty"`

	// Version is the expected version of the Pravega cluster.
	// The pravega-operator will eventually make the Pravega cluster version
	// equal to the expected version.
	//
	// The version must follow the [semver]( http://semver.org) format, for example "3.2.13".
	// Only Pravega released versions are supported: https://github.com/pravega/pravega/releases
	//
	// If version is not set, default value will be set.
	// +optional
	Version string `json:"version"`

	// BookkeeperUri specifies the hostname/IP address and port in the format
	// "hostname:port".
	// comma delimited list of BK server URLs
	// bookkeeper-bookie-0.bookkeeper-bookie-headless.default:3181,
	// bookkeeper-bookie-1.bookkeeper-bookie-headless.default:3181,
	// bookkeeper-bookie-2.bookkeeper-bookie-headless.default:3181
	// +optional
	BookkeeperUri string `json:"bookkeeperUri"`

	// Pravega configuration
	// +optional
	Pravega *PravegaSpec `json:"pravega"`
}

func (s *ClusterSpec) withDefaults(p *PravegaCluster) (changed bool) {
	if s.ZookeeperUri == "" {
		changed = true
		s.ZookeeperUri = DefaultZookeeperUri
	}

	if s.ExternalAccess == nil {
		changed = true
		s.ExternalAccess = &ExternalAccess{}
	}

	if s.ExternalAccess.withDefaults() {
		changed = true
	}

	if s.TLS == nil {
		changed = true
		s.TLS = &TLSPolicy{
			Static: &StaticTLS{},
		}
	}

	if s.Authentication == nil {
		changed = true
		s.Authentication = &AuthenticationParameters{}
	}

	if s.Version == "" {
		s.Version = DefaultPravegaVersion
		changed = true
	}

	if s.BookkeeperUri == "" {
		s.BookkeeperUri = DefaultBookkeeperUri
		changed = true
	}

	if s.Pravega == nil {
		changed = true
		s.Pravega = &PravegaSpec{}
	}

	if s.Pravega.withDefaults() {
		changed = true
	}

	if s.Pravega.ControllerPodAffinity == nil {
		changed = true
		s.Pravega.ControllerPodAffinity = util.PodAntiAffinity("pravega-controller", p.GetName())
	}

	if s.Pravega.SegmentStorePodAffinity == nil {
		changed = true
		s.Pravega.SegmentStorePodAffinity = util.PodAntiAffinity("pravega-segmentstore", p.GetName())
	}

	if util.IsVersionBelow(s.Version, "0.7.0") && s.Pravega.CacheVolumeClaimTemplate == nil {
		changed = true
		s.Pravega.CacheVolumeClaimTemplate = &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(DefaultPravegaCacheVolumeSize),
				},
			},
		}
	}

	return changed
}

// ExternalAccess defines the configuration of the external access
type ExternalAccess struct {
	// Enabled specifies whether or not external access is enabled
	// By default, external access is not enabled
	// +optional
	Enabled bool `json:"enabled"`

	// Type specifies the service type to achieve external access.
	// Options are "LoadBalancer" and "NodePort".
	// By default, if external access is enabled, it will use "LoadBalancer"
	Type corev1.ServiceType `json:"type,omitempty"`

	// Domain Name to be used for External Access
	// This value is ignored if External Access is disabled
	DomainName string `json:"domainName,omitempty"`
}

func (e *ExternalAccess) withDefaults() (changed bool) {
	if e.Enabled == false && (e.Type != "" || e.DomainName != "") {
		changed = true
		e.Type = ""
		e.DomainName = ""
	}
	return changed
}

type TLSPolicy struct {
	// Static TLS means keys/certs are generated by the user and passed to an operator.
	Static *StaticTLS `json:"static,omitempty"`
}

type StaticTLS struct {
	ControllerSecret   string `json:"controllerSecret,omitempty"`
	SegmentStoreSecret string `json:"segmentStoreSecret,omitempty"`
	CaBundle           string `json:"caBundle,omitempty"`
}

func (tp *TLSPolicy) IsSecureController() bool {
	if tp == nil || tp.Static == nil {
		return false
	}
	return len(tp.Static.ControllerSecret) != 0
}

func (tp *TLSPolicy) IsSecureSegmentStore() bool {
	if tp == nil || tp.Static == nil {
		return false
	}
	return len(tp.Static.SegmentStoreSecret) != 0
}

func (tp *TLSPolicy) IsCaBundlePresent() bool {
	if tp == nil || tp.Static == nil {
		return false
	}
	return len(tp.Static.CaBundle) != 0
}

type AuthenticationParameters struct {
	// Enabled specifies whether or not authentication is enabled
	// By default, authentication is not enabled
	// +optional
	Enabled bool `json:"enabled"`

	// name of Secret containing Password based Authentication Parameters like username, password and acl
	// optional - used only by PasswordAuthHandler for authentication
	PasswordAuthSecret string `json:"passwordAuthSecret,omitempty"`

	// name of secret containg TokenSigningKey
	ControllerTokenSecret string `json:"controllerTokenSecret,omitempty"`

	// name of secret containg TokenSigningKey and AuthToken
	SegmentStoreTokenSecret string `json:"segmentStoreTokenSecret,omitempty"`
}

func (ap *AuthenticationParameters) IsEnabled() bool {
	if ap == nil {
		return false
	}
	return ap.Enabled
}

// ImageSpec defines the fields needed for a Docker repository image
type ImageSpec struct {
	// +optional
	Repository string `json:"repository"`
	// +optional
	PullPolicy corev1.PullPolicy `json:"pullPolicy"`
}

func (src *PravegaCluster) ConvertTo(dstRaw conversion.Hub) error {
	//do nothing here as we never want to move from v1beta1 to v1alpha1
	return nil
}

func (dst *PravegaCluster) convertSpec(srcObj *v1alpha1.PravegaCluster) error {
	dst.Spec.Authentication = &AuthenticationParameters{
		Enabled:            srcObj.Spec.Authentication.Enabled,
		PasswordAuthSecret: srcObj.Spec.Authentication.PasswordAuthSecret,
	}

	dst.Spec.BookkeeperUri = getBookkeeperUri(srcObj)

	if srcObj.Spec.ExternalAccess != nil {
		dst.Spec.ExternalAccess = &ExternalAccess{
			Enabled:    srcObj.Spec.ExternalAccess.Enabled,
			Type:       srcObj.Spec.ExternalAccess.Type,
			DomainName: srcObj.Spec.ExternalAccess.DomainName,
		}
	}

	if srcObj.Spec.TLS != nil {
		dst.Spec.TLS = &TLSPolicy{
			Static: &StaticTLS{
				ControllerSecret:   srcObj.Spec.TLS.Static.ControllerSecret,
				SegmentStoreSecret: srcObj.Spec.TLS.Static.SegmentStoreSecret,
			},
		}
	}

	dst.Spec.Version = srcObj.Spec.Version
	dst.Spec.ZookeeperUri = srcObj.Spec.ZookeeperUri
	dst.Spec.Pravega = &PravegaSpec{
		ControllerReplicas:   srcObj.Spec.Pravega.ControllerReplicas,
		SegmentStoreReplicas: srcObj.Spec.Pravega.SegmentStoreReplicas,
		DebugLogging:         srcObj.Spec.Pravega.DebugLogging,
		Image: &ImageSpec{
			Repository: srcObj.Spec.Pravega.Image.Repository,
			PullPolicy: srcObj.Spec.Pravega.Image.PullPolicy,
		},
		Options:                         srcObj.Spec.Pravega.Options,
		ControllerJvmOptions:            srcObj.Spec.Pravega.ControllerJvmOptions,
		SegmentStoreJVMOptions:          srcObj.Spec.Pravega.SegmentStoreJVMOptions,
		CacheVolumeClaimTemplate:        srcObj.Spec.Pravega.CacheVolumeClaimTemplate,
		ControllerServiceAccountName:    srcObj.Spec.Pravega.ControllerServiceAccountName,
		SegmentStoreServiceAccountName:  srcObj.Spec.Pravega.SegmentStoreServiceAccountName,
		ControllerExternalServiceType:   srcObj.Spec.Pravega.ControllerExternalServiceType,
		ControllerServiceAnnotations:    srcObj.Spec.Pravega.ControllerServiceAnnotations,
		SegmentStoreExternalServiceType: srcObj.Spec.Pravega.SegmentStoreExternalServiceType,
		SegmentStoreServiceAnnotations:  srcObj.Spec.Pravega.SegmentStoreServiceAnnotations,
	}

	if srcObj.Spec.Pravega.Tier2.FileSystem != nil {

		dst.Spec.Pravega.LongTermStorage = &LongTermStorageSpec{
			FileSystem: &FileSystemSpec{
				PersistentVolumeClaim: srcObj.Spec.Pravega.Tier2.FileSystem.PersistentVolumeClaim,
			},
		}
	} else if srcObj.Spec.Pravega.Tier2.Ecs != nil {
		dst.Spec.Pravega.LongTermStorage = &LongTermStorageSpec{
			Ecs: &ECSSpec{
				ConfigUri:   srcObj.Spec.Pravega.Tier2.Ecs.ConfigUri,
				Bucket:      srcObj.Spec.Pravega.Tier2.Ecs.Bucket,
				Prefix:      srcObj.Spec.Pravega.Tier2.Ecs.Prefix,
				Credentials: srcObj.Spec.Pravega.Tier2.Ecs.Credentials,
			},
		}
	} else if srcObj.Spec.Pravega.Tier2.Hdfs != nil {
		dst.Spec.Pravega.LongTermStorage = &LongTermStorageSpec{
			Hdfs: &HDFSSpec{
				Uri:               srcObj.Spec.Pravega.Tier2.Hdfs.Uri,
				Root:              srcObj.Spec.Pravega.Tier2.Hdfs.Root,
				ReplicationFactor: srcObj.Spec.Pravega.Tier2.Hdfs.ReplicationFactor,
			},
		}
	}

	// Controller Resources
	if srcObj.Spec.Pravega.ControllerResources != nil {
		dst.Spec.Pravega.ControllerResources = &corev1.ResourceRequirements{
			Requests: srcObj.Spec.Pravega.ControllerResources.Requests,
			Limits:   srcObj.Spec.Pravega.ControllerResources.Limits,
		}
	}
	// SegmentStore Resources
	if srcObj.Spec.Pravega.SegmentStoreResources != nil {
		dst.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
			Requests: srcObj.Spec.Pravega.SegmentStoreResources.Requests,
			Limits:   srcObj.Spec.Pravega.SegmentStoreResources.Limits,
		}
	}
	return nil
}

func (dst *PravegaCluster) ConvertFrom(srcRaw conversion.Hub) error {
	log.Printf("Converting Pravega CR version from v1alpha1 to v1beta1.")
	//logic for conveting from v1alpha1 to v1beta1
	srcObj := srcRaw.(*v1alpha1.PravegaCluster)
	dst.ObjectMeta = srcObj.ObjectMeta
	err := dst.convertSpec(srcObj)
	if err != nil {
		log.Fatalf("Error converting CR object from version v1alpha1 to v1beta1 %v", err)
		return err
	}
	err = dst.migrateBookkeeper(srcObj)
	if err != nil {
		return err
	}
	err = dst.updatePravegaOwnerReferences()
	if err != nil {
		return err
	}
	log.Print("Version migration completed successfully.")
	return nil
}

func (dst *PravegaCluster) updatePravegaOwnerReferences() error {
	ownerRefs := &[]metav1.OwnerReference{
		{
			APIVersion:         "pravega.pravega.io/v1beta1",
			Kind:               "PravegaCluster",
			Name:               dst.GetName(),
			UID:                dst.GetUID(),
			BlockOwnerDeletion: pointer.BoolPtr(true),
			Controller:         pointer.BoolPtr(true),
		},
	}
	err := dst.updateSSSReferences(*ownerRefs)
	if err != nil {
		return err
	}
	log.Print("Updated SegmentStore OwnerReferences.")
	err = dst.updateControllerReferences(*ownerRefs)
	if err != nil {
		return err
	}
	log.Print("Updated Controller OwnerReferences.")
	return nil
}

func (dst *PravegaCluster) updateSSSReferences(ownerRefs []metav1.OwnerReference) error {
	if util.IsVersionBelow(dst.Spec.Version, "0.7.0") {
		numPvcs := int(dst.Spec.Pravega.SegmentStoreReplicas)
		for i := 0; i < numPvcs; i++ {
			pvcName := "cache-" + dst.StatefulSetNameForSegmentstoreBelow07() + "-" + strconv.Itoa(i)
			pvc := &corev1.PersistentVolumeClaim{}
			err := Mgr.GetClient().Get(context.TODO(),
				types.NamespacedName{Name: pvcName, Namespace: dst.Namespace}, pvc)
			if err != nil {
				return fmt.Errorf("failed to get pvc (%s): %v", pvcName, err)
			}
			pvc.SetOwnerReferences(ownerRefs)
			err = Mgr.GetClient().Update(context.TODO(), pvc)
			if err != nil {
				return err
			}
		}

		if dst.Spec.ExternalAccess.Enabled {
			numSvcs := dst.Spec.Pravega.SegmentStoreReplicas
			for i := int32(0); i < numSvcs; i++ {
				extSvc := &corev1.Service{}
				name := dst.ServiceNameForSegmentStore(i)
				err := Mgr.GetClient().Get(context.TODO(),
					types.NamespacedName{Name: name, Namespace: dst.Namespace}, extSvc)
				if err != nil {
					return err
				}
				extSvc.SetOwnerReferences(ownerRefs)
				err = Mgr.GetClient().Update(context.TODO(), extSvc)
				if err != nil {
					return err
				}
			}
		}
	}

	configmap := &corev1.ConfigMap{}
	name := dst.ConfigMapNameForSegmentstore()
	err := Mgr.GetClient().Get(context.TODO(),
		types.NamespacedName{Name: name, Namespace: dst.Namespace}, configmap)
	if err != nil {
		return err
	}
	configmap.SetOwnerReferences(ownerRefs)
	err = Mgr.GetClient().Update(context.TODO(), configmap)
	if err != nil {
		return err
	}

	headlessservice := &corev1.Service{}
	name = dst.HeadlessServiceNameForSegmentStore()
	err = Mgr.GetClient().Get(context.TODO(),
		types.NamespacedName{Name: name, Namespace: dst.Namespace}, headlessservice)
	if err != nil {
		return err
	}
	headlessservice.SetOwnerReferences(ownerRefs)
	err = Mgr.GetClient().Update(context.TODO(), headlessservice)
	if err != nil {
		return err
	}

	pdb := &policyv1beta1.PodDisruptionBudget{}
	name = dst.PdbNameForSegmentstore()
	err = Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: name, Namespace: dst.Namespace}, pdb)
	if err != nil {
		return err
	}
	pdb.SetOwnerReferences(ownerRefs)
	err = Mgr.GetClient().Update(context.TODO(), pdb)
	if err != nil {
		return err
	}
	return nil
}

func (dst *PravegaCluster) updateControllerReferences(ownerRefs []metav1.OwnerReference) error {
	configmap := &corev1.ConfigMap{}
	name := dst.ConfigMapNameForController()
	err := Mgr.GetClient().Get(context.TODO(),
		types.NamespacedName{Name: name, Namespace: dst.Namespace}, configmap)
	if err != nil {
		return err
	}
	configmap.SetOwnerReferences(ownerRefs)
	err = Mgr.GetClient().Update(context.TODO(), configmap)
	if err != nil {
		return err
	}

	svc := &corev1.Service{}
	name = dst.ServiceNameForController()

	err = Mgr.GetClient().Get(context.TODO(),
		types.NamespacedName{Name: name, Namespace: dst.Namespace}, svc)
	if err != nil {
		return err
	}
	svc.SetOwnerReferences(ownerRefs)
	err = Mgr.GetClient().Update(context.TODO(), svc)
	if err != nil {
		return err
	}

	pdb := &policyv1beta1.PodDisruptionBudget{}
	name = dst.PdbNameForController()
	err = Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: name, Namespace: dst.Namespace}, pdb)
	if err != nil {
		return err
	}
	pdb.SetOwnerReferences(ownerRefs)
	err = Mgr.GetClient().Update(context.TODO(), pdb)
	if err != nil {
		return err
	}

	deploy := &appsv1.Deployment{}
	name = dst.DeploymentNameForController()
	err = Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: name, Namespace: dst.Namespace}, deploy)
	if err != nil {
		return fmt.Errorf("failed to get deployment (%s): %v", name, err)
	}
	deploy.SetOwnerReferences(ownerRefs)
	err = Mgr.GetClient().Update(context.TODO(), deploy)
	if err != nil {
		return err
	}
	return nil
}

func createConfigMap(p *v1alpha1.PravegaCluster) error {
	cmName := fmt.Sprintf("%s-configmap", p.Name)
	cfgMap := &corev1.ConfigMap{}
	err := Mgr.GetClient().Get(context.TODO(),
		types.NamespacedName{Name: cmName, Namespace: p.Namespace}, cfgMap)
	if err != nil {
		if errors.IsNotFound(err) {
			cm := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: p.Namespace,
				},
				Data: map[string]string{
					"PRAVEGA_CLUSTER_NAME": p.Name,
					"WAIT_FOR":             p.Spec.ZookeeperUri,
				},
			}
			err = Mgr.GetClient().Create(context.TODO(), cm)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

func (p *PravegaCluster) migrateBookkeeper(srcObj *v1alpha1.PravegaCluster) error {
	b := &bkapi.BookkeeperCluster{}
	err := Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: p.Name, Namespace: p.Namespace}, b)
	if err != nil {
		if errors.IsNotFound(err) {
			b.WithDefaults()
			b.Name = p.GetName()
			b.Namespace = p.GetNamespace()
			err = createConfigMap(srcObj)
			if err != nil {
				return err
			}
			specCopy(srcObj, b)
			errBk := Mgr.GetClient().Create(context.TODO(), b)
			if errBk != nil {
				log.Fatalf("Error creating BK Cluster object %v", errBk)
				return errBk
			}
		} else {
			return err
		}
	}

	log.Printf("Created Bookkeeper CR by name %s", b.Name)
	configmap := fmt.Sprintf("%s-configmap", srcObj.Name)
	err = migrateConfigMap(srcObj, b, configmap)
	if err != nil {
		log.Fatalf("Error releasing BK CM %s %v", configmap, err)
		return err
	}
	configmap = nameForBookie(srcObj.Name)
	err = migrateConfigMap(srcObj, b, configmap)
	if err != nil {
		log.Fatalf("Error releasing BK CM %s %v", configmap, err)
		return err
	}
	log.Print("Migrated Bookkeeper ConfigMaps.")

	err = migratePVC(srcObj, b)
	if err != nil {
		log.Fatalf("Error releasing BK PVC %v", err)
		return err
	}
	log.Print("Migrated Bookkeeper PVCs.")

	err = migratePDB(srcObj, b)
	if err != nil {
		log.Fatalf("Error releasing BK PDB %v", err)
		return err
	}
	log.Print("Migrated Bookkeeper PDB.")

	err = migrateSTS(srcObj, b)
	if err != nil {
		log.Fatalf("Error releasing BK STS %v", err)
		return err
	}
	log.Print("Deleted Bookkeeper STS (to be recreated by BK-Operator).")
	err = migrateHeadlessSvc(srcObj, b)
	if err != nil {
		log.Fatalf("Error releasing BK HeadlessSvc %v", err)
		return err
	}
	log.Print("Deleted Bookkeeper SVC (to be recreated by BK-Operator).")

	return nil
}

func nameForBookie(clusterName string) string {
	return fmt.Sprintf("%s-bookie", clusterName)
}

func migrateSTS(p *v1alpha1.PravegaCluster, b *bkapi.BookkeeperCluster) error {
	sts := &appsv1.StatefulSet{}
	name := nameForBookie(p.Name)
	err := Mgr.GetClient().Get(context.TODO(),
		types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	appVal, appExists := sts.Spec.Selector.MatchLabels["app"]
	_, pcExists := sts.Spec.Selector.MatchLabels["pravega_cluster"]

	if pcExists && appExists && appVal == "pravega-cluster" {
		return Mgr.GetClient().Delete(context.TODO(), sts)
	}
	return nil
}

func migrateConfigMap(p *v1alpha1.PravegaCluster, b *bkapi.BookkeeperCluster, name string) error {
	configmap := &corev1.ConfigMap{}
	err := Mgr.GetClient().Get(context.TODO(),
		types.NamespacedName{Name: name, Namespace: p.Namespace}, configmap)
	if err != nil {
		return err
	}
	ownerRef := &[]metav1.OwnerReference{
		{
			APIVersion:         "bookkeeper.pravega.io/v1alpha1",
			Kind:               "BookkeeperCluster",
			Name:               b.GetName(),
			UID:                b.GetUID(),
			BlockOwnerDeletion: pointer.BoolPtr(true),
			Controller:         pointer.BoolPtr(true),
		},
	}
	configmap.SetOwnerReferences(*ownerRef)
	err = Mgr.GetClient().Update(context.TODO(), configmap)
	if err != nil {
		return err
	}
	return nil
}

func migrateHeadlessSvc(p *v1alpha1.PravegaCluster, b *bkapi.BookkeeperCluster) error {
	headlessservice := &corev1.Service{}
	name := fmt.Sprintf("%s-bookie-headless", p.Name)
	err := Mgr.GetClient().Get(context.TODO(),
		types.NamespacedName{Name: name, Namespace: p.Namespace}, headlessservice)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	appVal, appExists := headlessservice.Spec.Selector["app"]
	_, pcExists := headlessservice.Spec.Selector["pravega_cluster"]

	if pcExists && appExists && appVal == "pravega-cluster" {
		return Mgr.GetClient().Delete(context.TODO(), headlessservice)
	}
	return nil
}

func movePVC(pvcType string, p *v1alpha1.PravegaCluster, b *bkapi.BookkeeperCluster) error {
	ownerRef := &[]metav1.OwnerReference{
		{
			APIVersion:         "bookkeeper.pravega.io/v1alpha1",
			Kind:               "BookkeeperCluster",
			Name:               b.GetName(),
			UID:                b.GetUID(),
			BlockOwnerDeletion: pointer.BoolPtr(true),
			Controller:         pointer.BoolPtr(true),
		},
	}
	pvc := &corev1.PersistentVolumeClaim{}
	name := pvcType + "-" + p.Name + "-bookie-"
	for i := int32(0); i < p.Spec.Bookkeeper.Replicas; i++ {
		pvcName := name + fmt.Sprintf("%d", i)
		err := Mgr.GetClient().Get(context.TODO(),
			types.NamespacedName{Name: pvcName, Namespace: p.Namespace}, pvc)
		if err != nil {
			return err
		}
		pvc.SetOwnerReferences(*ownerRef)
		err = Mgr.GetClient().Update(context.TODO(), pvc)
		if err != nil {
			return err
		}
	}
	return nil
}

func migratePVC(p *v1alpha1.PravegaCluster, b *bkapi.BookkeeperCluster) error {
	err := movePVC("index", p, b)
	if err != nil {
		return err
	}
	log.Printf("Updating owner reference for index PVCs")
	err = movePVC("journal", p, b)
	if err != nil {
		return err
	}
	log.Printf("Updating owner reference for journal PVCs")
	err = movePVC("ledger", p, b)
	if err != nil {
		return err
	}
	log.Printf("Updating owner reference for ledger PVCs")
	return nil
}

func migratePDB(p *v1alpha1.PravegaCluster, b *bkapi.BookkeeperCluster) error {
	ownerRef := &[]metav1.OwnerReference{
		{
			APIVersion:         "bookkeeper.pravega.io/v1alpha1",
			Kind:               "BookkeeperCluster",
			Name:               b.GetName(),
			UID:                b.GetUID(),
			BlockOwnerDeletion: pointer.BoolPtr(true),
			Controller:         pointer.BoolPtr(true),
		},
	}
	pdb := &policyv1beta1.PodDisruptionBudget{}
	name := nameForBookie(p.Name)
	log.Printf("Updating owner reference for PDB %s", name)
	err := Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, pdb)
	if err != nil {
		return err
	}
	pdb.SetOwnerReferences(*ownerRef)
	return Mgr.GetClient().Update(context.TODO(), pdb)
}

func specCopy(srcObj *v1alpha1.PravegaCluster, b *bkapi.BookkeeperCluster) {
	bkSpec := srcObj.Spec.Bookkeeper
	if bkSpec.Image != nil {
		b.Spec.Image.PullPolicy = bkSpec.Image.PullPolicy
		b.Spec.Image.Repository = bkSpec.Image.Repository
		b.Spec.Image.Tag = bkSpec.Image.Tag
	}

	b.Spec.Replicas = bkSpec.Replicas

	if bkSpec.Storage != nil {
		b.Spec.Storage.IndexVolumeClaimTemplate = bkSpec.Storage.IndexVolumeClaimTemplate
		b.Spec.Storage.JournalVolumeClaimTemplate = bkSpec.Storage.JournalVolumeClaimTemplate
		b.Spec.Storage.LedgerVolumeClaimTemplate = bkSpec.Storage.LedgerVolumeClaimTemplate
	}

	if bkSpec.AutoRecovery != nil {
		b.Spec.AutoRecovery = bkSpec.AutoRecovery
	}

	b.Spec.ServiceAccountName = bkSpec.ServiceAccountName

	if bkSpec.Resources != nil {
		b.Spec.Resources = bkSpec.Resources
	}

	if bkSpec.Options != nil {
		b.Spec.Options = bkSpec.Options
	}

	if bkSpec.BookkeeperJVMOptions != nil {
		b.Spec.JVMOptions.ExtraOpts = bkSpec.BookkeeperJVMOptions.ExtraOpts
		b.Spec.JVMOptions.GcLoggingOpts = bkSpec.BookkeeperJVMOptions.GcLoggingOpts
		b.Spec.JVMOptions.GcOpts = bkSpec.BookkeeperJVMOptions.GcOpts
		b.Spec.JVMOptions.MemoryOpts = bkSpec.BookkeeperJVMOptions.MemoryOpts
	}

	b.Spec.ZookeeperUri = srcObj.Spec.ZookeeperUri
	// name of config-map having pravega configuration
	b.Spec.EnvVars = fmt.Sprintf("%s-configmap", srcObj.Name)
	b.Spec.Version = srcObj.Spec.Version
}

func getBookkeeperUri(srcObj *v1alpha1.PravegaCluster) string {
	bkClusterSize := int(srcObj.Spec.Bookkeeper.Replicas)
	var bookieUrl string = ""
	for i := 0; i < bkClusterSize; i++ {
		bkStsName := fmt.Sprintf("%s-bookie", srcObj.Name)
		bkSvcName := fmt.Sprintf("%s-bookie-headless", srcObj.Name)
		bookieUrl += fmt.Sprintf("%s-%d.%s.%s:3181",
			bkStsName,
			i,
			bkSvcName,
			srcObj.Namespace)
		if i < bkClusterSize-1 {
			bookieUrl += ","
		}
	}
	return bookieUrl
}

var _ webhook.Validator = &PravegaCluster{}

func (p *PravegaCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	log.Print("Registering Webhook")
	return ctrl.NewWebhookManagedBy(mgr).
		For(&PravegaCluster{}).
		Complete()
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (p *PravegaCluster) ValidateCreate() error {
	log.Printf("validate create %s", p.Name)
	err := p.ValidatePravegaVersion()
	if err != nil {
		return err
	}
	err = p.ValidateSegmentStore()
	if err != nil {
		return err
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (p *PravegaCluster) ValidateUpdate(old runtime.Object) error {
	log.Printf("validate update %s", p.Name)
	err := p.ValidatePravegaVersion()
	if err != nil {
		return err
	}
	err = p.validateConfigMap()
	if err != nil {
		return err
	}
	err = p.ValidateSegmentStore()
	if err != nil {
		return err
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (p *PravegaCluster) ValidateDelete() error {
	log.Printf("validate delete %s", p.Name)
	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (p *PravegaCluster) ValidatePravegaVersion() error {
	if p.Spec.Version == "" {
		p.Spec.Version = DefaultPravegaVersion
	}

	requestVersion := p.Spec.Version

	if p.Status.IsClusterInUpgradingState() && requestVersion != p.Status.TargetVersion {
		return fmt.Errorf("failed to process the request, cluster is upgrading")
	}

	if p.Status.IsClusterInRollbackState() {
		if requestVersion != p.Status.GetLastVersion() {
			return fmt.Errorf("failed to process the request, rollback in progress.")
		}
	}
	if p.Status.IsClusterInUpgradeFailedState() {
		if requestVersion != p.Status.GetLastVersion() {
			return fmt.Errorf("Rollback to version %s not supported. Only rollback to version %s is supported.", requestVersion, p.Status.GetLastVersion())
		}
		return nil
	}

	if p.Status.IsClusterInErrorState() {
		return fmt.Errorf("failed to process the request, cluster is in error state.")
	}
	// Check if the request has a valid Pravega version
	normRequestVersion, err := util.NormalizeVersion(requestVersion)
	if err != nil {
		return fmt.Errorf("request version is not in valid format: %v", err)
	}

	if p.Status.CurrentVersion == "" {
		// we're deploying for the very first time
		return nil
	}

	// This is not an upgrade if CurrentVersion == requestVersion
	if p.Status.CurrentVersion == requestVersion {
		return nil
	}
	// This is an upgrade, check if requested version is in the upgrade path
	normFoundVersion, err := util.NormalizeVersion(p.Status.CurrentVersion)
	if err != nil {
		// It should never happen
		return fmt.Errorf("found version is not in valid format, something bad happens: %v", err)
	}

	if match, _ := util.CompareVersions(normRequestVersion, normFoundVersion, "<"); match {
		return fmt.Errorf("downgrading the cluster from version %s to %s is not supported", p.Status.CurrentVersion, requestVersion)
	}
	log.Printf("ValidatePravegaVersion:: normFoundVersion %s", normFoundVersion)

	log.Print("ValidatePravegaVersion:: No error found...returning...")
	return nil
}

func (p *PravegaCluster) validateConfigMap() error {
	configmap := &corev1.ConfigMap{}
	err := Mgr.GetClient().Get(context.TODO(),
		types.NamespacedName{Name: p.ConfigMapNameForController(), Namespace: p.Namespace}, configmap)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		} else {
			return fmt.Errorf("failed to get configmap (%s): %v", configmap.Name, err)
		}
	}
	data := strings.Split(configmap.Data["JAVA_OPTS"], " ")
	eq := false
	if val, ok := p.Spec.Pravega.Options["controller.containerCount"]; ok {
		key := fmt.Sprintf("-Dcontroller.containerCount=%v", val)
		for _, checkstring := range data {
			if checkstring == key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("controller.containerCount should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["controller.container.count"]; ok {
		old_key := fmt.Sprintf("-Dcontroller.containerCount=%v", val)
		new_key := fmt.Sprintf("-Dcontroller.container.count=%v", val)
		for _, checkstring := range data {
			if checkstring == old_key || checkstring == new_key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("controller.container.count should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["pravegaservice.containerCount"]; ok {
		key := fmt.Sprintf("-Dpravegaservice.containerCount=%v", val)
		for _, checkstring := range data {
			if checkstring == key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("pravegaservice.containerCount should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["pravegaservice.container.count"]; ok {
		old_key := fmt.Sprintf("-Dpravegaservice.containerCount=%v", val)
		new_key := fmt.Sprintf("-Dpravegaservice.container.count=%v", val)
		for _, checkstring := range data {
			if checkstring == old_key || checkstring == new_key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("pravegaservice.container.count should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["bookkeeper.bkLedgerPath"]; ok {
		key := fmt.Sprintf("-Dbookkeeper.bkLedgerPath=%v", val)
		for _, checkstring := range data {
			if checkstring == key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("bookkeeper.bkLedgerPath should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["bookkeeper.ledger.path"]; ok {
		old_key := fmt.Sprintf("-Dbookkeeper.bkLedgerPath=%v", val)
		new_key := fmt.Sprintf("-Dbookkeeper.ledger.path=%v", val)
		for _, checkstring := range data {
			if checkstring == old_key || checkstring == new_key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("bookkeeper.ledger.path should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["controller.retention.bucketCount"]; ok {
		key := fmt.Sprintf("-Dcontroller.retention.bucketCount=%v", val)
		for _, checkstring := range data {
			if checkstring == key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("controller.retention.bucketCount should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["controller.retention.bucket.count"]; ok {
		old_key := fmt.Sprintf("-Dcontroller.retention.bucketCount=%v", val)
		new_key := fmt.Sprintf("-Dcontroller.retention.bucket.count=%v", val)
		for _, checkstring := range data {
			if checkstring == old_key || checkstring == new_key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("controller.retention.bucket.count should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["controller.watermarking.bucketCount"]; ok {
		key := fmt.Sprintf("-Dcontroller.watermarking.bucketCount=%v", val)
		for _, checkstring := range data {
			if checkstring == key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("controller.watermarking.bucketCount should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["controller.watermarking.bucket.count"]; ok {
		old_key := fmt.Sprintf("-Dcontroller.watermarking.bucketCount=%v", val)
		new_key := fmt.Sprintf("-Dcontroller.watermarking.bucket.count=%v", val)
		for _, checkstring := range data {
			if checkstring == old_key || checkstring == new_key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("controller.watermarking.bucket.count should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["pravegaservice.dataLogImplementation"]; ok {
		key := fmt.Sprintf("-Dpravegaservice.dataLogImplementation=%v", val)
		for _, checkstring := range data {
			if checkstring == key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("pravegaservice.dataLogImplementation should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["pravegaservice.dataLog.impl.name"]; ok {
		old_key := fmt.Sprintf("-Dpravegaservice.dataLogImplementation=%v", val)
		new_key := fmt.Sprintf("-Dpravegaservice.dataLog.impl.name=%v", val)
		for _, checkstring := range data {
			if checkstring == old_key || checkstring == new_key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("pravegaservice.dataLog.impl.name should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["pravegaservice.storageImplementation"]; ok {
		key := fmt.Sprintf("-Dpravegaservice.storageImplementation=%v", val)
		for _, checkstring := range data {
			if checkstring == key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("pravegaservice.storageImplementation should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["pravegaservice.storage.impl.name"]; ok {
		old_key := fmt.Sprintf("-Dpravegaservice.storageImplementation=%v", val)
		new_key := fmt.Sprintf("-Dpravegaservice.storage.impl.name=%v", val)
		for _, checkstring := range data {
			if checkstring == old_key || checkstring == new_key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("pravegaservice.storage.impl.name should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["storageextra.storageNoOpMode"]; ok {
		key := fmt.Sprintf("-Dstorageextra.storageNoOpMode=%v", val)
		for _, checkstring := range data {
			if checkstring == key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("storageextra.storageNoOpMode should not be modified")
		}
	}
	eq = false
	if val, ok := p.Spec.Pravega.Options["storageextra.noOp.mode.enable"]; ok {
		old_key := fmt.Sprintf("-Dstorageextra.storageNoOpMode=%v", val)
		new_key := fmt.Sprintf("-Dstorageextra.noOp.mode.enable=%v", val)
		for _, checkstring := range data {
			if checkstring == old_key || checkstring == new_key {
				eq = true
			}
		}
		if !eq {
			return fmt.Errorf("storageextra.noOp.mode.enable should not be modified")
		}
	}
	log.Print("validateConfigMap:: No error found...returning...")
	return nil
}

func (p *PravegaCluster) ValidateSegmentStore() error {
	totalMemoryLimitsQuantity := p.Spec.Pravega.SegmentStoreResources.Limits[corev1.ResourceMemory]
	totalMemoryRequestsQuantity := p.Spec.Pravega.SegmentStoreResources.Requests[corev1.ResourceMemory]
	if (resource.Quantity{}) == totalMemoryLimitsQuantity {
		return fmt.Errorf("Missing required value for field spec.pravega.segmentStoreResources.limits.memory")
	}

	totalMemoryLimits := totalMemoryLimitsQuantity.Value()
	totalMemoryRequests := totalMemoryRequestsQuantity.Value()
	if totalMemoryLimits < totalMemoryRequests {
		return fmt.Errorf("spec.pravega.segmentStoreResources.requests.memory value must be less than or equal to spec.pravega.segmentStoreResources.limits.memory")
	}

	cacheSizeString := p.Spec.Pravega.Options["pravegaservice.cache.size.max"]
	if cacheSizeString == "" {
		return fmt.Errorf("Missing required value for option pravegaservice.cache.size.max")
	}
	cacheSizeQuantity := resource.MustParse(cacheSizeString)
	maxDirectMemoryString := ""
	xmxString := ""

	for _, value := range p.Spec.Pravega.SegmentStoreJVMOptions {
		if strings.Contains(value, "-Xmx") {
			xmxString = strings.ToUpper(strings.TrimPrefix(value, "-Xmx")) + "i"
		}

		if strings.Contains(value, "-XX:MaxDirectMemorySize=") {
			maxDirectMemoryString = strings.ToUpper(strings.TrimPrefix(value, "-XX:MaxDirectMemorySize=")) + "i"
		}
	}

	if xmxString == "" {
		return fmt.Errorf("Missing required value for Segment Store JVM Option -Xmx")
	}
	xmxQuantity := resource.MustParse(xmxString)

	if maxDirectMemoryString == "" {
		return fmt.Errorf("Missing required value for Segment Store JVM option -XX:MaxDirectMemorySize")
	}
	maxDirectMemoryQuantity := resource.MustParse(maxDirectMemoryString)

	xmx := xmxQuantity.Value()
	maxDirectMemorySize := maxDirectMemoryQuantity.Value()
	cacheSize := cacheSizeQuantity.Value()

	if totalMemoryLimits <= maxDirectMemorySize+xmx {
		return fmt.Errorf("MaxDirectMemorySize(%v B) along with JVM Xmx value(%v B) is greater than or equal to the total available memory(%v B)!", maxDirectMemorySize, xmx, totalMemoryLimits)
	}

	if maxDirectMemorySize <= xmx {
		return fmt.Errorf("JVM Xmx(%v B) configured is greater than or equal to the JVM MaxDirectMemorySize(%v B) value", xmx, maxDirectMemorySize)
	}

	if maxDirectMemorySize <= cacheSize {
		return fmt.Errorf("Cache size(%v B) configured is greater than or equal to the JVM MaxDirectMemorySize(%v B) value", cacheSize, maxDirectMemorySize)
	}

	return nil
}

//to return name of segmentstore based on the version
func (p *PravegaCluster) StatefulSetNameForSegmentstore() string {
	if util.IsVersionBelow(p.Spec.Version, "0.7.0") {
		return p.StatefulSetNameForSegmentstoreBelow07()
	}
	return p.StatefulSetNameForSegmentstoreAbove07()
}

//if version is above or equals to 0.7 this name will be assigned
func (p *PravegaCluster) StatefulSetNameForSegmentstoreAbove07() string {
	return fmt.Sprintf("%s-%s", p.Name, p.Spec.Pravega.SegmentStoreStsNameSuffix)
}

//if version is below 0.7 this name will be assigned
func (p *PravegaCluster) StatefulSetNameForSegmentstoreBelow07() string {
	return fmt.Sprintf("%s-pravega-segmentstore", p.Name)
}

func (p *PravegaCluster) PravegaControllerServiceURL() string {
	return fmt.Sprintf("tcp://%v.%v:%v", p.ServiceNameForController(), p.Namespace, "9090")
}

func (p *PravegaCluster) LabelsForController() map[string]string {
	labels := p.LabelsForPravegaCluster()
	if p.Spec.Pravega != nil && p.Spec.Pravega.ControllerPodLabels != nil {
		for k, v := range p.Spec.Pravega.ControllerPodLabels {
			labels[k] = v
		}
	}
	labels["component"] = "pravega-controller"
	return labels
}

func (p *PravegaCluster) LabelsForSegmentStore() map[string]string {
	labels := p.LabelsForPravegaCluster()
	if p.Spec.Pravega != nil && p.Spec.Pravega.SegmentStorePodLabels != nil {
		for k, v := range p.Spec.Pravega.SegmentStorePodLabels {
			labels[k] = v
		}
	}
	labels["component"] = "pravega-segmentstore"
	return labels
}

func (p *PravegaCluster) AnnotationsForController() map[string]string {
	annotations := map[string]string{"pravega.version": p.Spec.Version}
	if p.Spec.Pravega != nil && p.Spec.Pravega.ControllerPodAnnotations != nil {
		for k, v := range p.Spec.Pravega.ControllerPodAnnotations {
			annotations[k] = v
		}
	}
	return annotations
}

func (p *PravegaCluster) AnnotationsForSegmentStore() map[string]string {
	annotations := map[string]string{"pravega.version": p.Spec.Version}
	if p.Spec.Pravega != nil && p.Spec.Pravega.SegmentStorePodAnnotations != nil {
		for k, v := range p.Spec.Pravega.SegmentStorePodAnnotations {
			annotations[k] = v
		}
	}
	return annotations
}

func (pravegaCluster *PravegaCluster) LabelsForPravegaCluster() map[string]string {
	return map[string]string{
		"app":             "pravega-cluster",
		"pravega_cluster": pravegaCluster.Name,
	}
}

func (p *PravegaCluster) PdbNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
}

func (p *PravegaCluster) ConfigMapNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
}

func (p *PravegaCluster) ServiceNameForController() string {
	return fmt.Sprintf("%s-%s", p.Name, p.Spec.Pravega.ControllerSvcNameSuffix)
}

func (p *PravegaCluster) ServiceNameForSegmentStore(index int32) string {
	if util.IsVersionBelow(p.Spec.Version, "0.7.0") {
		return p.ServiceNameForSegmentStoreBelow07(index)
	}
	return p.ServiceNameForSegmentStoreAbove07(index)
}

func (p *PravegaCluster) ServiceNameForSegmentStoreBelow07(index int32) string {
	return fmt.Sprintf("%s-pravega-segmentstore-%d", p.Name, index)
}

func (p *PravegaCluster) ServiceNameForSegmentStoreAbove07(index int32) string {
	return fmt.Sprintf("%s-%s-%d", p.Name, p.Spec.Pravega.SegmentStoreStsNameSuffix, index)
}

func (p *PravegaCluster) HeadlessServiceNameForSegmentStore() string {
	return fmt.Sprintf("%s-%s", p.Name, p.Spec.Pravega.SegmentStoreHeadlessSvcNameSuffix)
}

func (p *PravegaCluster) HeadlessServiceNameForBookie() string {
	return fmt.Sprintf("%s-bookie-headless", p.Name)
}

func (p *PravegaCluster) DeploymentNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
}

func (p *PravegaCluster) PdbNameForSegmentstore() string {
	return fmt.Sprintf("%s-segmentstore", p.Name)
}

func (p *PravegaCluster) ConfigMapNameForSegmentstore() string {
	return fmt.Sprintf("%s-pravega-segmentstore", p.Name)
}

func (p *PravegaCluster) GetClusterExpectedSize() (size int) {
	return int(p.Spec.Pravega.ControllerReplicas + p.Spec.Pravega.SegmentStoreReplicas)
}

func (p *PravegaCluster) PravegaImage() (image string) {
	return fmt.Sprintf("%s:%s", p.Spec.Pravega.Image.Repository, p.Spec.Version)
}

func (p *PravegaCluster) PravegaTargetImage() (string, error) {
	if p.Status.TargetVersion == "" {
		return "", fmt.Errorf("target version is not set")
	}
	return fmt.Sprintf("%s:%s", p.Spec.Pravega.Image.Repository, p.Status.TargetVersion), nil
}

// Wait for pods in cluster to be terminated
func (p *PravegaCluster) WaitForClusterToTerminate(kubeClient client.Client) (err error) {
	listOptions := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(p.LabelsForPravegaCluster()),
	}
	err = wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		podList := &corev1.PodList{}
		err = kubeClient.List(context.TODO(), podList, listOptions)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range podList.Items {
			pod := &podList.Items[i]
			names = append(names, pod.Name)
		}

		if len(names) != 0 {
			return false, nil
		}
		return true, nil
	})

	return err
}

func (p *PravegaCluster) NewEvent(name string, reason string, message string, eventType string) *corev1.Event {
	now := metav1.Now()
	operatorName, _ := k8s.GetOperatorName()
	generateName := name + "-"
	event := corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Namespace:    p.Namespace,
			Labels:       p.LabelsForPravegaCluster(),
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion:      p.APIVersion,
			Kind:            "PravegaCluster",
			Name:            p.GetName(),
			Namespace:       p.GetNamespace(),
			ResourceVersion: p.GetResourceVersion(),
			UID:             p.GetUID(),
		},
		Reason:              reason,
		Message:             message,
		FirstTimestamp:      now,
		LastTimestamp:       now,
		Type:                eventType,
		ReportingController: operatorName,
		ReportingInstance:   os.Getenv("POD_NAME"),
	}
	return &event
}

func (p *PravegaCluster) NewApplicationEvent(name string, reason string, message string, eventType string) *corev1.Event {
	now := metav1.Now()
	operatorName, _ := k8s.GetOperatorName()
	generateName := name + "-"
	event := corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Namespace:    p.Namespace,
			Labels:       p.LabelsForPravegaCluster(),
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion: "app.k8s.io/v1beta1",
			Kind:       "Application",
			Name:       "pravega-cluster",
			Namespace:  p.GetNamespace(),
		},
		Reason:              reason,
		Message:             message,
		FirstTimestamp:      now,
		LastTimestamp:       now,
		Type:                eventType,
		ReportingController: operatorName,
		ReportingInstance:   os.Getenv("POD_NAME"),
	}
	return &event
}
