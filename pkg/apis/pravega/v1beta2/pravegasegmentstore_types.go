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
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	k8s "github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/pravega/pravega-operator/pkg/controller/config"
	"github.com/pravega/pravega-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var Mgr manager.Manager

const (
	// DefaultPravegaImageRepository is the default Docker repository for
	// the Pravega image
	DefaultPravegaImageRepository = "pravega/pravega"

	// DefaultPravegaImagePullPolicy is the default image pull policy used
	// for the Pravega Docker image
	DefaultPravegaImagePullPolicy = v1.PullAlways
	// DefaultZookeeperUri is the default ZooKeeper URI in the form of "hostname:port"
	DefaultZookeeperUri = "zookeeper-client:2181"

	// DefaultBookkeeperUri is the default ZooKeeper URI in the form of "hostname:port"
	DefaultBookkeeperUri = "bookkeeper-bookie-0.bookkeeper-bookie-headless.default.svc.cluster.local:3181,bookkeeper-bookie-1.bookkeeper-bookie-headless.default.svc.cluster.local:3181,bookkeeper-bookie-2.bookkeeper-bookie-headless.default.svc.cluster.local:3181"

	// DefaultServiceType is the default service type for external access
	DefaultServiceType = corev1.ServiceTypeLoadBalancer

	// DefaultPravegaVersion is the default tag used for for the Pravega
	// Docker image
	DefaultSegmentStoreVersion = "0.7.0"
	// DefaultSegmentStoreRequestCPU is the default CPU request for Pravega
	DefaultSegmentStoreRequestCPU = "500m"

	// DefaultSegmentStoreLimitCPU is the default CPU limit for Pravega
	DefaultSegmentStoreLimitCPU = "1"

	// DefaultSegmentStoreRequestMemory is the default memory request for Pravega
	DefaultSegmentStoreRequestMemory = "1Gi"

	// DefaultSegmentStoreLimitMemory is the default memory limit for Pravega
	DefaultSegmentStoreLimitMemory = "2Gi"
	// DefaultPravegaLTSClaimName is the default volume claim name used as Tier 2
	DefaultPravegaLTSClaimName = "pravega-tier2"
)

func init() {
	SchemeBuilder.Register(&PravegaSegmentStore{}, &PravegaSegmentStoreList{})
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PravegaSegmentStoreList contains a list of PravegaSegmentStore
type PravegaSegmentStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PravegaSegmentStore `json:"items"`
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
// PravegaSegmentStore is the Schema for the PravegaSegmentStores API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type PravegaSegmentStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PravegaSegmentStoreSpec   `json:"spec,omitempty"`
	Status PravegaSegmentStoreStatus `json:"status,omitempty"`
}

// WithDefaults set default values when not defined in the spec.
func (p *PravegaSegmentStore) WithDefaults() (changed bool) {
	changed = p.Spec.withDefaults()
	return changed
}

