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

	framework "github.com/operator-framework/operator-sdk/pkg/test"

	pravega_e2eutil "github.com/pravega/pravega-operator/test/e2e/e2eutil"
)

func testCreateDefaultCluster(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	f := framework.Global

	pravega, err := pravega_e2eutil.CreateCluster(t, f, ctx, pravega_e2eutil.NewDefaultCluster(namespace))
	if err != nil {
		t.Fatal(err)
	}

	// A default Pravega cluster should have 5 pods: 3 bookies, 1 controller, 1 segment store
	podSize := 5
	err = pravega_e2eutil.WaitForPravegaCluster(t, f, ctx, pravega, podSize)
	if err != nil {
		t.Fatal(err)
	}

	err = pravega_e2eutil.WriteAndReadData(t, f, ctx, pravega)
	if err != nil {
		t.Fatal(err)
	}
}
