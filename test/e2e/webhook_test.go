/**
 * Copyright (c) 2019-2022 Dell Inc., or its subsidiaries. All Rights Reserved.
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
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/pravega/pravega-operator/test/e2e/e2eutil"
)

var _ = Describe("Create and recreate Pravega cluster with same name", func() {
	namespace := "default"
	var err error
	var pravega *v1beta1.PravegaCluster
	Context("When creating a new cluster with invalid version", func() {
		It("should fail", func() {
			// creating the setup for running the test
			err = e2eutil.InitialSetup(k8sClient, namespace)
			Expect(err).NotTo(HaveOccurred())

			// Test webhook with an invalid Pravega cluster version format
			invalidVersion := e2eutil.NewClusterWithVersion(namespace, "999")
			invalidVersion.WithDefaults()
			_, err = e2eutil.CreatePravegaCluster(k8sClient, invalidVersion)
			Expect(err).To(HaveOccurred(), "Should reject deployment of invalid version format")
			Expect(err.Error()).To(ContainSubstring("request version is not in valid format:"))
		})
	})
	Context("When creating a cluster with no segmentStoreResources", func() {
		It("should fail", func() {
			// Test webhook with no segementStoreResources object
			noSegmentStoreResource := e2eutil.NewDefaultCluster(namespace)
			noSegmentStoreResource.WithDefaults()
			noSegmentStoreResource.Spec.Pravega.SegmentStoreResources = nil
			_, err = e2eutil.CreatePravegaCluster(k8sClient, noSegmentStoreResource)
			Expect(err).To(HaveOccurred(), "Spec.Pravega.SegmentStoreResources cannot be empty")
			Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources cannot be empty"))

			// Test webhook with with no segementStoreResources.Limits object
			noSegmentStoreResourceLimits := e2eutil.NewDefaultCluster(namespace)
			noSegmentStoreResourceLimits.WithDefaults()
			noSegmentStoreResourceLimits.Spec.Pravega.SegmentStoreResources.Limits = nil
			_, err = e2eutil.CreatePravegaCluster(k8sClient, noSegmentStoreResourceLimits)
			Expect(err).To(HaveOccurred(), "Spec.Pravega.SegmentStoreResources.Limits cannot be empty")
			Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources.limits cannot be empty"))
		})
	})
	Context("When creating a cluster with no segmentStoreRequests", func() {
		It("should succeed", func() {
			// Test webhook with with no segementStoreResources.Requests object
			noSegmentStoreResourceRequests := e2eutil.NewDefaultCluster(namespace)
			noSegmentStoreResourceRequests.WithDefaults()
			noSegmentStoreResourceRequests.Spec.Pravega.SegmentStoreResources.Requests = nil
			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, noSegmentStoreResourceRequests)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeletePravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			It("should tear down the cluster successfully", func() {
				err = e2eutil.DeletePravegaCluster(k8sClient, pravega)
				Expect(err).NotTo(HaveOccurred())
				Eventually(e2eutil.WaitForPravegaClusterToTerminate(k8sClient, pravega), timeout).Should(Succeed())
			})
		})
	})

	Context("Testing segment store cpu and memory limits/requests", func() {
		// creating the setup for running the test
		err = e2eutil.InitialSetup(k8sClient, namespace)
		Expect(err).NotTo(HaveOccurred())

		It("should fail to create pravega cluster with no segment store memory limits", func() {
			// Test webhook with no value for segment store memory limits
			noMemoryLimits := e2eutil.NewClusterWithNoSegmentStoreMemoryLimits(namespace)
			noMemoryLimits.WithDefaults()
			_, err = e2eutil.CreatePravegaCluster(k8sClient, noMemoryLimits)
			Expect(err).To(HaveOccurred(), "Segment Store memory limits cannot be empty")
			Expect(err.Error()).To(ContainSubstring("Missing required value for field spec.pravega.segmentStoreResources.limits.memory"))
		})
		It("should fail to create pravega cluster with no segment store CPU limits", func() {
			// Test webhook with no value for segment store cpu limits
			noCpuLimits := e2eutil.NewClusterWithNoSegmentStoreCpuLimits(namespace)
			noCpuLimits.WithDefaults()
			_, err = e2eutil.CreatePravegaCluster(k8sClient, noCpuLimits)
			Expect(err).To(HaveOccurred(), "Segment Store cpu limits cannot be empty")
			Expect(err.Error()).To(ContainSubstring("Missing required value for field spec.pravega.segmentStoreResources.limits.cpu"))
		})

		It("should fail to create pravega cluster with memory requests > memory limits", func() {
			// Test webhook with segment store memory requests being greater than memory limits
			memoryRequestsGreaterThanLimits := e2eutil.NewDefaultCluster(namespace)
			memoryRequestsGreaterThanLimits.WithDefaults()
			memoryRequestsGreaterThanLimits.Spec.Pravega.SegmentStoreResources.Requests[corev1.ResourceMemory] = resource.MustParse("5Gi")
			_, err = e2eutil.CreatePravegaCluster(k8sClient, memoryRequestsGreaterThanLimits)
			Expect(err).To(HaveOccurred(), "Segment Store memory requests should be less than or equal to limits")
			Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources.requests.memory value must be less than or equal to spec.pravega.segmentStoreResources.limits.memory"))
		})

		It("should fail to create pravega cluster with cpu requests > cpu limits", func() {
			// Test webhook with segment store cpu requests being greater than cpu limits
			cpuRequestsGreaterThanLimits := e2eutil.NewDefaultCluster(namespace)
			cpuRequestsGreaterThanLimits.WithDefaults()
			cpuRequestsGreaterThanLimits.Spec.Pravega.SegmentStoreResources.Requests[corev1.ResourceCPU] = resource.MustParse("3000m")
			_, err = e2eutil.CreatePravegaCluster(k8sClient, cpuRequestsGreaterThanLimits)
			Expect(err).To(HaveOccurred(), "Segment Store cpu requests should be less than or equal to limits")
			Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources.requests.cpu value must be less than or equal to spec.pravega.segmentStoreResources.limits.cpu"))
		})

		It("should fail to create pravega cluster with no value for option pravegaservice.cache.size.max", func() {
			// Test webhook with no value for option pravegaservice.cache.size.max
			cacheSizeMax := e2eutil.NewDefaultCluster(namespace)
			cacheSizeMax.WithDefaults()
			cacheSizeMax.Spec.Pravega.Options["pravegaservice.cache.size.max"] = ""
			_, err = e2eutil.CreatePravegaCluster(k8sClient, cacheSizeMax)
			Expect(err).To(HaveOccurred(), "pravegaservice.cache.size.max cannot be empty")
			Expect(err.Error()).To(ContainSubstring("Missing required value for option pravegaservice.cache.size.max"))
		})

		It("should fail to create pravega cluster with no valu efor JVM option -Xmx", func() {
			// Test Webhook with no value for JVM option -Xmx
			noXmx := e2eutil.NewDefaultCluster(namespace)
			noXmx.WithDefaults()
			noXmx.Spec.Pravega.SegmentStoreJVMOptions = []string{"-XX:MaxDirectMemorySize=2560m"}
			_, err = e2eutil.CreatePravegaCluster(k8sClient, noXmx)
			Expect(err).To(HaveOccurred(), "JVM option -Xmx cannot be empty")
			Expect(err.Error()).To(ContainSubstring("Missing required value for Segment Store JVM Option -Xmx"))
		})

		It("should fail to create pravega cluster with no value for JVM option -XX:MaxDirectMemorySize", func() {
			// Test Webhook with no value for JVM option -XX:MaxDirectMemorySize
			noMaxDirectMemorySize := e2eutil.NewDefaultCluster(namespace)
			noMaxDirectMemorySize.WithDefaults()
			noMaxDirectMemorySize.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g"}
			_, err = e2eutil.CreatePravegaCluster(k8sClient, noMaxDirectMemorySize)
			Expect(err).To(HaveOccurred(), "JVM Option -XX:MaxDirectMemorySize cannot be empty")
			Expect(err.Error()).To(ContainSubstring("Missing required value for Segment Store JVM option -XX:MaxDirectMemorySize"))
		})

		It("should fail to create pravega cluster with sum of MaxDirectMemorySize and Xmx being greater than total memory limit", func() {
			// Test Webhook with sum of MaxDirectMemorySize and Xmx being greater than total memory limit
			sumMaxDirectMemorySizeAndXmx := e2eutil.NewDefaultCluster(namespace)
			sumMaxDirectMemorySizeAndXmx.WithDefaults()
			sumMaxDirectMemorySizeAndXmx.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx2g", "-XX:MaxDirectMemorySize=2560m"}
			_, err = e2eutil.CreatePravegaCluster(k8sClient, sumMaxDirectMemorySizeAndXmx)
			Expect(err).To(HaveOccurred(), "sum of MaxDirectMemorySize and Xmx should be less than total memory limit")
			Expect(err.Error()).To(ContainSubstring("MaxDirectMemorySize(2684354560 B) along with JVM Xmx value(2147483648 B) should be less than the total available memory(4294967296 B)!"))
		})

		It("should fail with pravegaservice.cache.size.max being greater than MaxDirectMemorySize", func() {
			// Test Webhook with pravegaservice.cache.size.max being greater than MaxDirectMemorySize
			cacheSizeGreaterThanMaxDirectMemorySize := e2eutil.NewDefaultCluster(namespace)
			cacheSizeGreaterThanMaxDirectMemorySize.WithDefaults()
			cacheSizeGreaterThanMaxDirectMemorySize.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "3221225472"
			_, err = e2eutil.CreatePravegaCluster(k8sClient, cacheSizeGreaterThanMaxDirectMemorySize)
			Expect(err).To(HaveOccurred(), "cache size configured should be less than MaxDirectMemorySize")
			Expect(err.Error()).To(ContainSubstring("Cache size(3221225472 B) configured should be less than the JVM MaxDirectMemorySize(2684354560 B) value"))
		})

		It("should succeed to install with valid version", func() {
			// Test webhook with a valid Pravega cluster version format
			validVersion := e2eutil.NewClusterWithVersion(namespace, "0.6.0")
			validVersion.WithDefaults()
			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, validVersion)
			Expect(err).NotTo(HaveOccurred())

			podCount := 2
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
			Eventually(e2eutil.WaitForPravegaClusterToBecomeReady(k8sClient, pravega, podCount), timeout).Should(Succeed())
		})
		It("should fail to downgrade", func() {
			// Try to downgrade the cluster
			pravega, err = e2eutil.GetPravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			pravega.Spec.Version = "0.5.0"
			err = e2eutil.UpdatePravegaCluster(k8sClient, pravega)
			Expect(err).To(HaveOccurred(), "Should not allow downgrade")
			Expect(err.Error()).To(ContainSubstring("downgrading the cluster from version 0.6.0 to 0.5.0 is not supported"))
		})

		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeletePravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForPravegaClusterToTerminate(k8sClient, pravega), timeout).Should(Succeed())
		})

		It("should fail to create pravega cluster with invalidEnsembleSize", func() {
			// creating the setup for running the bookkeeper validation check
			err = e2eutil.InitialSetup(k8sClient, namespace)
			Expect(err).NotTo(HaveOccurred())
			invalidEnsembleSize := e2eutil.NewDefaultCluster(namespace)
			invalidEnsembleSize.WithDefaults()
			invalidEnsembleSize.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3.4"
			_, err = e2eutil.CreatePravegaCluster(k8sClient, invalidEnsembleSize)
			Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.ensemble.size")
			Expect(err.Error()).To(ContainSubstring("Cannot convert ensemble size from string to integer"))
		})

		It("should fail to create pravega cluster with invalidWriteQuorumSize", func() {
			invalidWriteQuorumSize := e2eutil.NewDefaultCluster(namespace)
			invalidWriteQuorumSize.WithDefaults()
			invalidWriteQuorumSize.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "3!4!"
			_, err = e2eutil.CreatePravegaCluster(k8sClient, invalidWriteQuorumSize)
			Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.write.quorum.size")
			Expect(err.Error()).To(ContainSubstring("Cannot convert write quorum size from string to integer"))
		})

		It("should fail to create pravega cluster with invalidAckQuorumSize", func() {
			invalidAckQuorumSize := e2eutil.NewDefaultCluster(namespace)
			invalidAckQuorumSize.WithDefaults()
			invalidAckQuorumSize.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "!44"
			_, err = e2eutil.CreatePravegaCluster(k8sClient, invalidAckQuorumSize)
			Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.ack.quorum.size")
			Expect(err.Error()).To(ContainSubstring("Cannot convert ack quorum size from string to integer"))
		})

		It("should fail to create pravega cluster with invalidMinimumRacksCountEnable", func() {
			invalidMinimumRacksCountEnable := e2eutil.NewDefaultCluster(namespace)
			invalidMinimumRacksCountEnable.WithDefaults()
			invalidMinimumRacksCountEnable.Spec.Pravega.Options["bookkeeper.write.quorum.racks.minimumCount.enable"] = "True"
			_, err = e2eutil.CreatePravegaCluster(k8sClient, invalidMinimumRacksCountEnable)
			Expect(err).To(HaveOccurred(), "Invalid value for option bookkeeper.write.quorum.racks.minimumCount.enable")
			Expect(err.Error()).To(ContainSubstring("bookkeeper.write.quorum.racks.minimumCount.enable can be only set to \"true\" \"false\" or \"\""))
		})

		It("should fail to create pravega cluster with ensembleSizeToOneAndRacksMinimumCountToTrue", func() {
			ensembleSizeToOneAndRacksMinimumCountToTrue := e2eutil.NewDefaultCluster(namespace)
			ensembleSizeToOneAndRacksMinimumCountToTrue.WithDefaults()
			ensembleSizeToOneAndRacksMinimumCountToTrue.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "1"
			ensembleSizeToOneAndRacksMinimumCountToTrue.Spec.Pravega.Options["bookkeeper.write.quorum.racks.minimumCount.enable"] = "true"
			_, err = e2eutil.CreatePravegaCluster(k8sClient, ensembleSizeToOneAndRacksMinimumCountToTrue)
			Expect(err).To(HaveOccurred(), "Minimum Racks count should not be set to true when ensemble size is 1")
			Expect(err.Error()).To(ContainSubstring("bookkeeper.write.quorum.racks.minimumCount.enable should be set to false if bookkeeper.ensemble.size is 1"))
		})

		It("should fail to create pravega cluster with ensembleSizeLessThanWriteQuorumSize", func() {
			ensembleSizeLessThanWriteQuorumSize := e2eutil.NewDefaultCluster(namespace)
			ensembleSizeLessThanWriteQuorumSize.WithDefaults()
			ensembleSizeLessThanWriteQuorumSize.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
			ensembleSizeLessThanWriteQuorumSize.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "4"
			ensembleSizeLessThanWriteQuorumSize.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
			_, err = e2eutil.CreatePravegaCluster(k8sClient, ensembleSizeLessThanWriteQuorumSize)
			Expect(err).To(HaveOccurred(), "Ensemble size should be greater than write quorum size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.write.quorum.size should be less than or equal to the value of option bookkeeper.ensemble.size"))
		})

		It("should fail tocreate pravega cluster with ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault", func() {
			ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault := e2eutil.NewDefaultCluster(namespace)
			ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.WithDefaults()
			ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "2"
			ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = ""
			ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = ""
			_, err = e2eutil.CreatePravegaCluster(k8sClient, ensembleSizeLessThanEqualToTwoWriteQuorumSizeSetToDefault)
			Expect(err).To(HaveOccurred(), "Ensemble size should be greater than the default value of write quorum size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.ensemble.size should be greater than or equal to the value of option bookkeeper.write.quorum.size (default is 3)"))
		})

		It("should create pravega cluster with ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree", func() {
			ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree := e2eutil.NewDefaultCluster(namespace)
			ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.WithDefaults()
			ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ensemble.size"] = ""
			ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "4"
			ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
			_, err = e2eutil.CreatePravegaCluster(k8sClient, ensembleSizeSetToDefaultWriteQuorumSizeGreaterThanThree)
			Expect(err).To(HaveOccurred(), "The value for write quorum size should be less than default value of ensemble size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.write.quorum.size should be less than or equal to the value of option bookkeeper.ensemble.size (default is 3)"))
		})

		It("should fail to create pravega cluster with writeQuorumSizeLessThanAckQuorumSize", func() {
			writeQuorumSizeLessThanAckQuorumSize := e2eutil.NewDefaultCluster(namespace)
			writeQuorumSizeLessThanAckQuorumSize.WithDefaults()
			writeQuorumSizeLessThanAckQuorumSize.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
			writeQuorumSizeLessThanAckQuorumSize.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "2"
			writeQuorumSizeLessThanAckQuorumSize.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
			_, err = e2eutil.CreatePravegaCluster(k8sClient, writeQuorumSizeLessThanAckQuorumSize)
			Expect(err).To(HaveOccurred(), "The value for write quorum size should be greater than or equal to ack quorum size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.ack.quorum.size should be less than or equal to the value of option bookkeeper.write.quorum.size"))
		})

		It("should fail to create pravega cluster with lessThanEqualToTwoAckQuorumSizeSet", func() {
			writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault := e2eutil.NewDefaultCluster(namespace)
			writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.WithDefaults()
			writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
			writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "2"
			writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = ""
			_, err = e2eutil.CreatePravegaCluster(k8sClient, writeQuorumSizeLessThanEqualToTwoAckQuorumSizeSetToDefault)
			Expect(err).To(HaveOccurred(), "Write quorum size should be greater than the default value of ack quorum size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.write.quorum.size should be greater than or equal to the value of option bookkeeper.ack.quorum.size (default is 3)"))
		})

		It("should fail to create pravega cluster with writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree", func() {
			writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree := e2eutil.NewDefaultCluster(namespace)
			writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.WithDefaults()
			writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
			writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = ""
			writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "4"
			_, err = e2eutil.CreatePravegaCluster(k8sClient, writeQuorumSizeSetToDefaultAckQuorumSizeGreaterThanThree)
			Expect(err).To(HaveOccurred(), "The value for ack quorum size should be less than default value of write quorum size")
			Expect(err.Error()).To(ContainSubstring("The value provided for the option bookkeeper.ack.quorum.size should be less than or equal to the value of option bookkeeper.write.quorum.size (default is 3)"))
		})

		It("should create pravega cluster with bookkeeper options", func() {
			validBookkeeperSettings := e2eutil.NewDefaultCluster(namespace)
			validBookkeeperSettings.WithDefaults()
			validBookkeeperSettings.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "4"
			validBookkeeperSettings.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "3"
			validBookkeeperSettings.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "2"
			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, validBookkeeperSettings)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeletePravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForPravegaClusterToTerminate(k8sClient, pravega), timeout).Should(Succeed())
		})

		It("should fail with invalid auth settings", func() {
			// creating the setup for running the bookkeeper validation check
			err = e2eutil.InitialSetup(k8sClient, namespace)
			Expect(err).NotTo(HaveOccurred())
			authsettingsValidation := e2eutil.NewDefaultCluster(namespace)
			authsettingsValidation.WithDefaults()
			authsettingsValidation.Spec.Authentication.Enabled = true
			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "The field autoScale.controller.connect.security.auth.enable should be present")
			Expect(err.Error()).To(ContainSubstring("autoScale.controller.connect.security.auth.enable field is not present"))

			authsettingsValidation.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "false"
			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "The value for autoScale.controller.connect.security.auth.enable should not be false")
			Expect(err.Error()).To(ContainSubstring("autoScale.controller.connect.security.auth.enable should be set to true"))

			authsettingsValidation.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "dummy"
			authsettingsValidation.Spec.Pravega.Options["autoScale.authEnabled"] = "dummy"
			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "The value for autoScale.controller.connect.security.auth.enable/autoScale.authEnabled should not be incorrect")
			Expect(err.Error()).To(ContainSubstring("autoScale.controller.connect.security.auth.enable and autoScale.authEnabled should be set to true"))

			authsettingsValidation.Spec.Pravega.Options["autoScale.authEnabled"] = ""
			authsettingsValidation.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "true"
			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "Controller token sigining key should be present")
			Expect(err.Error()).To(ContainSubstring("controller.security.auth.delegationToken.signingKey.basis field is not present"))

			authsettingsValidation.Spec.Pravega.Options["controller.security.auth.delegationToken.signingKey.basis"] = "secret"
			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "Segmentstore token sigining key should be present")
			Expect(err.Error()).To(ContainSubstring("autoScale.security.auth.token.signingKey.basis field is not present"))

			authsettingsValidation.Spec.Pravega.Options["autoScale.security.auth.token.signingKey.basis"] = "secret1"
			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "Segmentstore and controller token sigining key should be same ")
			Expect(err.Error()).To(ContainSubstring("controller and segmentstore token signing key should have same value"))

			authsettingsValidation.Spec.Authentication.Enabled = false
			authsettingsValidation.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "true"
			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, authsettingsValidation)
			Expect(err).To(HaveOccurred(), "The field autoScale.controller.connect.security.auth.enable should not be set")
			Expect(err.Error()).To(ContainSubstring("autoScale.controller.connect.security.auth.enable/autoScale.authEnabled should not be set to true"))
		})
	})
})
