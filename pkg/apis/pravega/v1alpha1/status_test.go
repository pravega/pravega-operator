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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
)

func TestV1beta1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PravegaCluster Status")
}

var _ = Describe("PravegaCluster Status", func() {

	var p v1alpha1.PravegaCluster

	BeforeEach(func() {
		p = v1alpha1.PravegaCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		}
	})

	Context("manually set pods ready condition to be true", func() {
		BeforeEach(func() {
			condition := v1alpha1.ClusterCondition{
				Type:               v1alpha1.ClusterConditionPodsReady,
				Status:             corev1.ConditionTrue,
				Reason:             "",
				Message:            "",
				LastUpdateTime:     "",
				LastTransitionTime: "",
			}
			p.Status.Conditions = append(p.Status.Conditions, condition)
		})

		It("should contains pods ready condition and it is true status", func() {
			Ω(p.Status.ContainsCondition(v1alpha1.ClusterConditionPodsReady, corev1.ConditionTrue)).To(BeTrue())
		})
	})

	Context("set conditions", func() {
		Context("set pods ready condition to be true", func() {
			BeforeEach(func() {
				p.Status.SetPodsReadyConditionTrue()
			})
			It("should have pods ready condition with true status", func() {
				Ω(p.Status.ContainsCondition(v1alpha1.ClusterConditionPodsReady, corev1.ConditionTrue)).To(BeTrue())
			})
		})

		Context("set pod ready condition to be false", func() {
			BeforeEach(func() {
				p.Status.SetPodsReadyConditionFalse()
			})

			It("should have ready condition with false status", func() {
				Ω(p.Status.ContainsCondition(v1alpha1.ClusterConditionPodsReady, corev1.ConditionFalse)).To(BeTrue())
			})
		})

		Context("set scaling condition to be true", func() {
			BeforeEach(func() {
				p.Status.SetScalingConditionTrue()
			})

			It("should have scaling condition with true status", func() {
				Ω(p.Status.ContainsCondition(v1alpha1.ClusterConditionScaling, corev1.ConditionTrue)).To(BeTrue())
			})
		})

		Context("set scaling condition to be false", func() {
			BeforeEach(func() {
				p.Status.SetScalingConditionFalse()
			})

			It("should have ready condition with false status", func() {
				Ω(p.Status.ContainsCondition(v1alpha1.ClusterConditionScaling, corev1.ConditionFalse)).To(BeTrue())
			})
		})

		Context("set error condition to be true", func() {
			BeforeEach(func() {
				p.Status.SetErrorConditionTrue("")
			})

			It("should have error condition with true status", func() {
				Ω(p.Status.ContainsCondition(v1alpha1.ClusterConditionError, corev1.ConditionTrue)).To(BeTrue())
			})
		})

		Context("set error condition to be false", func() {
			BeforeEach(func() {
				p.Status.SetErrorConditionFalse()
			})

			It("should have ready condition with false status", func() {
				Ω(p.Status.ContainsCondition(v1alpha1.ClusterConditionError, corev1.ConditionFalse)).To(BeTrue())
			})
		})
	})

	Context("with defaults", func() {

		BeforeEach(func() {
			p.WithDefaults()
		})

		It("should set pods ready condition to be false", func() {
			Ω(p.Status.ContainsCondition(v1alpha1.ClusterConditionPodsReady, corev1.ConditionFalse)).To(BeTrue())
		})

		It("should set scaling condition to be false", func() {
			Ω(p.Status.ContainsCondition(v1alpha1.ClusterConditionScaling, corev1.ConditionFalse)).To(BeTrue())
		})

		It("should set error condition to be false", func() {
			Ω(p.Status.ContainsCondition(v1alpha1.ClusterConditionError, corev1.ConditionFalse)).To(BeTrue())
		})
	})
})
