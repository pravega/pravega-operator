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
		})

		It("should return as changed", func() {
			Ω(changed).Should(BeTrue())
		})

		It("should set zookeeper uri", func() {
			Ω(p.Spec.ZookeeperUri).Should(Equal("zk-client:2181"))
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
	})
})