// ClusterSpec defines the desired state of PravegaSegmentStore
type PravegaSegmentStoreSpec struct {
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
	ExternalAccess *SegmentStoreExternalAccess `json:"externalAccess"`

	// TLS is the Pravega security configuration that is passed to the Pravega processes.
	// See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md
	// +optional
	TLS *SegmentStoreTLSPolicy `json:"tls,omitempty"`

	// Authentication can be enabled for authorizing all communication from clients to controller and segment store
	// See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md
	Authentication *SegmentStoreAuthenticationParameters `json:"authentication,omitempty"`

	// Version is the expected version of the Pravega cluster.
	// The pravega-operator will eventually make the Pravega cluster version
	// equal to the expected version.
	//
	// The version must follow the [semver]( http://semver.org) format, for example "3.2.13".
	// Only Pravega released versions are supported: https://github.com/pravega/pravega/releases
	//
	// If version is not set, default is "0.4.0".
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

	// SegmentStoreReplicas defines the number of Segment Store replicas.
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
	Image *SegmentStoreImageSpec `json:"image"`

	// Options is the Pravega configuration that is passed to the Pravega processes
	// as JAVA_OPTS. See the following file for a complete list of options:
	// https://github.com/pravega/pravega/blob/master/config/config.properties
	// +optional
	Options map[string]string `json:"options"`

	// SegmentStoreJVMOptions is the JVM options for Segmentstore. It will be passed to the JVM
	// for performance tuning. If this field is not specified, the operator will use a set of default
	// options that is good enough for general deployment.
	// +optional
	JvmOptions []string `json:"jvmOptions"`

	// CacheVolumeClaimTemplate is the spec to describe PVC for the Pravega cache.
	// This field is optional. If no PVC spec, stateful containers will use
	// emptyDir as volume
	// +optional
	CacheVolumeClaimTemplate *v1.PersistentVolumeClaimSpec `json:"cacheVolumeClaimTemplate,omitempty"`

	// LongTermStorage is the configuration of Pravega's tier 2 storage. If no configuration
	// is provided, it will assume that a PersistentVolumeClaim called "pravega-longterm"
	// is present and it will use it as Tier 2
	// +optional
	LongTermStorage *LongTermStorageSpec `json:"longtermStorage"`

	// SegmentStoreServiceAccountName configures the service account used on segment store instances.
	// If not specified, Kubernetes will automatically assign the default service account in the namespace
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// SegmentStoreResources specifies the request and limit of resources that segmentStore can have.
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// Provides the name of the configmap created by the user to provide additional key-value pairs
	// that need to be configured into the ss pod as environmental variables
	EnvVars string `json:"envVars,omitempty"`

	// SegmentStoreSecret specifies whether or not any secret needs to be configured into the ss pod
	// either as an environment variable or by mounting it to a volume
	// +optional
	Secret *Secret `json:"Secret"`

	// Type specifies the service type to achieve external access.
	// Options are "LoadBalancer" and "NodePort".
	// By default, if external access is enabled, it will use "LoadBalancer"
	ExternalServiceType v1.ServiceType `json:"extServiceType,omitempty"`

	// Annotations to be added to the external service
	// +optional
	ServiceAnnotations map[string]string `json:"svcAnnotations"`

	// Specifying this IP would ensure we use same IP address for all the ss services
	LoadBalancerIP string `json:"loadBalancerIP,omitempty"`

	// SegmentStoreExternalTrafficPolicy defines the ExternalTrafficPolicy it can have cluster or local
	ExternalTrafficPolicy string `json:"externalTrafficPolicy,omitempty"`

	// SegmentStoreSecurityContext holds security configuration that will be applied to a container
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
}

func (s *PravegaSegmentStoreSpec) withDefaults() (changed bool) {
	if s.ZookeeperUri == "" {
		changed = true
		s.ZookeeperUri = DefaultZookeeperUri
	}

	if s.ExternalAccess == nil {
		changed = true
		s.ExternalAccess = &SegmentStoreExternalAccess{}
	}

	if s.ExternalAccess.withDefaults() {
		changed = true
	}

	if s.TLS == nil {
		changed = true
		s.TLS = &SegmentStoreTLSPolicy{
			Static: &SegmentStoreStaticTLS{},
		}
	}

	if s.Authentication == nil {
		changed = true
		s.Authentication = &SegmentStoreAuthenticationParameters{}
	}

	if s.Version == "" {
		s.Version = DefaultSegmentStoreVersion
		changed = true
	}

	if s.BookkeeperUri == "" {
		s.BookkeeperUri = DefaultBookkeeperUri
		changed = true
	}

	if !config.TestMode && s.Replicas < 1 {
		changed = true
		s.Replicas = 1
	}

	if s.Image == nil {
		changed = true
		s.Image = &SegmentStoreImageSpec{}
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

	if s.LongTermStorage == nil {
		changed = true
		s.LongTermStorage = &LongTermStorageSpec{}
	}

	if s.LongTermStorage.withDefaults() {
		changed = true
	}

	if s.Resources == nil {
		changed = true
		s.Resources = &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(DefaultSegmentStoreRequestCPU),
				v1.ResourceMemory: resource.MustParse(DefaultSegmentStoreRequestMemory),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(DefaultSegmentStoreLimitCPU),
				v1.ResourceMemory: resource.MustParse(DefaultSegmentStoreLimitMemory),
			},
		}
	}

	if s.Secret == nil {
		changed = true
		s.Secret = &Secret{}
	}

	if s.Secret.withDefaults() {
		changed = true
	}

	if s.ServiceAnnotations == nil {
		changed = true
		s.ServiceAnnotations = map[string]string{}
	}

	return changed
}

