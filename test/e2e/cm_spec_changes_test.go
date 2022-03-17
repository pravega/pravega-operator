/**
 * Copyright (c) 2018-2022 Dell Inc., or its subsidiaries. All Rights Reserved.
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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	"github.com/pravega/pravega-operator/test/e2e/e2eutil"
)

var _ = Describe("Create and recreate Pravega cluster with same name", func() {
	namespace := "default"
	defaultCluster := e2eutil.NewDefaultCluster(namespace)
	var jvmOptsController []string
	var jvmOptsSegmentStore []string

	BeforeEach(func() {
		defaultCluster.WithDefaults()
		jvmOptsController = []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50.0"}
		jvmOptsSegmentStore = append(defaultCluster.Spec.Pravega.SegmentStoreJVMOptions, "-XX:MaxRAMPercentage=50.0")

	})
	Context("When creating a new cluster", func() {
		var (
			pravega  *v1beta1.PravegaCluster
			err      error
			podCount int
		)

		It("should succeed", func() {
			// creating the setup for running the test
			err = e2eutil.InitialSetup(k8sClient, namespace)
			Expect(err).NotTo(HaveOccurred())

			defaultCluster.Spec.Pravega.Options["pravegaservice.container.count"] = "3"
			defaultCluster.Spec.Pravega.ControllerJvmOptions = jvmOptsController
			defaultCluster.Spec.Pravega.SegmentStoreJVMOptions = jvmOptsSegmentStore

			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			// A default Pravega cluster should have 2 pods:  1 controller, 1 segment store
			podCount = 2
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
			Eventually(e2eutil.WaitForPravegaClusterToBecomeReady(k8sClient, pravega, podCount), timeout).Should(Succeed())
		})

		It("should have configmap updated", func() {
			// This is to get the latest Pravega cluster object
			pravega, err = e2eutil.GetPravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			// Check configmap has correct values
			c_cm := pravega.ConfigMapNameForController()
			ss_cm := pravega.ConfigMapNameForSegmentstore()
			err = e2eutil.CheckConfigMapUpdated(k8sClient, pravega, c_cm, "JAVA_OPTS", jvmOptsController)
			Expect(err).NotTo(HaveOccurred())
			var jvmOptsSegmentStore []string
			jvmOptsSegmentStore = append(jvmOptsSegmentStore, "pravegaservice.service.listener.port=12345")
			err = e2eutil.CheckConfigMapUpdated(k8sClient, pravega, ss_cm, "JAVA_OPTS", jvmOptsSegmentStore)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update the pravega cluster successfully with new options", func() {
			// updating pravega options
			jvmOptsController = []string{"-XX:MaxDirectMemorySize=4g", "-XX:MaxRAMPercentage=60.0", "-XX:+UseContainerSupport"}
			jvmOptsSegmentStore := []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m", "-XX:MaxRAMPercentage=60.0", "-XX:+UseContainerSupport"}
			pravega.Spec.Pravega.ControllerJvmOptions = jvmOptsController
			pravega.Spec.Pravega.SegmentStoreJVMOptions = jvmOptsSegmentStore
			pravega.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "2"
			pravega.Spec.Pravega.Options["pravegaservice.service.listener.port"] = "443"
			pravega.Spec.Pravega.SegmentStoreServiceAccountName = "pravega-components"
			pravega.Spec.Pravega.ControllerServiceAccountName = "pravega-components"

			// updating pravegacluster
			err = e2eutil.UpdatePravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			// checking if the upgrade of option was successful
			Eventually(e2eutil.WaitForCMPravegaClusterToUpgrade(k8sClient, pravega), timeout).Should(Succeed())
		})

		It("should have the right resources after updating", func() {
			// This is to get the latest Pravega cluster object
			pravega, err = e2eutil.GetPravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			stsName := pravega.StatefulSetNameForSegmentstore()
			sts, err1 := e2eutil.GetSts(k8sClient, stsName)
			Expect(err1).NotTo(HaveOccurred())
			Expect(sts.Spec.Template.Spec.ServiceAccountName).To(Equal("pravega-components"))

			deployName := pravega.DeploymentNameForController()
			deploy, err2 := e2eutil.GetDeployment(k8sClient, deployName)
			Expect(err2).NotTo(HaveOccurred())
			Expect(deploy.Spec.Template.Spec.ServiceAccountName).To(Equal("pravega-components"))
			// Sleeping for 1 min before read/write data
			time.Sleep(60 * time.Second)

			// Check configmap is  Updated
			c_cm := pravega.ConfigMapNameForController()
			ss_cm := pravega.ConfigMapNameForSegmentstore()
			jvmOptsController = append(jvmOptsController, "bookkeeper.ack.quorum.size=2")
			err = e2eutil.CheckConfigMapUpdated(k8sClient, pravega, c_cm, "JAVA_OPTS", jvmOptsController)
			Expect(err).NotTo(HaveOccurred())
			jvmOptsSegmentStore = append(jvmOptsSegmentStore, "pravegaservice.service.listener.port=443")
			err = e2eutil.CheckConfigMapUpdated(k8sClient, pravega, ss_cm, "JAVA_OPTS", jvmOptsSegmentStore)
			Expect(err).NotTo(HaveOccurred())

			err = e2eutil.WriteAndReadData(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail to update the container count", func() {
			// updating pravega option
			pravega.Spec.Pravega.Options["pravegaservice.container.count"] = "10"

			// updating pravegacluster
			err = e2eutil.UpdatePravegaCluster(k8sClient, pravega)

			// should give an error
			Expect(strings.ContainsAny(err.Error(), "controller.container.count should not be changed")).To(Equal(true))
		})

		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeletePravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForPravegaClusterToTerminate(k8sClient, pravega), timeout).Should(Succeed())
		})
	})
})
