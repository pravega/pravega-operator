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
	pravega_e2eutil "github.com/pravega/pravega-operator/test/e2e/e2eutil"
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
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

	err = pravega_e2eutil.RunTestPod(t, f, ctx, pravega)
	if err != nil {
		return err
	}
	return nil
}

func testScaleUp(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, namespace string) error {
	pravega, err := pravega_e2eutil.CreateCluster(t, f, ctx, pravega_e2eutil.NewDefaultCluster(namespace))
	if err != nil {
		return err
	}

	err = pravega_e2eutil.WaitForPravegaCluster(t, f, ctx, pravega)
	if err != nil {
		return err
	}

	pravega.Spec.Pravega.ControllerReplicas = pravega.Spec.Pravega.ControllerReplicas + 1
	pravega.Spec.Pravega.SegmentStoreReplicas = pravega.Spec.Pravega.SegmentStoreReplicas + 1
	pravega.Spec.Bookkeeper.Replicas = pravega.Spec.Bookkeeper.Replicas + 1

	err = pravega_e2eutil.UpdateCluster(t, f, ctx, pravega)
	if err != nil {
		return err
	}

	err = pravega_e2eutil.WaitForPravegaCluster(t, f, ctx, pravega)
	if err != nil {
		return err
	}

	err = pravega_e2eutil.RunTestPod(t, f, ctx, pravega)
	if err != nil {
		return err
	}
	return nil
}

func testPvcWhenScalingDown(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, namespace string) error {
	pravega, err := pravega_e2eutil.CreateCluster(t, f, ctx, pravega_e2eutil.NewStandardCluster(namespace))
	if err != nil {
		return err
	}

	err = pravega_e2eutil.WaitForPravegaCluster(t, f, ctx, pravega)
	if err != nil {
		return err
	}

	pravega.Spec.Pravega.SegmentStoreReplicas = 1
	pravega.Spec.Bookkeeper.Replicas = 1

	err = pravega_e2eutil.UpdateCluster(t, f, ctx, pravega)
	if err != nil {
		return err
	}

	err = pravega_e2eutil.WaitForPravegaCluster(t, f, ctx, pravega)
	if err != nil {
		return err
	}

	err = pravega_e2eutil.WaitForPvc(t, f, ctx, pravega)
	if err != nil {
		return err
	}

	err = pravega_e2eutil.RunTestPod(t, f, ctx, pravega)
	if err != nil {
		return err
	}
	return nil
}