// ExternalAccess defines the configuration of the external access
type SegmentStoreExternalAccess struct {
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

func (e *SegmentStoreExternalAccess) withDefaults() (changed bool) {
	if e.Enabled == false && (e.Type != "" || e.DomainName != "") {
		changed = true
		e.Type = ""
		e.DomainName = ""
	}
	return changed
}

type SegmentStoreTLSPolicy struct {
	// Static TLS means keys/certs are generated by the user and passed to an operator.
	Static *SegmentStoreStaticTLS `json:"static,omitempty"`
}

type SegmentStoreStaticTLS struct {
	SegmentStoreSecret string `json:"segmentStoreSecret,omitempty"`
	CaBundle           string `json:"caBundle,omitempty"`
}

func (tp *SegmentStoreTLSPolicy) IsSecureSegmentStore() bool {
	if tp == nil || tp.Static == nil {
		return false
	}
	return len(tp.Static.SegmentStoreSecret) != 0
}

func (tp *SegmentStoreTLSPolicy) IsCaBundlePresent() bool {
	if tp == nil || tp.Static == nil {
		return false
	}
	return len(tp.Static.CaBundle) != 0
}

type SegmentStoreAuthenticationParameters struct {
	// Enabled specifies whether or not authentication is enabled
	// By default, authentication is not enabled
	// +optional
	Enabled bool `json:"enabled"`

	// name of Secret containing Password based Authentication Parameters like username, password and acl
	// optional - used only by PasswordAuthHandler for authentication
	PasswordAuthSecret string `json:"passwordAuthSecret,omitempty"`
}

func (ap *SegmentStoreAuthenticationParameters) IsEnabled() bool {
	if ap == nil {
		return false
	}
	return ap.Enabled
}

// ImageSpec defines the fields needed for a Docker repository image
type SegmentStoreImageSpec struct {
	// +optional
	Repository string `json:"repository"`
	// +optional
	PullPolicy corev1.PullPolicy `json:"pullPolicy"`
}

// LongTermStorageSpec configures the Tier 2 storage type to use with Pravega.
// If not specified, Tier 2 will be configured in filesystem mode and will try
// to use a PersistentVolumeClaim with the name "pravega-longterm"
type LongTermStorageSpec struct {
	// FileSystem is used to configure a pre-created Persistent Volume Claim
	// as Tier 2 backend.
	// It is default Tier 2 mode.
	FileSystem *FileSystemSpec `json:"filesystem,omitempty"`

	// Ecs is used to configure a Dell EMC ECS system as a Tier 2 backend
	Ecs *ECSSpec `json:"ecs,omitempty"`

	// Hdfs is used to configure an HDFS system as a Tier 2 backend
	Hdfs *HDFSSpec `json:"hdfs,omitempty"`
}

func (s *LongTermStorageSpec) withDefaults() (changed bool) {
	if s.FileSystem == nil && s.Ecs == nil && s.Hdfs == nil {
		changed = true
		fs := &FileSystemSpec{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: DefaultPravegaLTSClaimName,
			},
		}
		s.FileSystem = fs
	}

	return changed
}

// FileSystemSpec contains the reference to a PVC.
type FileSystemSpec struct {
	// +optional
	PersistentVolumeClaim *v1.PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim"`
}

// ECSSpec contains the connection details to a Dell EMC ECS system
type ECSSpec struct {
	// +optional
	ConfigUri string `json:"configUri"`
	// +optional
	Bucket string `json:"bucket"`
	// +optional
	Prefix string `json:"prefix"`
	// +optional
	Credentials string `json:"credentials"`
}

