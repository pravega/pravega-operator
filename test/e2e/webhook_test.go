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
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
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

	/*
		//Test webhook with an unsupported Pravega cluster version
		pravega := &api.PravegaCluster{}
		pravega.Name = "pravega"
		pravega.Namespace = namespace
		pravega.WithDefaults()
		pravega.Spec.Version = "99.0.0"
		_, err = pravega_e2eutil.CreateCluster(t, f, ctx, pravega)
		g.Expect(err).To(HaveOccurred(), "failed to reject request with unsupported version")
		g.Expect(err.Error()).To(ContainSubstring("unsupported Pravega cluster version 99.0.0"))
	*/

	// Test webhook with a supported Pravega cluster version
	pravega := &api.PravegaCluster{}
	pravega.Name = "pravega"
	pravega.Namespace = namespace
	pravega.WithDefaults()
	pravega.Spec.Version = "0.5.0"
	pravega, err = pravega_e2eutil.CreateCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	podSize := 2
	err = pravega_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, pravega, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// Try to upgrade to a non-supported version
	pravega, err = pravega_e2eutil.GetCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	pravega.Spec.Version = "99.0.0"
	err = pravega_e2eutil.UpdateCluster(t, f, ctx, pravega)
	g.Expect(err).To(HaveOccurred(), "failed to reject request with unsupported version")
	g.Expect(err.Error()).To(ContainSubstring("unsupported Pravega cluster version 99.0.0"))

	// Delete cluster
	err = pravega_e2eutil.DeleteCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = pravega_e2eutil.WaitForClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())
}
