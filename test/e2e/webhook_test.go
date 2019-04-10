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
	"fmt"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
	"testing"
)

func testWebhook(t *testing.T) {
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

	//Test webhook with an unsupported Pravega cluster version
	invalidVersion := pravega_e2eutil.NewClusterWithVersion(namespace, "1.0.0")
	_, err = pravega_e2eutil.CreateCluster(t, f, ctx, invalidVersion)
	if err == nil {
		t.Fatal(fmt.Errorf("failed to reject request with unsupported version"))
	}

	// Test webhook with a supported Pravega cluster version
	validVersion := pravega_e2eutil.NewClusterWithVersion(namespace, "0.3.0")

	pravega, err := pravega_e2eutil.CreateCluster(t, f, ctx, validVersion)
	if err != nil {
		t.Fatal(err)
	}

	podSize := 5
	err = pravega_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, pravega, podSize)
	if err != nil {
		t.Fatal(err)
	}

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
