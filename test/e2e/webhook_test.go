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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Webhook validation", func() {
	Context("Validating  webhook checks", func() {
		It("Webhook validations should be successful", func() {

			//creating the setup for running the test

			Expect(pravega_e2eutil.InitialSetup(&t, k8sClient, testNamespace)).NotTo(HaveOccurred())

			//Test webhook with an invalid Pravega cluster version format
			invalidVersion := pravega_e2eutil.NewClusterWithVersion(testNamespace, "999")
			invalidVersion.WithDefaults()
			_, err := pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, invalidVersion)
			Expect(err).To(HaveOccurred(), "Should reject deployment of invalid version format")
			Expect(err.Error()).To(ContainSubstring("request version is not in valid format:"))

			// Test webhook with with no segementStoreResources object
			noSegmentStoreResource := pravega_e2eutil.NewDefaultCluster(testNamespace)
			noSegmentStoreResource.WithDefaults()
			noSegmentStoreResource.Spec.Pravega.SegmentStoreResources = nil
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, noSegmentStoreResource)
			Expect(err).To(HaveOccurred(), "Spec.Pravega.SegmentStoreResources cannot be empty")
			Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources cannot be empty"))

			// Test webhook with with no segementStoreResources.Limits object
			noSegmentStoreResourceLimits := pravega_e2eutil.NewDefaultCluster(testNamespace)
			noSegmentStoreResourceLimits.WithDefaults()
			noSegmentStoreResourceLimits.Spec.Pravega.SegmentStoreResources.Limits = nil
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, noSegmentStoreResourceLimits)
			Expect(err).To(HaveOccurred(), "Spec.Pravega.SegmentStoreResources.Limits cannot be empty")
			Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources.limits cannot be empty"))

			// Test webhook with with no segementStoreResources.Requests object
			noSegmentStoreResourceRequests := pravega_e2eutil.NewDefaultCluster(testNamespace)
			noSegmentStoreResourceRequests.WithDefaults()
			noSegmentStoreResourceRequests.Spec.Pravega.SegmentStoreResources.Requests = nil
			pravega, err := pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, noSegmentStoreResourceRequests)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.DeletePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToTerminate(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			//creating the setup for running the test
			err = pravega_e2eutil.InitialSetup(&t, k8sClient, testNamespace)
			Expect(err).NotTo(HaveOccurred())

			// Test webhook with no value for segment store memory limits
			noMemoryLimits := pravega_e2eutil.NewClusterWithNoSegmentStoreMemoryLimits(testNamespace)
			noMemoryLimits.WithDefaults()
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, noMemoryLimits)
			Expect(err).To(HaveOccurred(), "Segment Store memory limits cannot be empty")
			Expect(err.Error()).To(ContainSubstring("Missing required value for field spec.pravega.segmentStoreResources.limits.memory"))

			// Test webhook with no value for segment store cpu limits
			noCpuLimits := pravega_e2eutil.NewClusterWithNoSegmentStoreCpuLimits(testNamespace)
			noCpuLimits.WithDefaults()
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, noCpuLimits)
			Expect(err).To(HaveOccurred(), "Segment Store cpu limits cannot be empty")
			Expect(err.Error()).To(ContainSubstring("Missing required value for field spec.pravega.segmentStoreResources.limits.cpu"))

			// Test webhook with segment store memory requests being greater than memory limits
			memoryRequestsGreaterThanLimits := pravega_e2eutil.NewDefaultCluster(testNamespace)
			memoryRequestsGreaterThanLimits.WithDefaults()
			memoryRequestsGreaterThanLimits.Spec.Pravega.SegmentStoreResources.Requests[corev1.ResourceMemory] = resource.MustParse("5Gi")
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, memoryRequestsGreaterThanLimits)
			Expect(err).To(HaveOccurred(), "Segment Store memory requests should be less than or equal to limits")
			Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources.requests.memory value must be less than or equal to spec.pravega.segmentStoreResources.limits.memory"))

			// Test webhook with segment store cpu requests being greater than cpu limits
			cpuRequestsGreaterThanLimits := pravega_e2eutil.NewDefaultCluster(testNamespace)
			cpuRequestsGreaterThanLimits.WithDefaults()
			cpuRequestsGreaterThanLimits.Spec.Pravega.SegmentStoreResources.Requests[corev1.ResourceCPU] = resource.MustParse("3000m")
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, cpuRequestsGreaterThanLimits)
			Expect(err).To(HaveOccurred(), "Segment Store cpu requests should be less than or equal to limits")
			Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources.requests.cpu value must be less than or equal to spec.pravega.segmentStoreResources.limits.cpu"))

			// Test webhook with no value for option pravegaservice.cache.size.max
			cacheSizeMax := pravega_e2eutil.NewDefaultCluster(testNamespace)
			cacheSizeMax.WithDefaults()
			cacheSizeMax.Spec.Pravega.Options["pravegaservice.cache.size.max"] = ""
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, cacheSizeMax)
			Expect(err).To(HaveOccurred(), "pravegaservice.cache.size.max cannot be empty")
			Expect(err.Error()).To(ContainSubstring("Missing required value for option pravegaservice.cache.size.max"))

			// Test Webhook with no value for JVM option -Xmx
			noXmx := pravega_e2eutil.NewDefaultCluster(testNamespace)
			noXmx.WithDefaults()
			noXmx.Spec.Pravega.SegmentStoreJVMOptions = []string{"-XX:MaxDirectMemorySize=2560m"}
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, noXmx)
			Expect(err).To(HaveOccurred(), "JVM option -Xmx cannot be empty")
			Expect(err.Error()).To(ContainSubstring("Missing required value for Segment Store JVM Option -Xmx"))

			// Test Webhook with no value for JVM option -XX:MaxDirectMemorySize
			noMaxDirectMemorySize := pravega_e2eutil.NewDefaultCluster(testNamespace)
			noMaxDirectMemorySize.WithDefaults()
			noMaxDirectMemorySize.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g"}
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, noMaxDirectMemorySize)
			Expect(err).To(HaveOccurred(), "JVM Option -XX:MaxDirectMemorySize cannot be empty")
			Expect(err.Error()).To(ContainSubstring("Missing required value for Segment Store JVM option -XX:MaxDirectMemorySize"))

			// Test Webhook with sum of MaxDirectMemorySize and Xmx being greater than total memory limit
			sumMaxDirectMemorySizeAndXmx := pravega_e2eutil.NewDefaultCluster(testNamespace)
			sumMaxDirectMemorySizeAndXmx.WithDefaults()
			sumMaxDirectMemorySizeAndXmx.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx2g", "-XX:MaxDirectMemorySize=2560m"}
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, sumMaxDirectMemorySizeAndXmx)
			Expect(err).To(HaveOccurred(), "sum of MaxDirectMemorySize and Xmx should be less than total memory limit")
			Expect(err.Error()).To(ContainSubstring("MaxDirectMemorySize(2684354560 B) along with JVM Xmx value(2147483648 B) should be less than the total available memory(4294967296 B)!"))

			// Test Webhook with pravegaservice.cache.size.max being greater than MaxDirectMemorySize
			cacheSizeGreaterThanMaxDirectMemorySize := pravega_e2eutil.NewDefaultCluster(testNamespace)
			cacheSizeGreaterThanMaxDirectMemorySize.WithDefaults()
			cacheSizeGreaterThanMaxDirectMemorySize.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "3221225472"
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, cacheSizeGreaterThanMaxDirectMemorySize)
			Expect(err).To(HaveOccurred(), "cache size configured should be less than MaxDirectMemorySize")
			Expect(err.Error()).To(ContainSubstring("Cache size(3221225472 B) configured should be less than the JVM MaxDirectMemorySize(2684354560 B) value"))

			// Test webhook with a valid Pravega cluster version format
			validVersion := pravega_e2eutil.NewClusterWithVersion(testNamespace, "0.6.0")
			validVersion.WithDefaults()
			pravega, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, validVersion)
			Expect(err).NotTo(HaveOccurred())

			podSize := 2
			err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(&t, k8sClient, pravega, podSize)
			Expect(err).NotTo(HaveOccurred())

			// Try to downgrade the cluster
			pravega, err = pravega_e2eutil.GetPravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			pravega.Spec.Version = "0.5.0"
			err = pravega_e2eutil.UpdatePravegaCluster(&t, k8sClient, pravega)
			Expect(err).To(HaveOccurred(), "Should not allow downgrade")
			Expect(err.Error()).To(ContainSubstring("downgrading the cluster from version 0.6.0 to 0.5.0 is not supported"))

			// Delete cluster
			err = pravega_e2eutil.DeletePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToTerminate(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			//creating the setup for running the bookkeeper validation check
			err = pravega_e2eutil.InitialSetup(&t, k8sClient, testNamespace)
			Expect(err).NotTo(HaveOccurred())

			invalidEnsembleSize := pravega_e2eutil.NewDefaultCluster(testNamespace)
			invalidEnsembleSize.WithDefaults()
			invalidEnsembleSize.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3.4"
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, invalidEnsembleSize)
			Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.ensemble.size")
			Expect(err.Error()).To(ContainSubstring("Cannot convert ensemble size from string to integer"))

			invalidWriteQuorumSize := pravega_e2eutil.NewDefaultCluster(testNamespace)
			invalidWriteQuorumSize.WithDefaults()
			invalidWriteQuorumSize.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "3!4!"
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, invalidWriteQuorumSize)
			Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.write.quorum.size")
			Expect(err.Error()).To(ContainSubstring("Cannot convert write quorum size from string to integer"))

			invalidAckQuorumSize := pravega_e2eutil.NewDefaultCluster(testNamespace)
			invalidAckQuorumSize.WithDefaults()
			invalidAckQuorumSize.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "!44"
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, invalidAckQuorumSize)
			Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.ack.quorum.size")
			Expect(err.Error()).To(ContainSubstring("Cannot convert ack quorum size from string to integer"))

			invalidMinimumRacksCountEnable := pravega_e2eutil.NewDefaultCluster(testNamespace)
			invalidMinimumRacksCountEnable.WithDefaults()
			invalidMinimumRacksCountEnable.Spec.Pravega.Options["bookkeeper.write.quorum.racks.minimumCount.enable"] = "True"
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, invalidMinimumRacksCountEnable)
			Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.write.quorum.racks.minimumCount.enable")
			Expect(err.Error()).To(ContainSubstring("bookkeeper.write.quorum.racks.minimumCount.enable can be only set to \"true\" \"false\" or \"\""))

			ensembleSizeToOneAndRacksMinimumCountToTrue := pravega_e2eutil.NewDefaultCluster(testNamespace)
			ensembleSizeToOneAndRacksMinimumCountToTrue.WithDefaults()
			ensembleSizeToOneAndRacksMinimumCountToTrue.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "1"
			ensembleSizeToOneAndRacksMinimumCountToTrue.Spec.Pravega.Options["bookkeeper.write.quorum.racks.minimumCount.enable"] = "true"
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, ensembleSizeToOneAndRacksMinimumCountToTrue)
			Expect(err).To(HaveOccurred(), "Minimum Racks count should not be set to true when ensemble size is 1")
			Expect(err.Error()).To(ContainSubstring("bookkeeper.write.quorum.racks.minimumCount.enable should be set to false if bookkeeper.ensemble.size is 1"))

			ensembleSizeLessThanWriteQuorumSize := pravega_e2eutil.NewDefaultCluster(testNamespace)
			ensembleSizeLessThanWriteQuorumSize.WithDefaults()
			ensembleSizeLessThanWriteQuorumSize.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
			ensembleSizeLessThanWriteQuorumSize.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "4"
			ensembleSizeLessThanWriteQuorumSize.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, ensembleSizeLessThanWriteQuorumSize)
			Expect(err).To(HaveOccurred(), "Ensemble size should be greater than write quorum size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.write.quorum.size should be less than or equal to the value of option bookkeeper.ensemble.size"))

			ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault := pravega_e2eutil.NewDefaultCluster(testNamespace)
			ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.WithDefaults()
			ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "2"
			ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = ""
			ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = ""
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault)
			Expect(err).To(HaveOccurred(), "Ensemble size should be greater than the default value of write quorum size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.ensemble.size should be greater than or equal to the value of option bookkeeper.write.quorum.size (default is 3)"))

			ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree := pravega_e2eutil.NewDefaultCluster(testNamespace)
			ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.WithDefaults()
			ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ensemble.size"] = ""
			ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "4"
			ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree)
			Expect(err).To(HaveOccurred(), "The value for write quorum size should be less than default value of ensemble size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.write.quorum.size should be less than or equal to the value of option bookkeeper.ensemble.size (default is 3)"))

			writeQuorumSizeLessThanAckQuorumSize := pravega_e2eutil.NewDefaultCluster(testNamespace)
			writeQuorumSizeLessThanAckQuorumSize.WithDefaults()
			writeQuorumSizeLessThanAckQuorumSize.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
			writeQuorumSizeLessThanAckQuorumSize.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "2"
			writeQuorumSizeLessThanAckQuorumSize.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, writeQuorumSizeLessThanAckQuorumSize)
			Expect(err).To(HaveOccurred(), "The value for write quorum size should be greater than or equal to ack quorum size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.ack.quorum.size should be less than or equal to the value of option bookkeeper.write.quorum.size"))

			writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault := pravega_e2eutil.NewDefaultCluster(testNamespace)
			writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.WithDefaults()
			writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
			writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "2"
			writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = ""
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault)
			Expect(err).To(HaveOccurred(), "Write quorum size should be greater than the default value of ack quorum size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.write.quorum.size should be greater than or equal to the value of option bookkeeper.ack.quorum.size (default is 3)"))

			writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree := pravega_e2eutil.NewDefaultCluster(testNamespace)
			writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.WithDefaults()
			writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
			writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = ""
			writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "4"
			_, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree)
			Expect(err).To(HaveOccurred(), "The value for ack quorum size should be less than default value of write quorum size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.ack.quorum.size should be less than or equal to the value of option bookkeeper.write.quorum.size (default is 3)"))

			validBookkeeperSettings := pravega_e2eutil.NewDefaultCluster(testNamespace)
			validBookkeeperSettings.WithDefaults()
			validBookkeeperSettings.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "4"
			validBookkeeperSettings.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "3"
			validBookkeeperSettings.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "2"
			pravega, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, validBookkeeperSettings)
			Expect(err).NotTo(HaveOccurred())

			// Deleting the bookkeeper validation check cluster
			err = pravega_e2eutil.DeletePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToTerminate(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			//creating the setup for running the bookkeeper validation check
			err = pravega_e2eutil.InitialSetup(&t, k8sClient, testNamespace)
			Expect(err).NotTo(HaveOccurred())
			authsettingsValidation := pravega_e2eutil.NewDefaultCluster(testNamespace)
			authsettingsValidation.WithDefaults()
			authsettingsValidation.Spec.Authentication.Enabled = true
			pravega, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "The field autoScale.controller.connect.security.auth.enable should be present")
			Expect(err.Error()).To(ContainSubstring("autoScale.controller.connect.security.auth.enable field is not present"))

			authsettingsValidation.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "false"
			pravega, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "The value for autoScale.controller.connect.security.auth.enable should not be false")
			Expect(err.Error()).To(ContainSubstring("autoScale.controller.connect.security.auth.enable should be set to true"))

			authsettingsValidation.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "dummy"
			authsettingsValidation.Spec.Pravega.Options["autoScale.authEnabled"] = "dummy"
			pravega, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "The value for autoScale.controller.connect.security.auth.enable/autoScale.authEnabled should not be incorrect")
			Expect(err.Error()).To(ContainSubstring("autoScale.controller.connect.security.auth.enable and autoScale.authEnabled should be set to true"))

			authsettingsValidation.Spec.Pravega.Options["autoScale.authEnabled"] = ""
			authsettingsValidation.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "true"
			pravega, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "Controller token sigining key should be present")
			Expect(err.Error()).To(ContainSubstring("controller.security.auth.delegationToken.signingKey.basis field is not present"))

			authsettingsValidation.Spec.Pravega.Options["controller.security.auth.delegationToken.signingKey.basis"] = "secret"
			pravega, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "Segmentstore token sigining key should be present")
			Expect(err.Error()).To(ContainSubstring("autoScale.security.auth.token.signingKey.basis field is not present"))

			authsettingsValidation.Spec.Pravega.Options["autoScale.security.auth.token.signingKey.basis"] = "secret1"
			pravega, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "Segmentstore and controller token sigining key should be same ")
			Expect(err.Error()).To(ContainSubstring("controller and segmentstore token signing key should have same value"))

			authsettingsValidation.Spec.Authentication.Enabled = false
			authsettingsValidation.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "true"
			pravega, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "The field autoScale.controller.connect.security.auth.enable should not be set")
			Expect(err.Error()).To(ContainSubstring("autoScale.controller.connect.security.auth.enable/autoScale.authEnabled should not be set to true"))

		})
	})
})
