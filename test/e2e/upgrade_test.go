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
	if err != nil {
		t.Fatal(err)
	}
	f := framework.Global

	cluster := pravega_e2eutil.NewDefaultCluster(namespace)

	cluster.WithDefaults()
	// TODO: use the official images
	initialVersion := "0.5.0-1"
	upgradeVersion := "0.5.0-2"
	cluster.Spec.Version = initialVersion
	cluster.Spec.Pravega.ImageRepository = "adrianmo/pravega"
	cluster.Spec.Bookkeeper.ImageRepository = "adrianmo/bookkeeper"

	pravega, err := pravega_e2eutil.CreateCluster(t, f, ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	// A default Pravega cluster should have 5 pods: 3 bookies, 1 controller, 1 segment store
	podSize := 5
	err = pravega_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, pravega, podSize)
	if err != nil {
		t.Fatal(err)
	}

	// This is to get the latest Pravega cluster object
	pravega, err = pravega_e2eutil.GetCluster(t, f, ctx, pravega)
	if err != nil {
		t.Fatal(err)
	}

	g.Expect(pravega.Status.CurrentVersion).To(Equal(initialVersion))

	pravega.Spec.Version = upgradeVersion

	err = pravega_e2eutil.UpdateCluster(t, f, ctx, pravega)
	if err != nil {
		t.Fatal(err)
	}

	err = pravega_e2eutil.WaitForClusterToUpgrade(t, f, ctx, pravega, upgradeVersion)
	if err != nil {
		t.Fatal(err)
	}

	// This is to get the latest Pravega cluster object
	pravega, err = pravega_e2eutil.GetCluster(t, f, ctx, pravega)
	if err != nil {
		t.Fatal(err)
	}

	g.Expect(pravega.Spec.Version).To(Equal(upgradeVersion))
	g.Expect(pravega.Status.CurrentVersion).To(Equal(upgradeVersion))
	g.Expect(pravega.Status.TargetVersion).To(Equal(""))

	// Delete cluster
	err = pravega_e2eutil.DeleteCluster(t, f, ctx, pravega)
	if err != nil {
		t.Fatal(err)
	}

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = pravega_e2eutil.WaitForClusterToTerminate(t, f, ctx, pravega)
	if err != nil {
		t.Fatal(err)
	}

	// A workaround for issue 93
	err = pravega_e2eutil.RestartTier2(t, f, ctx, namespace)
	if err != nil {
		t.Fatal(err)
	}
}
