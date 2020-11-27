/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1beta2

import (
	"context"
	"fmt"
	"os"
	"time"

	k8s "github.com/operator-framework/operator-sdk/pkg/k8sutil"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (

	// DefaultPravegaVersion is the default tag used for for the Pravega
	// Docker image
	DefaultControllerVersion = "0.8.0"
	// DefaultControllerRequestCPU is the default CPU request for Pravega
	DefaultControllerRequestCPU = "250m"

	// DefaultControllerLimitCPU is the default CPU limit for Pravega
	DefaultControllerLimitCPU = "500m"

	// DefaultControllerRequestMemory is the default memory request for Pravega
	DefaultControllerRequestMemory = "512Mi"

	// DefaultControllerLimitMemory is the default memory limit for Pravega
	DefaultControllerLimitMemory = "1Gi"
	// DefaultPravegaImageRepository is the default Docker repository for
	// the Pravega image
	DefaultControllerImageRepository = "pravega/pravega"

	// DefaultPravegaImagePullPolicy is the default image pull policy used
	// for the Pravega Docker image
	DefaultControllerImagePullPolicy = v1.PullAlways
	// DefaultZookeeperUri is the default ZooKeeper URI in the form of "hostname:port"
)

func init() {
	SchemeBuilder.Register(&PravegaController{}, &PravegaControllerList{})
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PravegaControllerList contains a list of PravegaController
type PravegaControllerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PravegaController `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PravegaCluster is the Schema for the pravegaclusters API
// +k8s:openapi-gen=true
type PravegaController struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PravegaControllerSpec   `json:"spec,omitempty"`
	Status PravegaControllerStatus `json:"status,omitempty"`
}

// WithDefaults set default values when not defined in the spec.
func (pc *PravegaController) WithDefaults() (changed bool) {
	changed = pc.Spec.withDefaults()
	return changed
}

// ClusterSpec defines the desired state of PravegaCluster
type PravegaControllerSpec struct {
	// ZookeeperUri specifies the hostname/IP address and port in the format
	// "hostname:port".
	// By default, the value "zk-client:2181" is used, that corresponds to the
	// default Zookeeper service created by the Pravega Zookkeeper operator
	// available at: https://github.com/pravega/zookeeper-operator
	ZookeeperUri string `json:"zookeeperUri"`

	// ExternalAccess specifies whether or not to allow external access
	// to clients and the service type to use to achieve it
	// By default, external access is not enabled
	ExternalAccess *ControllerExternalAccess `json:"externalAccess"`

	// TLS is the Pravega security configuration that is passed to the Pravega processes.
	// See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md
	TLS *ControllerTLSPolicy `json:"tls,omitempty"`

	// Authentication can be enabled for authorizing all communication from clients to controller and segment store
	// See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md
	Authentication *ControllerAuthenticationParameters `json:"authentication,omitempty"`

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

	// ControllerReplicas defines the number of Controller replicas.
	// Defaults to 1.
	// +kubebuilder:validation:Minimum=0
	// +optional
	Replicas int32 `json:"replicas"`
	// DebugLogging indicates whether or not debug level logging is enabled.
	// Defaults to false.
	// +optional
	DebugLogging bool `json:"debugLogging"`

	// Image defines the Pravega Docker image to use.
	// By default, "pravega/pravega" will be used.
	// +optional
	Image *ControllerImageSpec `json:"image"`

	// Options is the Pravega configuration that is passed to the Pravega processes
	// as JAVA_OPTS. See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/config/config.properties
	// +optional
	Options map[string]string `json:"options"`

	// ControllerJvmOptions is the JVM options for controller. It will be passed to the JVM
	// for performance tuning. If this field is not specified, the operator will use a set of default
	// options that is good enough for general deployment.
	// +optional
	JvmOptions []string `json:"jvmOptions"`

	// ControllerServiceAccountName configures the service account used on controller instances.
	// If not specified, Kubernetes will automatically assign the default service account in the namespace
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Resources specifies the request and limit of resources that controller can have.
	// CrResources includes CPU and memory resources
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Type specifies the service type to achieve external access.
	// Options are "LoadBalancer" and "NodePort".
	// By default, if external access is enabled, it will use "LoadBalancer"
	ExternalServiceType v1.ServiceType `json:"extServiceType,omitempty"`

	// Annotations to be added to the external service
	// +optional
	ServiceAnnotations map[string]string `json:"svcAnnotations"`

	// ControllerSecurityContext holds security configuration that will be applied to a container
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
}

// ExternalAccess defines the configuration of the external access
type ControllerExternalAccess struct {
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

// ImageSpec defines the fields needed for a Docker repository image
type ControllerImageSpec struct {
	// +optional
	Repository string `json:"repository"`
	// +optional
	PullPolicy corev1.PullPolicy `json:"pullPolicy"`
}

type ControllerAuthenticationParameters struct {
	// Enabled specifies whether or not authentication is enabled
	// By default, authentication is not enabled
	// +optional
	Enabled bool `json:"enabled"`

	// name of Secret containing Password based Authentication Parameters like username, password and acl
	// optional - used only by PasswordAuthHandler for authentication
	PasswordAuthSecret string `json:"passwordAuthSecret,omitempty"`
}

type ControllerTLSPolicy struct {
	// Static TLS means keys/certs are generated by the user and passed to an operator.
	Static *ControllerStaticTLS `json:"static,omitempty"`
}

type ControllerStaticTLS struct {
	ControllerSecret string `json:"controllerStoreSecret,omitempty"`
	CaBundle         string `json:"caBundle,omitempty"`
}

func (cs *ControllerImageSpec) withDefaults() (changed bool) {
	if cs.Repository == "" {
		changed = true
		cs.Repository = DefaultControllerImageRepository
	}

	if cs.PullPolicy == "" {
		changed = true
		cs.PullPolicy = DefaultControllerImagePullPolicy
	}

	return changed
}

func (e *ControllerExternalAccess) withDefaults() (changed bool) {
	if e.Enabled == false && (e.Type != "" || e.DomainName != "") {
		changed = true
		e.Type = ""
		e.DomainName = ""
	}
	return changed
}

func (s *PravegaControllerSpec) withDefaults() (changed bool) {
	if s.ZookeeperUri == "" {
		changed = true
		s.ZookeeperUri = DefaultZookeeperUri
	}

	if s.ExternalAccess == nil {
		changed = true
		s.ExternalAccess = &ControllerExternalAccess{}
	}

	if s.ExternalAccess.withDefaults() {
		changed = true
	}

	if s.TLS == nil {
		changed = true
		s.TLS = &ControllerTLSPolicy{
			Static: &ControllerStaticTLS{},
		}
	}

	if s.Authentication == nil {
		changed = true
		s.Authentication = &ControllerAuthenticationParameters{}
	}

	if s.Version == "" {
		s.Version = DefaultControllerVersion
		changed = true
	}

	if s.BookkeeperUri == "" {
		s.BookkeeperUri = DefaultBookkeeperUri
		changed = true
	}
	if s.Image == nil {
		changed = true
		s.Image = &ControllerImageSpec{}
	}
	if s.Image.withDefaults() {
		changed = true
	}
	if s.Options == nil {
		changed = true
		s.Options = map[string]string{}
	}

	if s.JvmOptions == nil {
		changed = true
		s.JvmOptions = []string{}
	}

	if s.Resources == nil {
		changed = true
		s.Resources = &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(DefaultControllerRequestCPU),
				v1.ResourceMemory: resource.MustParse(DefaultControllerRequestMemory),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(DefaultControllerLimitCPU),
				v1.ResourceMemory: resource.MustParse(DefaultControllerLimitMemory),
			},
		}
	}
	if s.ServiceAnnotations == nil {
		changed = true
		s.ServiceAnnotations = map[string]string{}
	}
	return changed
}

func (ap *ControllerAuthenticationParameters) IsEnabled() bool {
	if ap == nil {
		return false
	}
	return ap.Enabled
}

func (tp *ControllerTLSPolicy) IsSecureController() bool {
	if tp == nil || tp.Static == nil {
		return false
	}
	return len(tp.Static.ControllerSecret) != 0
}

func (p *PravegaController) ConfigMapName() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
}

func (p *PravegaController) DeploymentNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
}

func (p *PravegaController) PdbNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
}

func (p *PravegaController) LabelsForPravegaController() map[string]string {
	labels := map[string]string{
		"app":                "pravega-controller",
		"pravega_controller": p.Name,
	}
	labels["component"] = "pravega-controller"
	return labels
}
func (p *PravegaController) LabelsForController() map[string]string {
	labels := map[string]string{
		"app":                "pravega-controller",
		"pravega_controller": p.Name,
	}
	labels["component"] = "pravega-controller"
	return labels
}

func (p *PravegaController) ControllerImage() (image string) {
	return fmt.Sprintf("%s:%s", p.Spec.Image.Repository, p.Spec.Version)
}

func (p *PravegaController) ServiceNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
}

func (p *PravegaController) PravegaControllerServiceURL() string {
	return fmt.Sprintf("tcp://%v.%v:%v", p.ServiceNameForController(), p.Namespace, "9090")
}

func (p *PravegaController) GetControllerClusterExpectedSize() (size int) {
	return int(p.Spec.Replicas)
}

func (p *PravegaController) NewApplicationEvent(name string, reason string, message string, eventType string) *corev1.Event {
	now := metav1.Now()
	operatorName, _ := k8s.GetOperatorName()
	generateName := name + "-"
	event := corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Namespace:    p.Namespace,
			Labels:       p.LabelsForPravegaController(),
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion: "app.k8s.io/v1beta2",
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

// Wait for pods in cluster to be terminated
func (p *PravegaController) WaitForClusterToTerminate(kubeClient client.Client) (err error) {
	listOptions := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(p.LabelsForPravegaController()),
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

func (p *PravegaController) NewEvent(name string, reason string, message string, eventType string) *corev1.Event {
	now := metav1.Now()
	operatorName, _ := k8s.GetOperatorName()
	generateName := name + "-"
	event := corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Namespace:    p.Namespace,
			Labels:       p.LabelsForPravegaController(),
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

func (p *PravegaController) PravegaControllerTargetImage() (string, error) {
	if p.Status.TargetVersion == "" {
		return "", fmt.Errorf("target version is not set")
	}
	return fmt.Sprintf("%s:%s", p.Spec.Image.Repository, p.Status.TargetVersion), nil
}
