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
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
)

func testScaleCluster(t *testing.T) {
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

	defaultCluster := pravega_e2eutil.NewDefaultCluster(namespace)
	defaultCluster.WithDefaults()

	pravega, err := pravega_e2eutil.CreatePravegaCluster(t, f, ctx, defaultCluster)
	g.Expect(err).NotTo(HaveOccurred())

	// A default Pravega cluster should have 2 pods: 1 controller, 1 segment store
	podSize := 2
	err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(t, f, ctx, pravega, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Pravega cluster object
	pravega, err = pravega_e2eutil.GetPravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// Scale up Pravega cluster, increase segment store size by 1
	pravega.Spec.Pravega.SegmentStoreReplicas = 2
	pravega.Spec.Pravega.ControllerReplicas = 2
	podSize = 4

	err = pravega_e2eutil.UpdatePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(t, f, ctx, pravega, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Pravega cluster object
	pravega, err = pravega_e2eutil.GetPravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// Scale down Pravega cluster back to default
	pravega.Spec.Pravega.SegmentStoreReplicas = 1
	pravega.Spec.Pravega.ControllerReplicas = 1
	podSize = 2

	err = pravega_e2eutil.UpdatePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(t, f, ctx, pravega, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// Delete cluster
	err = pravega_e2eutil.DeletePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = pravega_e2eutil.WaitForPravegaClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

}
