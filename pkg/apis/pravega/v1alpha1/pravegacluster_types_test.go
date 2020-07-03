/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1alpha1_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
)

func TestV1alpha1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PravegaCluster API")
}

var _ = Describe("PravegaCluster Types Spec", func() {

	var p v1alpha1.PravegaCluster

	BeforeEach(func() {
		p = v1alpha1.PravegaCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		}
	})

	Context("#WithDefaults", func() {
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
		It("should set version to 0.4.0", func() {
			Ω(p.Spec.Version).Should(Equal("0.4.0"))
		})
		It("should set pravega spec", func() {
			Ω(p.Spec.Pravega).ShouldNot(BeNil())
		})
		It("should set bookkeeper spec", func() {
			Ω(p.Spec.Bookkeeper).ShouldNot(BeNil())
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
		It("IsEnabled should return false", func() {
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
		It("IsEnabled should return false", func() {
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
})
