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

package v1beta1

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/pravega/bookkeeper-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var pravegaclusterlog = logf.Log.WithName("pravegacluster-resource")

func (r *PravegaCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-pravegaclusters-pravega-pravega-io-v1beta1-pravegacluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=pravegaclusters.pravega.pravega.io,resources=pravegaclusters,verbs=create;update,versions=v1beta1,name=vpravegacluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &PravegaCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (p *PravegaCluster) ValidateCreate() error {
	pravegaclusterlog.Info("validate create", "name", p.Name)
	err := p.ValidatePravegaVersion()
	if err != nil {
		return err
	}
	err = p.ValidateSegmentStoreMemorySettings()
	if err != nil {
		return err
	}
	err = p.ValidateBookkeperSettings()
	if err != nil {
		return err
	}

	err = p.ValidateAuthenticationSettings()
	if err != nil {
		return err
	}
	return nil

}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (p *PravegaCluster) ValidateUpdate(old runtime.Object) error {
	pravegaclusterlog.Info("validate update", "name", p.Name)

	err := p.ValidatePravegaVersion()
	if err != nil {
		return err
	}
	err = p.validateConfigMap()
	if err != nil {
		return err
	}
	err = p.ValidateSegmentStoreMemorySettings()
	if err != nil {
		return err
	}
	err = p.ValidateBookkeperSettings()
	if err != nil {
		return err
	}
	err = p.ValidateAuthenticationSettings()
	if err != nil {
		return err
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (p *PravegaCluster) ValidateDelete() error {
	pravegaclusterlog.Info("validate delete", "name", p.Name)

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

// ValidateAuthenticationSettings checks for correct options passed to pravega
// when authentication is enabled/disabled.
func (p *PravegaCluster) ValidateAuthenticationSettings() error {
	if p.Spec.Authentication.Enabled == true {
		newkey, _ := p.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"]
		oldkey, _ := p.Spec.Pravega.Options["autoScale.authEnabled"]

		if newkey == "" && oldkey == "" {
			return fmt.Errorf("autoScale.controller.connect.security.auth.enable field is not present")
		}
		if newkey == "" {
			if oldkey != "true" {
				return fmt.Errorf("autoScale.authEnabled should be set to true")
			}
		}
		if oldkey == "" {
			if newkey != "true" {
				return fmt.Errorf("autoScale.controller.connect.security.auth.enable should be set to true")
			}
		}
		if oldkey != "" && newkey != "" {
			if newkey != "true" || oldkey != "true" {
				return fmt.Errorf("Both autoScale.controller.connect.security.auth.enable and autoScale.authEnabled should be set to true")
			}
		}
		signingkey1, ok := p.Spec.Pravega.Options["controller.security.auth.delegationToken.signingKey.basis"]
		if !ok {
			signingkey1, ok = p.Spec.Pravega.Options["controller.auth.tokenSigningKey"]
		}
		if !ok {
			return fmt.Errorf("controller.security.auth.delegationToken.signingKey.basis field is not present")
		}
		signingkey2, ok := p.Spec.Pravega.Options["autoScale.security.auth.token.signingKey.basis"]
		if !ok {
			signingkey2, ok = p.Spec.Pravega.Options["autoScale.tokenSigningKey"]
		}
		if !ok {
			return fmt.Errorf("autoScale.security.auth.token.signingKey.basis field is not present")
		}
		if signingkey1 != signingkey2 {
			return fmt.Errorf("controller and segmentstore token signing key should have same value")
		}
	} else {
		newkey, _ := p.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"]
		oldkey, _ := p.Spec.Pravega.Options["autoScale.authEnabled"]

		if oldkey == "true" || newkey == "true" {
			return fmt.Errorf("autoScale.controller.connect.security.auth.enable/autoScale.authEnabled should not be set to true")
		}
	}
	return nil
}

// ValidateSegmentStoreMemorySettings checks whether the user has passed the required values for segment store memory settings.
// The required values includes: SegmentStoreResources.limits.memory, pravegaservice.cache.size.max , -XX:MaxDirectMemorySize and -Xmx.
// Once the required values are set, the method also checks whether the following conditions are met:
// Total Memory > JVM Heap (-Xmx) + JVM Direct Memory (-XX:MaxDirectMemorySize)
// JVM Direct memory > Segment Store read cache size (pravegaservice.cache.size.max).
func (p *PravegaCluster) ValidateSegmentStoreMemorySettings() error {
	if p.Spec.Pravega.SegmentStoreResources == nil {
		return fmt.Errorf("spec.pravega.segmentStoreResources cannot be empty")
	}

	if p.Spec.Pravega.SegmentStoreResources.Limits == nil {
		return fmt.Errorf("spec.pravega.segmentStoreResources.limits cannot be empty")
	}

	if p.Spec.Pravega.SegmentStoreResources.Requests == nil {
		p.Spec.Pravega.SegmentStoreResources.Requests = map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    {},
			corev1.ResourceMemory: {},
		}
	}

	totalMemoryLimitsQuantity := p.Spec.Pravega.SegmentStoreResources.Limits[corev1.ResourceMemory]
	totalMemoryRequestsQuantity := p.Spec.Pravega.SegmentStoreResources.Requests[corev1.ResourceMemory]
	totalCpuLimitsQuantity := p.Spec.Pravega.SegmentStoreResources.Limits[corev1.ResourceCPU]
	totalCpuRequestsQuantity := p.Spec.Pravega.SegmentStoreResources.Requests[corev1.ResourceCPU]

	if totalMemoryLimitsQuantity == (resource.Quantity{}) {
		return fmt.Errorf("Missing required value for field spec.pravega.segmentStoreResources.limits.memory")
	}

	if totalCpuLimitsQuantity == (resource.Quantity{}) {
		return fmt.Errorf("Missing required value for field spec.pravega.segmentStoreResources.limits.cpu")
	}

	totalMemoryLimits := totalMemoryLimitsQuantity.Value()
	totalMemoryRequests := totalMemoryRequestsQuantity.Value()
	totalCpuLimits := totalCpuLimitsQuantity.Value()
	totalCpuRequests := totalCpuRequestsQuantity.Value()

	if totalMemoryLimits < totalMemoryRequests {
		return fmt.Errorf("spec.pravega.segmentStoreResources.requests.memory value must be less than or equal to spec.pravega.segmentStoreResources.limits.memory")
	}

	if totalCpuLimits < totalCpuRequests {
		return fmt.Errorf("spec.pravega.segmentStoreResources.requests.cpu value must be less than or equal to spec.pravega.segmentStoreResources.limits.cpu")
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

	if totalMemoryLimits <= (maxDirectMemorySize + xmx) {
		return fmt.Errorf("MaxDirectMemorySize(%v B) along with JVM Xmx value(%v B) should be less than the total available memory(%v B)!", maxDirectMemorySize, xmx, totalMemoryLimits)
	}

	if maxDirectMemorySize <= cacheSize {
		return fmt.Errorf("Cache size(%v B) configured should be less than the JVM MaxDirectMemorySize(%v B) value", cacheSize, maxDirectMemorySize)
	}

	return nil
}

// ValidateBookkeeperSettings checks that the value passed for the options bookkeeper.ensemble.size (E) bookkeeper.write.quorum.size (W)
// and bookkeeper.ack.quorum.size (A) adheres to the following rule, E >= W >= A.
// The method also checks for the option bookkeeper.write.quorum.racks.minimumCount.enable which should be set to false when bookkeeper.ensemble.size is 1.
// Note: The default value of E , W and A is 3.
func (p *PravegaCluster) ValidateBookkeperSettings() error {
	// Intializing ensemble size, write quorum size and ack quorum size to default value of 3
	ensembleSizeInt, writeQuorumSizeInt, ackQuorumSizeInt := 3, 3, 3
	var err error

	ensembleSize := p.Spec.Pravega.Options["bookkeeper.ensemble.size"]
	writeQuorumSize := p.Spec.Pravega.Options["bookkeeper.write.quorum.size"]
	ackQuorumSize := p.Spec.Pravega.Options["bookkeeper.ack.quorum.size"]
	writeQuorumRacks := p.Spec.Pravega.Options["bookkeeper.write.quorum.racks.minimumCount.enable"]

	if len(ensembleSize) > 0 {
		ensembleSizeInt, err = strconv.Atoi(ensembleSize)
		if err != nil {
			return fmt.Errorf("Cannot convert ensemble size from string to integer: %v", err)
		}
	}

	if len(writeQuorumSize) > 0 {
		writeQuorumSizeInt, err = strconv.Atoi(writeQuorumSize)
		if err != nil {
			return fmt.Errorf("Cannot convert write quorum size from string to integer: %v", err)
		}
	}

	if len(ackQuorumSize) > 0 {
		ackQuorumSizeInt, err = strconv.Atoi(ackQuorumSize)
		if err != nil {
			return fmt.Errorf("Cannot convert ack quorum size from string to integer: %v", err)
		}
	}

	if writeQuorumRacks != "true" && writeQuorumRacks != "false" && writeQuorumRacks != "" {
		return fmt.Errorf("bookkeeper.write.quorum.racks.minimumCount.enable can be only set to \"true\" \"false\" or \"\"")
	}

	if writeQuorumRacks == "true" && ensembleSizeInt == 1 {
		return fmt.Errorf("bookkeeper.write.quorum.racks.minimumCount.enable should be set to false if bookkeeper.ensemble.size is 1")
	}

	if ensembleSizeInt < writeQuorumSizeInt {
		if ensembleSize == "" {
			return fmt.Errorf("The value provided for the option bookkeeper.write.quorum.size should be less than or equal to the value of option bookkeeper.ensemble.size (default is 3)")
		}
		if writeQuorumSize == "" {
			return fmt.Errorf("The value provided for the option bookkeeper.ensemble.size should be greater than or equal to the value of option bookkeeper.write.quorum.size (default is 3)")
		}
		return fmt.Errorf("The value provided for the option bookkeeper.write.quorum.size should be less than or equal to the value of option bookkeeper.ensemble.size")
	}

	if writeQuorumSizeInt < ackQuorumSizeInt {
		if writeQuorumSize == "" {
			return fmt.Errorf("The value provided for the option bookkeeper.ack.quorum.size should be less than or equal to the value of option bookkeeper.write.quorum.size (default is 3)")
		}
		if ackQuorumSize == "" {
			return fmt.Errorf("The value provided for the option bookkeeper.write.quorum.size should be greater than or equal to the value of option bookkeeper.ack.quorum.size (default is 3)")
		}
		return fmt.Errorf("The value provided for the option bookkeeper.ack.quorum.size should be less than or equal to the value of option bookkeeper.write.quorum.size")
	}

	return nil
}
