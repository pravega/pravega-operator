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
	"time"

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
			pravega *v1beta1.PravegaCluster
			err     error
		)

		It("should succeed", func() {
			// creating the setup for running the test
			err = e2eutil.InitialSetup(k8sClient, namespace)
			Expect(err).NotTo(HaveOccurred())

			pravega, err = e2eutil.CreatePravegaCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			// A default Pravega cluster should have 2 pods: 1 controller, 1 segment store
			podCount := 2
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
			Eventually(e2eutil.WaitForPravegaClusterToBecomeReady(k8sClient, pravega, podCount), timeout).Should(Succeed())
		})
		It("should recreate deleted pods", func() {
			// This is to get the latest Pravega cluster object
			pravega, err = e2eutil.GetPravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			podDeleteCount := 1
			err = e2eutil.DeletePods(k8sClient, pravega, podDeleteCount)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(60 * time.Second)
			Eventually(e2eutil.WaitForPravegaClusterToBecomeReady(k8sClient, pravega, podDeleteCount), timeout).Should(Succeed())
		})

		It("should tear down the cluster successfully", func() {
			err = e2eutil.DeletePravegaCluster(k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForPravegaClusterToTerminate(k8sClient, pravega), timeout).Should(Succeed())
		})
	})
})
