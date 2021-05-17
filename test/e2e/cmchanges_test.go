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
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
)

func testCMUpgradeCluster(t *testing.T) {
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

	cluster := pravega_e2eutil.NewDefaultCluster(namespace)

	cluster.WithDefaults()
	jvmOpts := []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50.0"}
	jvmOptions := strings.Join(jvmOpts, " ")
	cluster.Spec.Pravega.Options["pravegaservice.containerCount"] = "3"
	cluster.Spec.Pravega.ControllerJvmOptions = jvmOpts
	cluster.Spec.Pravega.SegmentStoreJVMOptions = jvmOpts

	pravega, err := pravega_e2eutil.CreatePravegaCluster(t, f, ctx, cluster)
	g.Expect(err).NotTo(HaveOccurred())

	// A default Pravega cluster should have 2 pods:  1 controller, 1 segment store
	podSize := 2
	err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(t, f, ctx, pravega, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Pravega cluster object
	pravega, err = pravega_e2eutil.GetPravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// Check configmap has correct values
	c_cm := pravega.ConfigMapNameForController()
	ss_cm := pravega.ConfigMapNameForSegmentstore()
	ss_val := "pravegaservice.service.listener.port=12345"
	err = pravega_e2eutil.CheckConfigMapUpdated(t, f, ctx, pravega, c_cm, "JAVA_OPTS", jvmOptions)
	g.Expect(err).NotTo(HaveOccurred())
	err = pravega_e2eutil.CheckConfigMapUpdated(t, f, ctx, pravega, ss_cm, "JAVA_OPTS", jvmOptions)
	g.Expect(err).NotTo(HaveOccurred())
	err = pravega_e2eutil.CheckConfigMapUpdated(t, f, ctx, pravega, ss_cm, "JAVA_OPTS", ss_val)
	g.Expect(err).NotTo(HaveOccurred())

	//updating pravega options
	jvmOpts = []string{"-XX:MaxDirectMemorySize=4g", "-XX:MaxRAMPercentage=60.0", "-XX:+UseContainerSupport"}
	jvmOptions = strings.Join(jvmOpts, " ")
	cluster.Spec.Pravega.ControllerJvmOptions = jvmOpts
	cluster.Spec.Pravega.SegmentStoreJVMOptions = jvmOpts
	pravega.Spec.Pravega.Options["bookkeeper.bkAckQuorumSize"] = "2"
	pravega.Spec.Pravega.Options["pravegaservice.service.listener.port"] = "443"

	//updating pravegacluster
	err = pravega_e2eutil.UpdatePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	//checking if the upgrade of option was successful
	err = pravega_e2eutil.WaitForCMPravegaClusterToUpgrade(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Pravega cluster object
	pravega, err = pravega_e2eutil.GetPravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// Check configmap is  Updated
	ss_val = "pravegaservice.service.listener.port=443"
	err = pravega_e2eutil.CheckConfigMapUpdated(t, f, ctx, pravega, c_cm, "JAVA_OPTS", jvmOptions)
	g.Expect(err).NotTo(HaveOccurred())
	err = pravega_e2eutil.CheckConfigMapUpdated(t, f, ctx, pravega, ss_cm, "JAVA_OPTS", jvmOptions)
	g.Expect(err).NotTo(HaveOccurred())
	err = pravega_e2eutil.CheckConfigMapUpdated(t, f, ctx, pravega, ss_cm, "JAVA_OPTS", ss_val)
	g.Expect(err).NotTo(HaveOccurred())

	// Sleeping for 1 min before read/write data
	time.Sleep(60 * time.Second)

	err = pravega_e2eutil.WriteAndReadData(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	//updating pravega option
	pravega.Spec.Pravega.Options["pravegaservice.containerCount"] = "10"

	//updating pravegacluster
	err = pravega_e2eutil.UpdatePravegaCluster(t, f, ctx, pravega)

	//should give an error
	g.Expect(strings.ContainsAny(err.Error(), "controller.containerCount should not be changed")).To(Equal(true))

	// Delete cluster
	err = pravega_e2eutil.DeletePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = pravega_e2eutil.WaitForPravegaClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

}