// HDFSSpec contains the connection details to an HDFS system
type HDFSSpec struct {
	// +optional
	Uri string `json:"uri"`
	// +optional
	Root string `json:"root"`
	// +optional
	ReplicationFactor int32 `json:"replicationFactor"`
}

// SegmentStoreSecret defines the configuration of the secret for the Segment Store
type Secret struct {
	// Secret specifies the name of Secret which needs to be configured
	// +optional
	Secret string `json:"secret"`

	// Path to the volume where the secret will be mounted
	// This value is considered only when the secret is provided
	// If this value is provided, the secret is mounted to a Volume
	// else the secret is exposed as an Environment Variable
	// +optional
	MountPath string `json:"mountPath"`
}

func (s *Secret) withDefaults() (changed bool) {
	if s.Secret == "" {
		s.MountPath = ""
	}

	return changed
}
func (s *SegmentStoreImageSpec) withDefaults() (changed bool) {
	if s.Repository == "" {
		changed = true
		s.Repository = DefaultPravegaImageRepository
	}

	if s.PullPolicy == "" {
		changed = true
		s.PullPolicy = DefaultPravegaImagePullPolicy
	}

	return changed
}
func (src *PravegaSegmentStore) ConvertTo(dstRaw conversion.Hub) error {
	//do nothing here as we never want to move from v1beta2 to v1alpha1
	return nil
}

func (dst *PravegaSegmentStore) ConvertFrom(srcRaw conversion.Hub) error {
	log.Printf("Converting Pravega CR version from v1alpha1 to v1beta2.")
	//logic for conveting from v1alpha1 to v1beta2
	/*	srcObj := srcRaw.(*v1alpha1.PravegaSegmentStore)
		dst.ObjectMeta = srcObj.ObjectMeta
		err := dst.convertSpec(srcObj)
		if err != nil {
			log.Fatalf("Error converting CR object from version v1alpha1 to v1beta2 %v", err)
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
		log.Print("Version migration completed successfully.")*/
	return nil
}

/*func getBookkeeperUri(srcObj *v1alpha1.PravegaSegmentStore) string {
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
}*/

var _ webhook.Validator = &PravegaSegmentStore{}

