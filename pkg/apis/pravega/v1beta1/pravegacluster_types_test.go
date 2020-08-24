/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1beta1_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestV1beta1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PravegaCluster API")
}

var _ = Describe("PravegaCluster Types Spec", func() {

	var (
		p v1beta1.PravegaCluster
	)

	BeforeEach(func() {
		p = v1beta1.PravegaCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		}
	})

	Context("WithDefaults", func() {
		var changed bool

		BeforeEach(func() {
			changed = p.WithDefaults()
			p.Spec.ExternalAccess.Type = "LoadBalancer"
			p.Spec.ExternalAccess.DomainName = "example.com"
			p.WithDefaults()
		})

		It("should return as changed", func() {
			Ω(changed).Should(BeTrue())
		})

		It("should set zookeeper uri", func() {
			Ω(p.Spec.ZookeeperUri).Should(Equal("zookeeper-client:2181"))
		})

		It("should set external access", func() {
			Ω(p.Spec.ExternalAccess).ShouldNot(BeNil())
		})

		It("should set version to 0.7.0", func() {
			Ω(p.Spec.Version).Should(Equal("0.7.0"))
		})

		It("should set pravega spec", func() {
			Ω(p.Spec.Pravega).ShouldNot(BeNil())
		})

		It("should set bookkeeper uri", func() {
			Ω(p.Spec.BookkeeperUri).ShouldNot(BeNil())
		})
		It("IssecureController should return false", func() {
			Ω(p.Spec.TLS.IsSecureController()).To(Equal(false))
		})

		It("IsSecureSegmentStore should return false", func() {
			Ω(p.Spec.TLS.IsSecureSegmentStore()).To(Equal(false))
		})

		It("IsCaBundlePresent should return false", func() {
			Ω(p.Spec.TLS.IsCaBundlePresent()).To(Equal(false))
		})

		It("Autentication Enabled should return false", func() {
			Ω(p.Spec.Authentication.IsEnabled()).To(Equal(false))
		})

		It("should set external access type and domain name to empty", func() {
			Ω(string(p.Spec.ExternalAccess.Type)).Should(Equal(""))
			Ω(p.Spec.ExternalAccess.DomainName).Should(Equal(""))
		})
	})
	Context("Setting TLS and Autentication to nil", func() {
		BeforeEach(func() {
			p.Spec.Version = "0.6.0"
			p.WithDefaults()
			p.Spec.Authentication = nil
			p.Spec.TLS = nil
		})
		It("Autentication Enabled should return false", func() {
			Ω(p.Spec.Authentication.IsEnabled()).To(Equal(false))
		})
		It("IsCaBundlePresent should return false", func() {
			Ω(p.Spec.TLS.IsCaBundlePresent()).To(Equal(false))
		})
		It("IsSecureSegmentStore should return false", func() {
			Ω(p.Spec.TLS.IsSecureSegmentStore()).To(Equal(false))
		})
		It("IssecureController should return false", func() {
			Ω(p.Spec.TLS.IsSecureController()).To(Equal(false))
		})
	})
	Context("Checking various parameters", func() {
		BeforeEach(func() {
			p.Spec.Version = "0.6.0"
			p.WithDefaults()
		})
		It("should set volume claim template", func() {
			Ω(p.Spec.Pravega.CacheVolumeClaimTemplate).ShouldNot(BeNil())
		})

		name := p.StatefulSetNameForSegmentstore()
		It("Should return segmentstore sts name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.StatefulSetNameForSegmentstoreAbove07()
		It("Should return segmentstore sts name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.StatefulSetNameForSegmentstoreBelow07()
		It("Should return segmentstore sts name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.PravegaControllerServiceURL()
		It("Should return controller service url", func() {
			Ω(name).ShouldNot(BeNil())
		})

		labels := p.LabelsForController()
		It("Should return controller labels", func() {
			Ω(labels).ShouldNot(BeNil())
		})

		labels = p.LabelsForSegmentStore()
		It("Should return segmentstore labels", func() {
			Ω(labels).ShouldNot(BeNil())
		})

		labels = p.LabelsForPravegaCluster()
		It("Should return pravega labels", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.PdbNameForController()
		It("Event size should not be zero", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ConfigMapNameForController()
		It("Should return controller configmap name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ServiceNameForController()
		It("Should return controller service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ServiceNameForSegmentStore(0)
		It("Should return segmentstore service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ServiceNameForSegmentStoreBelow07(0)
		It("Should return segmentstore service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ServiceNameForSegmentStoreAbove07(0)
		It("Should return segmentstore service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.HeadlessServiceNameForSegmentStore()
		It("Should return segmentstore headless service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.HeadlessServiceNameForBookie()
		It("Should return bookie headless service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.DeploymentNameForController()
		It("Should return controller deployment name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.PdbNameForSegmentstore()
		It("Should return segmentstore pdb name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ConfigMapNameForSegmentstore()
		It("Should return segmentstore configmap name", func() {
			Ω(name).ShouldNot(BeNil())
		})

	})
	Context("checking event generation utility", func() {
		BeforeEach(func() {
			p.WithDefaults()
		})
		message := "upgrade failed"
		event := p.NewEvent("UPGRADE_ERROR", v1beta1.UpgradeErrorReason, message, "Error")
		It("Event size should not be zero", func() {
			Ω(event.Size()).ShouldNot(Equal(0))
		})
	})

	Context("WaitForClusterToTerminate", func() {
		var (
			client client.Client
			err    error
			p1     *v1beta1.PravegaCluster
		)

		BeforeEach(func() {
			p1 = &v1beta1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			p1.WithDefaults()
			s := scheme.Scheme
			s.AddKnownTypes(v1beta1.SchemeGroupVersion, p1)
			client = fake.NewFakeClient(p1)
			err = p1.WaitForClusterToTerminate(client)
		})
		It("should  be nil", func() {
			Ω(err).Should(BeNil())
		})
	})
})
