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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	"github.com/pravega/pravega-operator/test/e2e/e2eutil"
)

// Test create and recreate a Pravega cluster with the same name
var _ = Describe("Create and recreate Pravega cluster with same name", func() {
	namespace := "default"
	defaultCluster := e2eutil.NewDefaultCluster(namespace)

	BeforeEach(func() {
		defaultCluster.WithDefaults()
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

			defaultCluster = e2eutil.NewDefaultCluster(namespace)
			defaultCluster.WithDefaults()

			defaultCluster.Spec.Pravega.ControllerSvcNameSuffix = "testcontroller"
			defaultCluster.Spec.Pravega.SegmentStoreHeadlessSvcNameSuffix = "testsegstore"
			defaultCluster.Spec.Pravega.SegmentStoreStsNameSuffix = "segsts"

			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			// A default Pravega cluster should have 2 pods: 1 controller, 1 segment store
			podCount = 2
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
			Eventually(e2eutil.WaitForPravegaClusterToBecomeReady(k8sClient, pravega, podCount), timeout).Should(Succeed())
		})
		It("should have a valid service created", func() {
			svcName := fmt.Sprintf("%s-testcontroller", pravega.Name)
			err = e2eutil.CheckServiceExists(k8sClient, pravega, svcName)
			Expect(err).NotTo(HaveOccurred())

			svcName = fmt.Sprintf("%s-testsegstore", pravega.Name)
			err = e2eutil.CheckServiceExists(k8sClient, pravega, svcName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should have a valid statefulset created", func() {
			stsName := fmt.Sprintf("%s-segsts", pravega.Name)
			err = e2eutil.CheckStsExists(k8sClient, pravega, stsName)
			Expect(err).NotTo(HaveOccurred())

			err = e2eutil.WriteAndReadData(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeletePravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForPravegaClusterToTerminate(k8sClient, pravega), timeout).Should(Succeed())
		})

		It("should recreate successfully", func() {
			// creating the setup for running the test
			err = e2eutil.InitialSetup(k8sClient, namespace)
			Expect(err).NotTo(HaveOccurred())

			defaultCluster = e2eutil.NewDefaultCluster(namespace)
			defaultCluster.WithDefaults()

			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForPravegaClusterToBecomeReady(k8sClient, pravega, podCount), timeout).Should(Succeed())
		})

		It("should have a valid service", func() {
			svcName := fmt.Sprintf("%s-pravega-controller", pravega.Name)
			err = e2eutil.CheckServiceExists(k8sClient, pravega, svcName)
			Expect(err).NotTo(HaveOccurred())

			svcName = fmt.Sprintf("%s-pravega-segmentstore-headless", pravega.Name)
			err = e2eutil.CheckServiceExists(k8sClient, pravega, svcName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should have a valid statefulset created", func() {
			stsName := fmt.Sprintf("%s-pravega-segment-store", pravega.Name)
			err = e2eutil.CheckStsExists(k8sClient, pravega, stsName)
			Expect(err).NotTo(HaveOccurred())

			err = e2eutil.WriteAndReadData(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

		})

		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeletePravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForPravegaClusterToTerminate(k8sClient, pravega), timeout).Should(Succeed())
		})
	})
})
