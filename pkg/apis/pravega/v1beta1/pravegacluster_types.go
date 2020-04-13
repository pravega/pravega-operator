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
	"bufio"
	"fmt"
	"os"
	"strings"

	k8s "github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	// DefaultZookeeperUri is the default ZooKeeper URI in the form of "hostname:port"
	DefaultZookeeperUri = "zk-client:2181"

	// DefaultBookkeeperUri is the default ZooKeeper URI in the form of "hostname:port"
	DefaultBookkeeperUri = "pravega-bookie-0.pravega-bookie-headless.default.svc.cluster.local:3181,pravega-bookie-1.pravega-bookie-headless.default.svc.cluster.local:3181,pravega-bookie-2.pravega-bookie-headless.default.svc.cluster.local:3181"

	// DefaultServiceType is the default service type for external access
	DefaultServiceType = v1.ServiceTypeLoadBalancer

	// DefaultPravegaVersion is the default tag used for for the Pravega
	// Docker image
	DefaultPravegaVersion = "0.7.0"
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PravegaCluster is the Schema for the pravegaclusters API
// +k8s:openapi-gen=true
type PravegaCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// WithDefaults set default values when not defined in the spec.
func (p *PravegaCluster) WithDefaults() (changed bool) {
	changed = p.Spec.withDefaults()
	return changed
}

// ClusterSpec defines the desired state of PravegaCluster
type ClusterSpec struct {
	// ZookeeperUri specifies the hostname/IP address and port in the format
	// "hostname:port".
	// By default, the value "zk-client:2181" is used, that corresponds to the
	// default Zookeeper service created by the Pravega Zookkeeper operator
	// available at: https://github.com/pravega/zookeeper-operator
	ZookeeperUri string `json:"zookeeperUri"`

	// ExternalAccess specifies whether or not to allow external access
	// to clients and the service type to use to achieve it
	// By default, external access is not enabled
	ExternalAccess *ExternalAccess `json:"externalAccess"`

	// TLS is the Pravega security configuration that is passed to the Pravega processes.
	// See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md
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
	// If version is not set, default is "0.4.0".
	Version string `json:"version"`

	// BookkeeperUri specifies the hostname/IP address and port in the format
	// "hostname:port".
	// comma delimited list of BK server URLs
	//pravega-bookie-0.pravega-bookie-headless.default:3181,
	//pravega-bookie-1.pravega-bookie-headless.default:3181,
	//pravega-bookie-2.pravega-bookie-headless.default:3181
	BookkeeperUri string `json:"bookkeeperUri"`

	// Pravega configuration
	Pravega *PravegaSpec `json:"pravega"`
}

func (s *ClusterSpec) withDefaults() (changed bool) {
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

	return changed
}

