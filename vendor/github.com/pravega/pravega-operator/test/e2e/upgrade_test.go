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
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
)

func testUpgradeCluster(t *testing.T) {
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

	cluster := pravega_e2eutil.NewDefaultCluster(namespace)

	cluster.WithDefaults()
	// TODO: use the official 0.5 image when it is released
	initialVersion := "0.5.0-1"
	upgradeVersion := "0.5.0-2"
	cluster.Spec.Version = initialVersion
	cluster.Spec.Pravega.Image = &api.PravegaImageSpec{
		ImageSpec: api.ImageSpec{
			Repository: "adrianmo/pravega",
		},
	}
	cluster.Spec.Bookkeeper.Image = &api.BookkeeperImageSpec{
		ImageSpec: api.ImageSpec{
			Repository: "adrianmo/bookkeeper",
		},
	}

	pravega, err := pravega_e2eutil.CreateCluster(t, f, ctx, cluster)
	g.Expect(err).NotTo(HaveOccurred())

	// A default Pravega cluster should have 5 pods: 3 bookies, 1 controller, 1 segment store
	podSize := 5
	err = pravega_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, pravega, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Pravega cluster object
	pravega, err = pravega_e2eutil.GetCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(pravega.Status.CurrentVersion).To(Equal(initialVersion))

	pravega.Spec.Version = upgradeVersion

	err = pravega_e2eutil.UpdateCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WaitForClusterToUpgrade(t, f, ctx, pravega, upgradeVersion)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Pravega cluster object
	pravega, err = pravega_e2eutil.GetCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(pravega.Spec.Version).To(Equal(upgradeVersion))
	g.Expect(pravega.Status.CurrentVersion).To(Equal(upgradeVersion))
	g.Expect(pravega.Status.TargetVersion).To(Equal(""))

	// Delete cluster
	err = pravega_e2eutil.DeleteCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = pravega_e2eutil.WaitForClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// A workaround for issue 93
	err = pravega_e2eutil.RestartTier2(t, f, ctx, namespace)
	g.Expect(err).NotTo(HaveOccurred())
}
