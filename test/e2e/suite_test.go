/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (&the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package e2e

import (
	"context"

	bkapi "github.com/pravega/bookkeeper-operator/api/v1alpha1"
	api "github.com/pravega/pravega-operator/api/v1beta1"
	pravegav1beta1 "github.com/pravega/pravega-operator/api/v1beta1"
	pravegacontroller "github.com/pravega/pravega-operator/controllers"
	zkapi "github.com/pravega/zookeeper-operator/api/v1beta1"

	"os"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	cfg           *rest.Config
	k8sClient     client.Client // You'll be using this client in your tests.
	testEnv       *envtest.Environment
	ctx           context.Context
	cancel        context.CancelFunc
	testNamespace = "default"
	t             testing.T
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller e2e Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	enabled := true
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		Config:             cfg,
		UseExistingCluster: &enabled,
	}

	/*
		Then, we start the envtest cluster.
	*/
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = pravegav1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = bkapi.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = zkapi.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	/*
		After the schemas, you will see the following marker.
		This marker is what allows new schemas to be added here automatically when a new API is added to the project.
	*/

	//+kubebuilder:scaffold:scheme

	/*
		A client is created for our test CRUD operations.
	*/
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	if os.Getenv("RUN_LOCAL") == "true" {
		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:    scheme.Scheme,
			Namespace: testNamespace,
			Port:      9443,
			NewCache:  cache.MultiNamespacedCacheBuilder([]string{testNamespace}),
		})
		Expect(err).ToNot(HaveOccurred())

		err = (&pravegacontroller.PravegaClusterReconciler{
			Client: k8sManager.GetClient(),
			Scheme: k8sManager.GetScheme(),
		}).SetupWithManager(k8sManager)
		Expect(err).ToNot(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			err = k8sManager.Start(ctrl.SetupSignalHandler())
			Expect(err).ToNot(HaveOccurred(), "failed to run manager")
		}()
	}

}, 60)

/*
Kubebuilder also generates boilerplate functions for cleaning up envtest and actually running your test files in your controllers/ directory.
You won't need to touch these.
*/

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	bkList := &api.PravegaClusterList{}
	listOptions := []client.ListOption{
		client.InNamespace(testNamespace),
	}
	Expect(k8sClient.List(ctx, bkList, listOptions...)).NotTo(HaveOccurred())
	for _, bk := range bkList.Items {
		Expect(k8sClient.Delete(ctx, &bk)).NotTo(HaveOccurred())
	}
})
