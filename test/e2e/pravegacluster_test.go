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
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	apis "github.com/pravega/pravega-operator/pkg/apis"
	operator "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	pravega_e2eutil "github.com/pravega/pravega-operator/test/e2e/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPravegaCluster(t *testing.T) {
	pravegaClusterList := &operator.PravegaClusterList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PravegaCluster",
			APIVersion: "pravega.pravega.io/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, pravegaClusterList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("pravega-group", func(t *testing.T) {
		t.Run("Cluster", PravegaCluster)
	})
}

func PravegaCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: pravega_e2eutil.CleanupTimeout, RetryInterval: pravega_e2eutil.CleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for pravega-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "pravega-operator", 1, pravega_e2eutil.RetryInterval, pravega_e2eutil.Timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = testCreateDefaultCluster(t, f, ctx, namespace); err != nil {
		t.Fatal(err)
	}
}
