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
	"time"

	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
)

// Test create and recreate a Pravega cluster with the same name
func testExternalCreateRecreateCluster(t *testing.T) {
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

	pravega, err := pravega_e2eutil.CreatePravegaClusterForExternalAccess(t, f, ctx, defaultCluster)
	g.Expect(err).NotTo(HaveOccurred())

	time.Sleep(1 * time.Minute)

	// This is to get the latest Pravega cluster object
	pravega, err = pravega_e2eutil.GetPravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.CheckExternalAccesss(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.DeletePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = pravega_e2eutil.WaitForPravegaClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())
}