// ExternalAccess defines the configuration of the external access
type ExternalAccess struct {
	// Enabled specifies whether or not external access is enabled
	// By default, external access is not enabled
	Enabled bool `json:"enabled"`

	// Type specifies the service type to achieve external access.
	// Options are "LoadBalancer" and "NodePort".
	// By default, if external access is enabled, it will use "LoadBalancer"
	Type v1.ServiceType `json:"type,omitempty"`

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

type AuthenticationParameters struct {
	// Enabled specifies whether or not authentication is enabled
	// By default, authentication is not enabled
	Enabled bool `json:"enabled"`

	// name of Secret containing Password based Authentication Parameters like username, password and acl
	// optional - used only by PasswordAuthHandler for authentication
	PasswordAuthSecret string `json:"passwordAuthSecret,omitempty"`
}

func (ap *AuthenticationParameters) IsEnabled() bool {
	if ap == nil {
		return false
	}
	return ap.Enabled
}

// ImageSpec defines the fields needed for a Docker repository image
type ImageSpec struct {
	Repository string `json:"repository"`

	// Deprecated: Use `spec.Version` instead
	Tag string `json:"tag,omitempty"`

	PullPolicy v1.PullPolicy `json:"pullPolicy"`
}

func (src *PravegaCluster) ConvertTo(dstRaw conversion.Hub) error {
	//do nothing here as we never want to move from v1beta1 to v1alpha1
	log.Print("ConvertTo invoked")
	return nil
}

func (dst *PravegaCluster) ConvertFrom(srcRaw conversion.Hub) error {
	log.Printf("ConvertFrom: invoked")
	//logic for conveting from v1alpha1 to v1beta1

	srcObj := srcRaw.(*v1alpha1.PravegaCluster)
	dst.ObjectMeta = srcObj.ObjectMeta

	log.Printf("DST: Name %s, Namespace %s, UID %v", dst.Name, dst.Namespace, dst.UID)
	log.Printf("SRC: Name %s, Namespace %s, UID %v", srcObj.Name, srcObj.Namespace, srcObj.UID)

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
		dst.Spec.Pravega.Tier2 = &Tier2Spec{
			FileSystem: &FileSystemSpec{
				PersistentVolumeClaim: srcObj.Spec.Pravega.Tier2.FileSystem.PersistentVolumeClaim,
			},
		}
	} else if srcObj.Spec.Pravega.Tier2.Ecs != nil {
		dst.Spec.Pravega.Tier2 = &Tier2Spec{
			Ecs: &ECSSpec{
				ConfigUri:   srcObj.Spec.Pravega.Tier2.Ecs.ConfigUri,
				Bucket:      srcObj.Spec.Pravega.Tier2.Ecs.Bucket,
				Prefix:      srcObj.Spec.Pravega.Tier2.Ecs.Prefix,
				Credentials: srcObj.Spec.Pravega.Tier2.Ecs.Credentials,
			},
		}
	} else if srcObj.Spec.Pravega.Tier2.Hdfs != nil {
		dst.Spec.Pravega.Tier2 = &Tier2Spec{
			Hdfs: &HDFSSpec{
				Uri:               srcObj.Spec.Pravega.Tier2.Hdfs.Uri,
				Root:              srcObj.Spec.Pravega.Tier2.Hdfs.Root,
				ReplicationFactor: srcObj.Spec.Pravega.Tier2.Hdfs.ReplicationFactor,
			},
		}
	}

	// Controller Resources
	if srcObj.Spec.Pravega.ControllerResources != nil {
		dst.Spec.Pravega.ControllerResources = &v1.ResourceRequirements{
			Requests: srcObj.Spec.Pravega.ControllerResources.Requests,
			Limits:   srcObj.Spec.Pravega.ControllerResources.Limits,
		}
	}
	// SegmentStore Resources
	if srcObj.Spec.Pravega.SegmentStoreResources != nil {
		dst.Spec.Pravega.SegmentStoreResources = &v1.ResourceRequirements{
			Requests: srcObj.Spec.Pravega.SegmentStoreResources.Requests,
			Limits:   srcObj.Spec.Pravega.SegmentStoreResources.Limits,
		}
	}

	return nil
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
	return p.validatePravegaVersion()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (p *PravegaCluster) ValidateUpdate(old runtime.Object) error {
	log.Printf("validate update %s", p.Name)
	return p.validatePravegaVersion()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (p *PravegaCluster) ValidateDelete() error {
	log.Printf("validate delete %s", p.Name)
	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func getSupportedVersions() (map[string]string, error) {
	var supportedVersions = map[string]string{}
	file, err := os.Open("/tmp/config/keys")

	if err != nil {
		log.Fatalf("failed opening file: %v", err)
		return supportedVersions, nil
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	defer file.Close()

	for _, eachline := range txtlines {
		entry := strings.Split(eachline, ":")
		supportedVersions[entry[0]] = entry[1]
	}
	return supportedVersions, nil
}

func (p *PravegaCluster) validatePravegaVersion() error {
	supportedVersions, err := getSupportedVersions()
	if err != nil {
		return fmt.Errorf("Error retrieving suported versions %v", err)
	}
	if p.Spec.Version == "" {
		p.Spec.Version = DefaultPravegaVersion
	}

	requestVersion := p.Spec.Version

	if p.Status.IsClusterInUpgradeFailedState() {
		if requestVersion != p.Status.GetLastVersion() {
			return fmt.Errorf("Rollback to version %s not supported. Only rollback to version %s is supported.", requestVersion, p.Status.GetLastVersion())
		}
		return nil
	}

	// Check if the request has a valid Pravega version
	normRequestVersion, err := util.NormalizeVersion(requestVersion)
	if err != nil {
		return fmt.Errorf("request version is not in valid format: %v", err)
	}
	if _, ok := supportedVersions[normRequestVersion]; !ok {
		return fmt.Errorf("unsupported Pravega cluster version %s", requestVersion)
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
	upgradeString, ok := supportedVersions[normFoundVersion]
	//upgradeList, ok := supportedVersions[normFoundVersion]
	if !ok {
		// It should never happen
		return fmt.Errorf("failed to find current cluster version in the supported versions")
	}
	upgradeList := strings.Split(upgradeString, ",")
	if !util.ContainsVersion(upgradeList, normRequestVersion) {
		return fmt.Errorf("unsupported upgrade from version %s to %s", p.Status.CurrentVersion, requestVersion)
	}
	return nil
}

//to return name of segmentstore based on the version
func (p *PravegaCluster) StatefulSetNameForSegmentstore() string {
	if util.IsVersionBelow07(p.Spec.Version) {
		return p.StatefulSetNameForSegmentstoreBelow07()
	}
	return p.StatefulSetNameForSegmentstoreAbove07()
}

//if version is above or equals to 0.7 this name will be assigned
func (p *PravegaCluster) StatefulSetNameForSegmentstoreAbove07() string {
	return fmt.Sprintf("%s-pravega-segment-store", p.ObjectMeta.Name)
}

//if version is below 0.7 this name will be assigned
func (p *PravegaCluster) StatefulSetNameForSegmentstoreBelow07() string {
	return fmt.Sprintf("%s-pravega-segmentstore", p.ObjectMeta.Name)
}

func (p *PravegaCluster) PravegaControllerServiceURL() string {
	return fmt.Sprintf("tcp://%v.%v:%v", p.ServiceNameForController(), p.ObjectMeta.Namespace, "9090")
}

func (p *PravegaCluster) LabelsForController() map[string]string {
	labels := p.LabelsForPravegaCluster()
	labels["component"] = "pravega-controller"
	return labels
}

func (p *PravegaCluster) LabelsForSegmentStore() map[string]string {
	labels := p.LabelsForPravegaCluster()
	labels["component"] = "pravega-segmentstore"
	return labels
}

func (pravegaCluster *PravegaCluster) LabelsForPravegaCluster() map[string]string {
	return map[string]string{
		"app":             "pravega-cluster",
		"pravega_cluster": pravegaCluster.Name,
	}
}

func (p *PravegaCluster) PdbNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.ObjectMeta.Name)
}

func (p *PravegaCluster) ConfigMapNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.ObjectMeta.Name)
}

func (p *PravegaCluster) ServiceNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.ObjectMeta.Name)
}