func (p *PravegaSegmentStore) SetupWebhookWithManager(mgr ctrl.Manager) error {
	log.Print("Registering Webhook")
	return ctrl.NewWebhookManagedBy(mgr).
		For(&PravegaSegmentStore{}).
		Complete()
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (p *PravegaSegmentStore) ValidateCreate() error {
	log.Printf("validate create %s", p.Name)
	return p.ValidatePravegaVersion("")
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (p *PravegaSegmentStore) ValidateUpdate(old runtime.Object) error {
	log.Printf("validate update %s", p.Name)
	err := p.ValidatePravegaVersion("")
	if err != nil {
		return err
	}
	/*	err = p.validateConfigMap()
		if err != nil {
			return err
		}*/
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (p *PravegaSegmentStore) ValidateDelete() error {
	log.Printf("validate delete %s", p.Name)
	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func getSupportedVersions(filename string) (map[string]string, error) {
	var supportedVersions = map[string]string{}
	filepath := filename
	if filename == "" {
		filepath = "/tmp/config/keys"
	}

	file, err := os.Open(filepath)

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

func (p *PravegaSegmentStore) ValidatePravegaVersion(filename string) error {
	supportedVersions, err := getSupportedVersions(filename)
	if err != nil {
		return fmt.Errorf("Error retrieving suported versions %v", err)
	}

	if p.Spec.Version == "" {
		p.Spec.Version = DefaultSegmentStoreVersion
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

	log.Printf("ValidatePravegaVersion:: normFoundVersion %s", normFoundVersion)
	upgradeString, ok := supportedVersions[normFoundVersion]
	if !ok {
		// It should never happen
		return fmt.Errorf("failed to find current cluster version in the supported versions")
	}
	upgradeList := strings.Split(upgradeString, ",")
	if !util.ContainsVersion(upgradeList, normRequestVersion) {
		return fmt.Errorf("unsupported upgrade from version %s to %s", p.Status.CurrentVersion, requestVersion)
	}
	log.Print("ValidatePravegaVersion:: No error found...returning...")
	return nil
}

/*func (p *PravegaSegmentStore) validateConfigMap() error {
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
	if val, ok := p.Spec.Pravega.Options["controller.containerCount"]; ok {
		checkstring := fmt.Sprintf("-Dcontroller.containerCount=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("controller.containerCount should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["controller.container.count"]; ok {
		checkstring := fmt.Sprintf("-Dcontroller.container.count=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("controller.container.count should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["pravegaservice.containerCount"]; ok {
		checkstring := fmt.Sprintf("-Dpravegaservice.containerCount=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("pravegaservice.containerCount should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["pravegaservice.container.count"]; ok {
		checkstring := fmt.Sprintf("-Dpravegaservice.container.count=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("pravegaservice.container.count should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["bookkeeper.bkLedgerPath"]; ok {
		checkstring := fmt.Sprintf("-Dbookkeeper.bkLedgerPath=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("bookkeeper.bkLedgerPath should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["bookkeeper.ledger.path"]; ok {
		checkstring := fmt.Sprintf("-Dbookkeeper.ledger.path=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("bookkeeper.ledger.path should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["controller.retention.bucketCount"]; ok {
		checkstring := fmt.Sprintf("-Dcontroller.retention.bucketCount=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("controller.retention.bucketCount should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["controller.retention.bucket.count"]; ok {
		checkstring := fmt.Sprintf("-Dcontroller.retention.bucket.count=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("controller.retention.bucket.count should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["controller.watermarking.bucketCount"]; ok {
		checkstring := fmt.Sprintf("-Dcontroller.watermarking.bucketCount=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("controller.watermarking.bucketCount should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["controller.watermarking.bucket.count"]; ok {
		checkstring := fmt.Sprintf("-Dcontroller.watermarking.bucket.count=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("controller.watermarking.bucket.count should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["pravegaservice.dataLogImplementation"]; ok {
		checkstring := fmt.Sprintf("-Dpravegaservice.dataLogImplementation=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("pravegaservice.dataLogImplementation should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["pravegaservice.dataLog.impl.name"]; ok {
		checkstring := fmt.Sprintf("-Dpravegaservice.dataLog.impl.name=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("pravegaservice.dataLog.impl.name should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["pravegaservice.storageImplementation"]; ok {
		checkstring := fmt.Sprintf("-Dpravegaservice.storageImplementation=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("pravegaservice.storageImplementation should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["pravegaservice.storage.impl.name"]; ok {
		checkstring := fmt.Sprintf("-Dpravegaservice.storage.impl.name=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("pravegaservice.storage.impl.name should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["storageextra.storageNoOpMode"]; ok {
		checkstring := fmt.Sprintf("-Dstorageextra.storageNoOpMode=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("storageextra.storageNoOpMode should not be changed ")
		}
	}
	if val, ok := p.Spec.Pravega.Options["storageextra.noOp.mode.enable"]; ok {
		checkstring := fmt.Sprintf("-Dstorageextra.noOp.mode.enable=%v", val)
		eq := strings.Contains(configmap.Data["JAVA_OPTS"], checkstring)
		if !eq {
			return fmt.Errorf("storageextra.noOp.mode.enable should not be changed ")
		}
	}
	log.Print("validateConfigMap:: No error found...returning...")
	return nil
}*/

//to return name of segmentstore based on the version
func (p *PravegaSegmentStore) StatefulSetNameForSegmentstore() string {
	if util.IsVersionBelow07(p.Spec.Version) {
		return p.StatefulSetNameForSegmentstoreBelow07()
	}
	return p.StatefulSetNameForSegmentstoreAbove07()
}

//if version is above or equals to 0.7 this name will be assigned
func (p *PravegaSegmentStore) StatefulSetNameForSegmentstoreAbove07() string {
	return fmt.Sprintf("%s-pravega-segment-store", p.Name)
}

//if version is below 0.7 this name will be assigned
func (p *PravegaSegmentStore) StatefulSetNameForSegmentstoreBelow07() string {
	return fmt.Sprintf("%s-pravega-segmentstore", p.Name)
}

func (p *PravegaSegmentStore) LabelsForController() map[string]string {
	labels := p.LabelsForPravegaSegmentStore()
	labels["component"] = "pravega-controller"
	return labels
}

func (p *PravegaSegmentStore) LabelsForSegmentStore() map[string]string {
	labels := p.LabelsForPravegaSegmentStore()
	labels["component"] = "pravega-segmentstore"
	return labels
}

func (PravegaSegmentStore *PravegaSegmentStore) LabelsForPravegaSegmentStore() map[string]string {
	return map[string]string{
		"app":             "pravega-cluster",
		"pravega_cluster": PravegaSegmentStore.Name,
	}
}

func (p *PravegaSegmentStore) PdbNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
}

func (p *PravegaSegmentStore) ConfigMapNameForController() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
}

func (p *PravegaSegmentStore) ServiceNameForSegmentStore(index int32) string {
	if util.IsVersionBelow07(p.Spec.Version) {
		return p.ServiceNameForSegmentStoreBelow07(index)
	}
	return p.ServiceNameForSegmentStoreAbove07(index)
}

func (p *PravegaSegmentStore) ServiceNameForSegmentStoreBelow07(index int32) string {
	return fmt.Sprintf("%s-pravega-segmentstore-%d", p.Name, index)
}

func (p *PravegaSegmentStore) ServiceNameForSegmentStoreAbove07(index int32) string {
	return fmt.Sprintf("%s-pravega-segment-store-%d", p.Name, index)
}

func (p *PravegaSegmentStore) HeadlessServiceNameForSegmentStore() string {
	return fmt.Sprintf("%s-pravega-segmentstore-headless", p.Name)
}

func (p *PravegaSegmentStore) HeadlessServiceNameForBookie() string {
	return fmt.Sprintf("%s-bookie-headless", p.Name)
}

func (p *PravegaSegmentStore) PdbNameForSegmentstore() string {
	return fmt.Sprintf("%s-segmentstore", p.Name)
}

func (p *PravegaSegmentStore) ConfigMapNameForSegmentstore() string {
	return fmt.Sprintf("%s-pravega-segmentstore", p.Name)
}

func (p *PravegaSegmentStore) PravegaTargetImage() (string, error) {
	if p.Status.TargetVersion == "" {
		return "", fmt.Errorf("target version is not set")
	}
	return fmt.Sprintf("%s:%s", p.Spec.Image.Repository, p.Status.TargetVersion), nil
}

// Wait for pods in cluster to be terminated
func (p *PravegaSegmentStore) WaitForClusterToTerminate(kubeClient client.Client) (err error) {
	listOptions := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(p.LabelsForPravegaSegmentStore()),
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

func (p *PravegaSegmentStore) NewEvent(name string, reason string, message string, eventType string) *corev1.Event {
	now := metav1.Now()
	operatorName, _ := k8s.GetOperatorName()
	generateName := name + "-"
	event := corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Namespace:    p.Namespace,
			Labels:       p.LabelsForPravegaSegmentStore(),
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion:      p.APIVersion,
			Kind:            "PravegaSegmentStore",
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

func (p *PravegaSegmentStore) NewApplicationEvent(name string, reason string, message string, eventType string) *corev1.Event {
	now := metav1.Now()
	operatorName, _ := k8s.GetOperatorName()
	generateName := name + "-"
	event := corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Namespace:    p.Namespace,
			Labels:       p.LabelsForPravegaSegmentStore(),
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

func (p *PravegaSegmentStore) PravegaImage() (image string) {
	return fmt.Sprintf("%s:%s", p.Spec.Image.Repository, p.Spec.Version)
}

func (p *PravegaSegmentStore) ServiceNameForControllerForSegmentStore() string {
	return fmt.Sprintf("%s-pravega-controller", p.Name)
}

func (p *PravegaSegmentStore) PravegaControllerServiceURLForSegmentStore() string {
	return fmt.Sprintf("tcp://%v.%v:%v", p.ServiceNameForControllerForSegmentStore(), p.Namespace, "9090")
}
