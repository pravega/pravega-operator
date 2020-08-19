/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
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
	bkapi "github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
	zkapi "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
)

// Test create and recreate a Pravega cluster with the same name
func testCreateRecreateCluster(t *testing.T) {
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

	b := &bkapi.BookkeeperCluster{}
	b.WithDefaults()
	b.Name = "bookkeeper"
	b.Namespace = "default"
	err = pravega_e2eutil.BKDeleteCluster(t, f, ctx, b)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WaitForBKClusterToTerminate(t, f, ctx, b)
	g.Expect(err).NotTo(HaveOccurred())

	z := &zkapi.ZookeeperCluster{}
	z.WithDefaults()
	z.Name = "zookeeper"
	z.Namespace = "default"

	err = pravega_e2eutil.ZKDeleteCluster(t, f, ctx, z)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WaitForZKClusterToTerminate(t, f, ctx, z)
	g.Expect(err).NotTo(HaveOccurred())

	z.WithDefaults()
	z.Spec.Persistence.VolumeReclaimPolicy = "Delete"
	z.Spec.Replicas = 1
	z, err = pravega_e2eutil.ZKCreateCluster(t, f, ctx, z)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WaitForZookeeperClusterToBecomeReady(t, f, ctx, z, 1)
	g.Expect(err).NotTo(HaveOccurred())

	b, err = pravega_e2eutil.BKCreateCluster(t, f, ctx, b)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WaitForBookkeeperClusterToBecomeReady(t, f, ctx, b, 3)
	g.Expect(err).NotTo(HaveOccurred())

	// A workaround for issue 93
	err = pravega_e2eutil.RestartTier2(t, f, ctx, namespace)
	g.Expect(err).NotTo(HaveOccurred())

	defaultCluster := pravega_e2eutil.NewDefaultCluster(namespace)
	defaultCluster.WithDefaults()

	pravega, err := pravega_e2eutil.CreateCluster(t, f, ctx, defaultCluster)
	g.Expect(err).NotTo(HaveOccurred())

	// A default Pravega cluster should have 2 pods: 1 controller, 1 segment store
	podSize := 2
	err = pravega_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, pravega, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WriteAndReadData(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.DeleteCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WaitForClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// A workaround for issue 93
	err = pravega_e2eutil.RestartTier2(t, f, ctx, namespace)
	g.Expect(err).NotTo(HaveOccurred())

	defaultCluster = pravega_e2eutil.NewDefaultCluster(namespace)
	defaultCluster.WithDefaults()

	pravega, err = pravega_e2eutil.CreateCluster(t, f, ctx, defaultCluster)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, pravega, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WriteAndReadData(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.DeleteCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = pravega_e2eutil.WaitForClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())
}
