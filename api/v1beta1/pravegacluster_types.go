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
	"time"

	"github.com/pravega/pravega-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
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
	DefaultPravegaVersion = "0.11.0"

	// OperatorNameEnvVar is env variable for operator name
	OperatorNameEnvVar = "OPERATOR_NAME"
)

func init() {
	SchemeBuilder.Register(&PravegaCluster{}, &PravegaClusterList{})
}

// +kubebuilder:object:root=true

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

func (p *PravegaCluster) PdbNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
}

func (p *PravegaCluster) ConfigMapNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
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
	operatorName, _ := OperatorName()
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
	operatorName, _ := OperatorName()
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

// OperatorName returns the operator name
func OperatorName() (string, error) {
	operatorName, found := os.LookupEnv(OperatorNameEnvVar)
	if !found {
		return "", fmt.Errorf("environment variable %s is not set", OperatorNameEnvVar)
	}
	if len(operatorName) == 0 {
		return "", fmt.Errorf("environment variable %s is empty", OperatorNameEnvVar)
	}
	return operatorName, nil
}
