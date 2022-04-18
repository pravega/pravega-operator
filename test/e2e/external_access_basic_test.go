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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("External Access tests", func() {
	Context("Creating cluster with external access", func() {
		It("Should create pods successfully", func() {

			//creating the setup for running the test
			Expect(pravega_e2eutil.InitialSetup(&t, k8sClient, testNamespace)).NotTo(HaveOccurred())
			defaultCluster := pravega_e2eutil.NewDefaultCluster(testNamespace)
			defaultCluster.WithDefaults()

			pravega, err := pravega_e2eutil.CreatePravegaClusterForExternalAccess(&t, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(1 * time.Minute)

			// This is to get the latest Pravega cluster object
			pravega, err = pravega_e2eutil.GetPravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.CheckExternalAccesss(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.DeletePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToTerminate(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

		})
	})
})
