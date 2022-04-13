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

var _ = Describe("Test create pravega cluster with TLS", func() {
	Context("Create cluster with TLS", func() {
		It("Cluster creation should be successful", func() {

			Expect(pravega_e2eutil.InitialSetup(&t, k8sClient, testNamespace)).NotTo(HaveOccurred())
			defaultCluster := pravega_e2eutil.NewDefaultCluster(testNamespace)
			defaultCluster.WithDefaults()

			pravega, err := pravega_e2eutil.CreatePravegaClusterWithTlsAuth(&t, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			// A default Pravega cluster should have 2 pods: 1 controller, 1 segment store
			podSize := 2
			err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(&t, k8sClient, pravega, podSize)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.DeletePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToTerminate(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
