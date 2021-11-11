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
	g.Expect(err.Error()).To(ContainSubstring("Missing required value for field spec.pravega.segmentStoreResources.limits.memory"))

	// Test webhook with no value for segment store memory limits
	noMemoryLimits := pravega_e2eutil.NewClusterWithNoSegmentStoreMemoryLimits(namespace)
	noMemoryLimits.WithDefaults()
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, noMemoryLimits)
	g.Expect(err).To(HaveOccurred(), "Segment Store memory limits cannot be empty")
	g.Expect(err.Error()).To(ContainSubstring("Missing required value for field spec.pravega.segmentStoreResources.limits.memory"))

	// Test webhook with segment store memory requests being greater than memory limits
	memoryRequestsGreaterThanLimits := pravega_e2eutil.NewDefaultCluster(namespace)
	memoryRequestsGreaterThanLimits.WithDefaults()
	memoryRequestsGreaterThanLimits.Spec.Pravega.SegmentStoreResources.Requests[corev1.ResourceMemory] = resource.MustParse("5Gi")
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, memoryRequestsGreaterThanLimits)
	g.Expect(err).To(HaveOccurred(), "Segment Store memory requests should be less than or equal to limits")
	g.Expect(err.Error()).To(ContainSubstring("spec.pravega.segmentStoreResources.requests.memory value must be less than or equal to spec.pravega.segmentStoreResources.limits.memory"))

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
	g.Expect(err.Error()).To(ContainSubstring("MaxDirectMemorySize(2684354560 B) along with JVM Xmx value(2147483648 B) is greater than or equal to the total available memory(4294967296 B)!"))

	// Test Webhook with pravegaservice.cache.size.max being greater than MaxDirectMemorySize
	cacheSizeGreaterThanMaxDirectMemorySize := pravega_e2eutil.NewDefaultCluster(namespace)
	cacheSizeGreaterThanMaxDirectMemorySize.WithDefaults()
	cacheSizeGreaterThanMaxDirectMemorySize.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "3221225472"
	_, err = pravega_e2eutil.CreatePravegaCluster(t, f, ctx, cacheSizeGreaterThanMaxDirectMemorySize)
	g.Expect(err).To(HaveOccurred(), "cache size configured should be less than MaxDirectMemorySize")
	g.Expect(err.Error()).To(ContainSubstring("Cache size(3221225472 B) configured is greater than or equal to the JVM MaxDirectMemorySize(2684354560 B) value"))

	// Test webhook with a valid Pravega cluster version format
	validVersion := pravega_e2eutil.NewClusterWithVersion(namespace, "0.6.0")
	validVersion.WithDefaults()
	pravega, err := pravega_e2eutil.CreatePravegaCluster(t, f, ctx, validVersion)
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

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = pravega_e2eutil.WaitForPravegaClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

}
