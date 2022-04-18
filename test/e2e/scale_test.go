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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("Scale operation", func() {
	Context("Check Scaling of pods", func() {
		It("Scaling of pods should be successful", func() {

			//creating the setup for running the test
			Expect(pravega_e2eutil.InitialSetup(&t, k8sClient, testNamespace)).NotTo(HaveOccurred())
			defaultCluster := pravega_e2eutil.NewDefaultCluster(testNamespace)
			defaultCluster.WithDefaults()

			pravega, err := pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			// A default Pravega cluster should have 2 pods: 1 controller, 1 segment store
			podSize := 2
			err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(&t, k8sClient, pravega, podSize)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest Pravega cluster object
			pravega, err = pravega_e2eutil.GetPravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			// Scale up Pravega cluster, increase segment store size by 1
			pravega.Spec.Pravega.SegmentStoreReplicas = 2
			pravega.Spec.Pravega.ControllerReplicas = 2
			podSize = 4

			err = pravega_e2eutil.UpdatePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(&t, k8sClient, pravega, podSize)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest Pravega cluster object
			pravega, err = pravega_e2eutil.GetPravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			// Scale down Pravega cluster back to default
			pravega.Spec.Pravega.SegmentStoreReplicas = 1
			pravega.Spec.Pravega.ControllerReplicas = 1
			podSize = 2

			err = pravega_e2eutil.UpdatePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(&t, k8sClient, pravega, podSize)
			Expect(err).NotTo(HaveOccurred())

			// Delete cluster
			err = pravega_e2eutil.DeletePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToTerminate(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

		})
	})
})