func (p *PravegaCluster) ServiceNameForSegmentStore(index int32) string {
	if util.IsVersionBelow07(p.Spec.Version) {
		return fmt.Sprintf("%s-pravega-segmentstore-%d", p.Name, index)
	}
	return fmt.Sprintf("%s-pravega-segment-store-%d", p.Name, index)
}

func (p *PravegaCluster) HeadlessServiceNameForSegmentStore() string {
	return fmt.Sprintf("%s-pravega-segmentstore-headless", p.ObjectMeta.Name)
}

func (p *PravegaCluster) HeadlessServiceNameForBookie() string {
	return fmt.Sprintf("%s-bookie-headless", p.ObjectMeta.Name)
}

func (p *PravegaCluster) DeploymentNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.ObjectMeta.Name)
}

func (p *PravegaCluster) PdbNameForSegmentstore() string {
	return fmt.Sprintf("%s-segmentstore", p.ObjectMeta.Name)
}

func (p *PravegaCluster) ConfigMapNameForSegmentstore() string {
	return fmt.Sprintf("%s-pravega-segmentstore", p.ObjectMeta.Name)
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

func (p *PravegaCluster) NewEvent(name string, reason string, message string, eventType string) *corev1.Event {
	now := metav1.Now()
	operatorName, _ := k8s.GetOperatorName()
	event := v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: p.Namespace,
			Labels:    p.LabelsForPravegaCluster(),
		},
		InvolvedObject: v1.ObjectReference{
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
