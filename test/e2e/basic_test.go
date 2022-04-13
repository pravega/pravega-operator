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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pravega_e2eutil "github.com/pravega/pravega-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("Test create and recreate Pravega cluster with the same name", func() {
	Context("Check create/delete operations", func() {
		It("create and delete operations should be successful", func() {
			By("create Pravega cluster")
			Expect(pravega_e2eutil.InitialSetup(&t, k8sClient, testNamespace)).NotTo(HaveOccurred())
			defaultCluster := pravega_e2eutil.NewDefaultCluster(testNamespace)
			defaultCluster.WithDefaults()

			defaultCluster.Spec.Pravega.ControllerSvcNameSuffix = "testcontroller"
			defaultCluster.Spec.Pravega.SegmentStoreHeadlessSvcNameSuffix = "testsegstore"
			defaultCluster.Spec.Pravega.SegmentStoreStsNameSuffix = "segsts"

			pravega, err := pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())
			// A default Pravega cluster should have 2 pods: 1 controller, 1 segment store
			podSize := 2
			err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(&t, k8sClient, pravega, podSize)
			Expect(err).NotTo(HaveOccurred())

			svcName := fmt.Sprintf("%s-testcontroller", pravega.Name)
			err = pravega_e2eutil.CheckServiceExists(&t, k8sClient, pravega, svcName)
			Expect(err).NotTo(HaveOccurred())

			svcName = fmt.Sprintf("%s-testsegstore", pravega.Name)
			err = pravega_e2eutil.CheckServiceExists(&t, k8sClient, pravega, svcName)
			Expect(err).NotTo(HaveOccurred())

			stsName := fmt.Sprintf("%s-segsts", pravega.Name)
			err = pravega_e2eutil.CheckStsExists(&t, k8sClient, pravega, stsName)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WriteAndReadData(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.DeletePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToTerminate(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
			Expect(pravega_e2eutil.InitialSetup(&t, k8sClient, testNamespace)).NotTo(HaveOccurred())
			defaultCluster = pravega_e2eutil.NewDefaultCluster(testNamespace)
			defaultCluster.WithDefaults()

			pravega, err = pravega_e2eutil.CreatePravegaCluster(&t, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToBecomeReady(&t, k8sClient, pravega, podSize)
			Expect(err).NotTo(HaveOccurred())

			svcName = fmt.Sprintf("%s-pravega-controller", pravega.Name)
			err = pravega_e2eutil.CheckServiceExists(&t, k8sClient, pravega, svcName)
			Expect(err).NotTo(HaveOccurred())

			svcName = fmt.Sprintf("%s-pravega-segmentstore-headless", pravega.Name)
			err = pravega_e2eutil.CheckServiceExists(&t, k8sClient, pravega, svcName)
			Expect(err).NotTo(HaveOccurred())

			stsName = fmt.Sprintf("%s-pravega-segment-store", pravega.Name)
			err = pravega_e2eutil.CheckStsExists(&t, k8sClient, pravega, stsName)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WriteAndReadData(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.DeletePravegaCluster(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())

			err = pravega_e2eutil.WaitForPravegaClusterToTerminate(&t, k8sClient, pravega)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
