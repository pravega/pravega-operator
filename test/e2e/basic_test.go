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

func testCreateDefaultCluster(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, namespace string) error {
	pravega, err := pravega_e2eutil.CreateCluster(t, f, ctx, pravega_e2eutil.NewDefaultCluster(namespace))
	if err != nil {
		return err
	}

	err = pravega_e2eutil.WaitForPravegaCluster(t, f, ctx, pravega)
	if err != nil {
		return err
	}

	// TODO: Run test pod to write and read data from the cluster

	return nil
}
