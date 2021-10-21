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
	jvmOptsController := []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50.0"}
	jvmOptsSegmentStore := append(cluster.Spec.Pravega.SegmentStoreJVMOptions, "-XX:MaxRAMPercentage=50.0")
	cluster.Spec.Pravega.Options["pravegaservice.container.count"] = "3"
	cluster.Spec.Pravega.ControllerJvmOptions = jvmOptsController
	cluster.Spec.Pravega.SegmentStoreJVMOptions = jvmOptsSegmentStore

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
	err = pravega_e2eutil.CheckConfigMapUpdated(t, f, ctx, pravega, c_cm, "JAVA_OPTS", jvmOptsController)
	g.Expect(err).NotTo(HaveOccurred())
	jvmOptsSegmentStore = append(jvmOptsSegmentStore, "pravegaservice.service.listener.port=12345")
	err = pravega_e2eutil.CheckConfigMapUpdated(t, f, ctx, pravega, ss_cm, "JAVA_OPTS", jvmOptsSegmentStore)
	g.Expect(err).NotTo(HaveOccurred())

	//updating pravega options
	jvmOptsController = []string{"-XX:MaxDirectMemorySize=4g", "-XX:MaxRAMPercentage=60.0", "-XX:+UseContainerSupport"}
	jvmOptsSegmentStore = []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m", "-XX:MaxRAMPercentage=60.0", "-XX:+UseContainerSupport"}
	pravega.Spec.Pravega.ControllerJvmOptions = jvmOptsController
	pravega.Spec.Pravega.SegmentStoreJVMOptions = jvmOptsSegmentStore
	pravega.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "2"
	pravega.Spec.Pravega.Options["pravegaservice.service.listener.port"] = "443"
	pravega.Spec.Pravega.SegmentStoreServiceAccountName = "pravega-components"
	pravega.Spec.Pravega.ControllerServiceAccountName = "pravega-components"

	//updating pravegacluster
	err = pravega_e2eutil.UpdatePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	//checking if the upgrade of option was successful
	err = pravega_e2eutil.WaitForCMPravegaClusterToUpgrade(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Pravega cluster object
	pravega, err = pravega_e2eutil.GetPravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	stsName := pravega.StatefulSetNameForSegmentstore()
	sts, err1 := pravega_e2eutil.GetSts(t, f, ctx, stsName)
	g.Expect(err1).NotTo(HaveOccurred())
	g.Expect(sts.Spec.Template.Spec.ServiceAccountName).To(Equal("pravega-components"))

	deployName := pravega.DeploymentNameForController()
	deploy, err2 := pravega_e2eutil.GetDeployment(t, f, ctx, deployName)
	g.Expect(err2).NotTo(HaveOccurred())
	g.Expect(deploy.Spec.Template.Spec.ServiceAccountName).To(Equal("pravega-components"))
	// Sleeping for 1 min before read/write data
	time.Sleep(60 * time.Second)

	// Check configmap is  Updated
	jvmOptsController = append(jvmOptsController, "bookkeeper.ack.quorum.size=2")
	err = pravega_e2eutil.CheckConfigMapUpdated(t, f, ctx, pravega, c_cm, "JAVA_OPTS", jvmOptsController)
	g.Expect(err).NotTo(HaveOccurred())
	jvmOptsSegmentStore = append(jvmOptsSegmentStore, "pravegaservice.service.listener.port=443")
	err = pravega_e2eutil.CheckConfigMapUpdated(t, f, ctx, pravega, ss_cm, "JAVA_OPTS", jvmOptsSegmentStore)
	g.Expect(err).NotTo(HaveOccurred())

	err = pravega_e2eutil.WriteAndReadData(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	//updating pravega option
	pravega.Spec.Pravega.Options["pravegaservice.container.count"] = "10"

	//updating pravegacluster
	err = pravega_e2eutil.UpdatePravegaCluster(t, f, ctx, pravega)

	//should give an error
	g.Expect(strings.ContainsAny(err.Error(), "controller.container.count should not be changed")).To(Equal(true))

	// Delete cluster
	err = pravega_e2eutil.DeletePravegaCluster(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = pravega_e2eutil.WaitForPravegaClusterToTerminate(t, f, ctx, pravega)
	g.Expect(err).NotTo(HaveOccurred())

}
