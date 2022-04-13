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

	api "github.com/pravega/pravega-operator/api/v1beta1"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("Rollback tests", func() {
	Context("Verify rollback of pods", func() {
		It("Upgrading of pods should be successful", func() {

			//creating the setup for running the test

			Expect(pravega_e2eutil.InitialSetup(&t, k8sClient, testNamespace)).NotTo(HaveOccurred())

			cluster := pravega_e2eutil.NewDefaultCluster(testNamespace)
			cluster.WithDefaults()
			initialVersion := "0.6.1"
			upgradeVersion := "0.7.0.xyz"
			cluster.Spec.Version = initialVersion
			cluster.Spec.Pravega.Image = &api.ImageSpec{
				Repository: "pravega/pravega",
				PullPolicy: "IfNotPresent",
			}

			pravega, err := pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, cluster)
			Expect(err).NotTo(HaveOccurred())

			// A default Pravega cluster should have 2 pods: 1 controller, 1 segment store
			podSize := 2
			err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(&t, k8sClient, pravega, podSize)
			Expect(err).NotTo(HaveOccurred())
			// This is to get the latest Pravega cluster object
			pravega, err = pravega_e2eutil.GetPravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			Expect(pravega.Status.CurrentVersion).To(Equal(initialVersion))

			pravega.Spec.Version = upgradeVersion

			err = pravega_e2eutil.UpdatePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToFailUpgrade(&t, k8sClient, pravega, upgradeVersion)
			Expect(err).NotTo(HaveOccurred())

			upgradeVersion = initialVersion

			// This is to get the latest Pravega cluster object
			pravega, err = pravega_e2eutil.GetPravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			pravega.Spec.Version = upgradeVersion

			err = pravega_e2eutil.UpdatePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToRollback(&t, k8sClient, pravega, upgradeVersion)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest Pravega cluster object
			pravega, err = pravega_e2eutil.GetPravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			Expect(pravega.Spec.Version).To(Equal(upgradeVersion))
			Expect(pravega.Status.CurrentVersion).To(Equal(upgradeVersion))
			Expect(pravega.Status.TargetVersion).To(Equal(""))

			upgradeVersion = "0.7.0"

			pravega.Spec.Version = upgradeVersion

			err = pravega_e2eutil.UpdatePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToUpgrade(&t, k8sClient, pravega, upgradeVersion)
			Expect(err).NotTo(HaveOccurred())

			// This is to get the latest Pravega cluster object
			pravega, err = pravega_e2eutil.GetPravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			Expect(pravega.Spec.Version).To(Equal(upgradeVersion))
			Expect(pravega.Status.CurrentVersion).To(Equal(upgradeVersion))
			Expect(pravega.Status.TargetVersion).To(Equal(""))

			// Delete cluster
			err = pravega_e2eutil.DeletePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToTerminate(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
