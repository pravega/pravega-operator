/**
 * Copyright (c) 2019 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */
package e2e

import (
	"testing"

	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func testWebhook(t *testing.T) {
	g := NewGomegaWithT(t)

	doCleanup := true
	ctx := framework.NewTestCtx(t)
	defer func() {
		if doCleanup {
			ctx.Cleanup()
		}
	}()

	namespace, err := ctx.GetNamespace()
	g.Expect(err).NotTo(HaveOccurred())
	f := framework.Global

	//creating the setup for running the test
	err = pravega_e2eutil.InitialSetup(t, f, ctx, namespace)
	g.Expect(err).NotTo(HaveOccurred())

	//Test webhook with an invalid Pravega cluster version format
	invalidVersion := pravega_e2eutil.NewClusterWithVersion(namespace, "999")
	invalidVersion.WithDefaults()
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, invalidVersion)
	g.Expect(err).To(HaveOccurred(), "Should reject deployment of invalid version format")
	g.Expect(err.Error()).To(ContainSubstring("request version is not in valid format:"))

	// Test webhook with with no segementStoreResources object
	noSegmentStoreResource := pravega_e2eutil.NewDefaultCluster(namespace)
	noSegmentStoreResource.WithDefaults()
	noSegmentStoreResource.Spec.Pravega.SegmentStoreResources = nil
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, noSegmentStoreResource)
	g.Expect(err).To(HaveOccurred(), "Spec.Pravega.SegmentStoreResources cannot be empty")
	g.Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources cannot be empty"))

	// Test webhook with with no segementStoreResources.Limits object
	noSegmentStoreResourceLimits := pravega_e2eutil.NewDefaultCluster(namespace)
	noSegmentStoreResourceLimits.WithDefaults()
	noSegmentStoreResourceLimits.Spec.Pravega.SegmentStoreResources.Limits = nil
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, noSegmentStoreResourceLimits)
	g.Expect(err).To(HaveOccurred(), "Spec.Pravega.SegmentStoreResources.Limits cannot be empty")
	g.Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources.limits cannot be empty"))

	// Test webhook with with no segementStoreResources.Requests object
	noSegmentStoreResourceRequests := pravega_e2eutil.NewDefaultCluster(namespace)
	noSegmentStoreResourceRequests.WithDefaults()
	noSegmentStoreResourceRequests.Spec.Pravega.SegmentStoreResources.Requests = nil
	pravega, err := pravega_e2eutil.CreatePravegaCluster(t, f, ctx, noSegmentStoreResourceRequests)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.DeletePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WaitForPravegaClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	//creating the setup for running the test
	err = pravega_e2eutil.InitialSetup(t, f, ctx, namespace)
	g.Expect(err).NotTo(HaveOccurred())

	// Test webhook with no value for segment store memory limits
	noMemoryLimits := pravega_e2eutil.NewClusterWithNoSegmentStoreMemoryLimits(namespace)
	noMemoryLimits.WithDefaults()
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, noMemoryLimits)
	g.Expect(err).To(HaveOccurred(), "Segment Store memory limits cannot be empty")
	g.Expect(err.Error()).To(ContainSubstring("Missing required value for field spec.pravega.segmentStoreResources.limits.memory"))

	// Test webhook with no value for segment store cpu limits
	noCpuLimits := pravega_e2eutil.NewClusterWithNoSegmentStoreCpuLimits(namespace)
	noCpuLimits.WithDefaults()
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, noCpuLimits)
	g.Expect(err).To(HaveOccurred(), "Segment Store cpu limits cannot be empty")
	g.Expect(err.Error()).To(ContainSubstring("Missing required value for field spec.pravega.segmentStoreResources.limits.cpu"))

	// Test webhook with segment store memory requests being greater than memory limits
	memoryRequestsGreaterThanLimits := pravega_e2eutil.NewDefaultCluster(namespace)
	memoryRequestsGreaterThanLimits.WithDefaults()
	memoryRequestsGreaterThanLimits.Spec.Pravega.SegmentStoreResources.Requests[corev1.ResourceMemory] = resource.MustParse("5Gi")
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, memoryRequestsGreaterThanLimits)
	g.Expect(err).To(HaveOccurred(), "Segment Store memory requests should be less than or equal to limits")
	g.Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources.requests.memory value must be less than or equal to spec.pravega.segmentStoreResources.limits.memory"))

	// Test webhook with segment store cpu requests being greater than cpu limits
	cpuRequestsGreaterThanLimits := pravega_e2eutil.NewDefaultCluster(namespace)
	cpuRequestsGreaterThanLimits.WithDefaults()
	cpuRequestsGreaterThanLimits.Spec.Pravega.SegmentStoreResources.Requests[corev1.ResourceCPU] = resource.MustParse("3000m")
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, cpuRequestsGreaterThanLimits)
	g.Expect(err).To(HaveOccurred(), "Segment Store cpu requests should be less than or equal to limits")
	g.Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources.requests.cpu value must be less than or equal to spec.pravega.segmentStoreResources.limits.cpu"))

	// Test webhook with no value for option pravegaservice.cache.size.max
	cacheSizeMax := pravega_e2eutil.NewDefaultCluster(namespace)
	cacheSizeMax.WithDefaults()
	cacheSizeMax.Spec.Pravega.Options["pravegaservice.cache.size.max"] = ""
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, cacheSizeMax)
	g.Expect(err).To(HaveOccurred(), "pravegaservice.cache.size.max cannot be empty")
	g.Expect(err.Error()).To(ContainSubstring("Missing required value for option pravegaservice.cache.size.max"))

	// Test Webhook with no value for JVM option -Xmx
	noXmx := pravega_e2eutil.NewDefaultCluster(namespace)
	noXmx.WithDefaults()
	noXmx.Spec.Pravega.SegmentStoreJVMOptions = []string{"-XX:MaxDirectMemorySize=2560m"}
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, noXmx)
	g.Expect(err).To(HaveOccurred(), "JVM option -Xmx cannot be empty")
	g.Expect(err.Error()).To(ContainSubstring("Missing required value for Segment Store JVM Option -Xmx"))

	// Test Webhook with no value for JVM option -XX:MaxDirectMemorySize
	noMaxDirectMemorySize := pravega_e2eutil.NewDefaultCluster(namespace)
	noMaxDirectMemorySize.WithDefaults()
	noMaxDirectMemorySize.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g"}
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, noMaxDirectMemorySize)
	g.Expect(err).To(HaveOccurred(), "JVM Option -XX:MaxDirectMemorySize cannot be empty")
	g.Expect(err.Error()).To(ContainSubstring("Missing required value for Segment Store JVM option -XX:MaxDirectMemorySize"))

	// Test Webhook with sum of MaxDirectMemorySize and Xmx being greater than total memory limit
	sumMaxDirectMemorySizeAndXmx := pravega_e2eutil.NewDefaultCluster(namespace)
	sumMaxDirectMemorySizeAndXmx.WithDefaults()
	sumMaxDirectMemorySizeAndXmx.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx2g", "-XX:MaxDirectMemorySize=2560m"}
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, sumMaxDirectMemorySizeAndXmx)
	g.Expect(err).To(HaveOccurred(), "sum of MaxDirectMemorySize and Xmx should be less than total memory limit")
	g.Expect(err.Error()).To(ContainSubstring("MaxDirectMemorySize(2684354560 B) along with JVM Xmx value(2147483648 B) should be less than the total available memory(4294967296 B)!"))

	// Test Webhook with pravegaservice.cache.size.max being greater than MaxDirectMemorySize
	cacheSizeGreaterThanMaxDirectMemorySize := pravega_e2eutil.NewDefaultCluster(namespace)
	cacheSizeGreaterThanMaxDirectMemorySize.WithDefaults()
	cacheSizeGreaterThanMaxDirectMemorySize.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "3221225472"
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, cacheSizeGreaterThanMaxDirectMemorySize)
	g.Expect(err).To(HaveOccurred(), "cache size configured should be less than MaxDirectMemorySize")
	g.Expect(err.Error()).To(ContainSubstring("Cache size(3221225472 B) configured should be less than the JVM MaxDirectMemorySize(2684354560 B) value"))

	// Test webhook with a valid Pravega cluster version format
	validVersion := pravega_e2eutil.NewClusterWithVersion(namespace, "0.6.0")
	validVersion.WithDefaults()
	pravega, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, validVersion)
	g.Expect(err).NotTo(HaveOccurred())

	podSize := 2
	err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(t, f, ctx, pravega, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// Try to downgrade the cluster
	pravega, err = pravega_e2eutil.GetPravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	pravega.Spec.Version = "0.5.0"
	err = pravega_e2eutil.UpdatePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).To(HaveOccurred(), "Should not allow downgrade")
	g.Expect(err.Error()).To(ContainSubstring("downgrading the cluster from version 0.6.0 to 0.5.0 is not supported"))

	// Delete cluster
	err = pravega_e2eutil.DeletePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WaitForPravegaClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	//creating the setup for running the bookkeeper validation check
	err = pravega_e2eutil.InitialSetup(t, f, ctx, namespace)
	g.Expect(err).NotTo(HaveOccurred())

	invalidEnsembleSize := pravega_e2eutil.NewDefaultCluster(namespace)
	invalidEnsembleSize.WithDefaults()
	invalidEnsembleSize.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3.4"
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, invalidEnsembleSize)
	g.Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.ensemble.size")
	g.Expect(err.Error()).To(ContainSubstring("Cannot convert ensemble size from string to integer"))

	invalidWriteQuorumSize := pravega_e2eutil.NewDefaultCluster(namespace)
	invalidWriteQuorumSize.WithDefaults()
	invalidWriteQuorumSize.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "3!4!"
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, invalidWriteQuorumSize)
	g.Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.write.quorum.size")
	g.Expect(err.Error()).To(ContainSubstring("Cannot convert write quorum size from string to integer"))

	invalidAckQuorumSize := pravega_e2eutil.NewDefaultCluster(namespace)
	invalidAckQuorumSize.WithDefaults()
	invalidAckQuorumSize.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "!44"
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, invalidAckQuorumSize)
	g.Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.ack.quorum.size")
	g.Expect(err.Error()).To(ContainSubstring("Cannot convert ack quorum size from string to integer"))

	invalidMinimumRacksCountEnable := pravega_e2eutil.NewDefaultCluster(namespace)
	invalidMinimumRacksCountEnable.WithDefaults()
	invalidMinimumRacksCountEnable.Spec.Pravega.Options["bookkeeper.write.quorum.racks.minimumCount.enable"] = "True"
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, invalidMinimumRacksCountEnable)
	g.Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.write.quorum.racks.minimumCount.enable")
	g.Expect(err.Error()).To(ContainSubstring("bookkeeper.write.quorum.racks.minimumCount.enable can be only set to \"true\" \"false\" or \"\""))

	ensembleSizeToOneAndRacksMinimumCountToTrue := pravega_e2eutil.NewDefaultCluster(namespace)
	ensembleSizeToOneAndRacksMinimumCountToTrue.WithDefaults()
	ensembleSizeToOneAndRacksMinimumCountToTrue.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "1"
	ensembleSizeToOneAndRacksMinimumCountToTrue.Spec.Pravega.Options["bookkeeper.write.quorum.racks.minimumCount.enable"] = "true"
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, ensembleSizeToOneAndRacksMinimumCountToTrue)
	g.Expect(err).To(HaveOccurred(), "Minimum Racks count should not be set to true when ensemble size is 1")
	g.Expect(err.Error()).To(ContainSubstring("bookkeeper.write.quorum.racks.minimumCount.enable should be set to false if bookkeeper.ensemble.size is 1"))

	ensembleSizeLessThanWriteQuorumSize := pravega_e2eutil.NewDefaultCluster(namespace)
	ensembleSizeLessThanWriteQuorumSize.WithDefaults()
	ensembleSizeLessThanWriteQuorumSize.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
	ensembleSizeLessThanWriteQuorumSize.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "4"
	ensembleSizeLessThanWriteQuorumSize.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, ensembleSizeLessThanWriteQuorumSize)
	g.Expect(err).To(HaveOccurred(), "Ensemble size should be greater than write quorum size")
	g.Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.write.quorum.size should be less than or equal to the value of option bookkeeper.ensemble.size"))

	ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault := pravega_e2eutil.NewDefaultCluster(namespace)
	ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.WithDefaults()
	ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "2"
	ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = ""
	ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = ""
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault)
	g.Expect(err).To(HaveOccurred(), "Ensemble size should be greater than the default value of write quorum size")
	g.Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.ensemble.size should be greater than or equal to the value of option bookkeeper.write.quorum.size (default is 3)"))

	ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree := pravega_e2eutil.NewDefaultCluster(namespace)
	ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.WithDefaults()
	ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ensemble.size"] = ""
	ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "4"
	ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree)
	g.Expect(err).To(HaveOccurred(), "The value for write quorum size should be less than default value of ensemble size")
	g.Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.write.quorum.size should be less than or equal to the value of option bookkeeper.ensemble.size (default is 3)"))

	writeQuorumSizeLessThanAckQuorumSize := pravega_e2eutil.NewDefaultCluster(namespace)
	writeQuorumSizeLessThanAckQuorumSize.WithDefaults()
	writeQuorumSizeLessThanAckQuorumSize.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
	writeQuorumSizeLessThanAckQuorumSize.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "2"
	writeQuorumSizeLessThanAckQuorumSize.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, writeQuorumSizeLessThanAckQuorumSize)
	g.Expect(err).To(HaveOccurred(), "The value for write quorum size should be greater than or equal to ack quorum size")
	g.Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.ack.quorum.size should be less than or equal to the value of option bookkeeper.write.quorum.size"))

	writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault := pravega_e2eutil.NewDefaultCluster(namespace)
	writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.WithDefaults()
	writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
	writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "2"
	writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = ""
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault)
	g.Expect(err).To(HaveOccurred(), "Write quorum size should be greater than the default value of ack quorum size")
	g.Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.write.quorum.size should be greater than or equal to the value of option bookkeeper.ack.quorum.size (default is 3)"))

	writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree := pravega_e2eutil.NewDefaultCluster(namespace)
	writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.WithDefaults()
	writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
	writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = ""
	writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "4"
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree)
	g.Expect(err).To(HaveOccurred(), "The value for ack quorum size should be less than default value of write quorum size")
	g.Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.ack.quorum.size should be less than or equal to the value of option bookkeeper.write.quorum.size (default is 3)"))

	validBookkeeperSettings := pravega_e2eutil.NewDefaultCluster(namespace)
	validBookkeeperSettings.WithDefaults()
	validBookkeeperSettings.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "4"
	validBookkeeperSettings.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "3"
	validBookkeeperSettings.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "2"
	pravega, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, validBookkeeperSettings)
	g.Expect(err).NotTo(HaveOccurred())

	// Deleting the bookkeeper validation check cluster
	err = pravega_e2eutil.DeletePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = pravega_e2eutil.WaitForPravegaClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	//creating the setup for running the bookkeeper validation check
	err = pravega_e2eutil.InitialSetup(t, f, ctx, namespace)
	g.Expect(err).NotTo(HaveOccurred())
	authsettingsValidation := pravega_e2eutil.NewDefaultCluster(namespace)
	authsettingsValidation.WithDefaults()
	authsettingsValidation.Spec.Authentication.Enabled = true
	pravega, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, authsettingsValidation)
	g.Expect(err).To(HaveOccurred(), "The field autoScale.controller.connect.security.auth.enable should be present")
	g.Expect(err.Error()).To(ContainSubstring("autoScale.controller.connect.security.auth.enable field is not present"))
	authsettingsValidation.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "false"
	pravega, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, authsettingsValidation)
	g.Expect(err).To(HaveOccurred(), "The value for autoScale.controller.connect.security.auth.enable should not be false")
	g.Expect(err.Error()).To(ContainSubstring("autoScale.controller.connect.security.auth.enable/autoScale.authEnabled should be set to true"))
	authsettingsValidation.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "true"
	pravega, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, authsettingsValidation)
	g.Expect(err).To(HaveOccurred(), "Controller token sigining key should be present")
	g.Expect(err.Error()).To(ContainSubstring("controller.security.auth.delegationToken.signingKey.basis field is not present"))
	authsettingsValidation.Spec.Pravega.Options["controller.security.auth.delegationToken.signingKey.basis"] = "secret"
	pravega, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, authsettingsValidation)
	g.Expect(err).To(HaveOccurred(), "Segmentstore token sigining key should be present")
	g.Expect(err.Error()).To(ContainSubstring("autoScale.security.auth.token.signingKey.basis field is not present"))
	authsettingsValidation.Spec.Pravega.Options["autoScale.security.auth.token.signingKey.basis"] = "secret1"
	pravega, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, authsettingsValidation)
	g.Expect(err).To(HaveOccurred(), "Segmentstore and controller token sigining key should be same ")
	g.Expect(err.Error()).To(ContainSubstring("controller and segmentstore token signing key should have same value"))

}
