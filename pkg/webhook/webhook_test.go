/**
 * Copyright (c) 2019 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package webhook

import (
	"context"
	"testing"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook")
}

var _ = Describe("Admission webhook", func() {
	const (
		Name      = "example"
		Namespace = "default"
	)

	var (
		s = scheme.Scheme
	)

	Context("Version", func() {
		var (
			p   *v1alpha1.PravegaCluster
			pwh *pravegaWebhookHandler
		)

		BeforeEach(func() {
			p = &v1alpha1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			s.AddKnownTypes(v1alpha1.SchemeGroupVersion, p)
		})

		Context("Valid version", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				client = fake.NewFakeClient()
				pwh = &pravegaWebhookHandler{client: client}
			})

			Context("Standard version", func() {
				It("should pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "0.4",
					}
					err = pwh.validatePravegaManifest(context.TODO(), p)
					Ω(err).Should(BeNil())
				})
			})

			Context("Version with release candidate", func() {
				It("should pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "0.3.2-rc2",
					}
					err = pwh.validatePravegaManifest(context.TODO(), p)
					Ω(err).Should(BeNil())
				})
			})

			Context("Empty version field", func() {
				Context("Empty pravega tag field", func() {
					It("should pass", func() {
						p.Spec = v1alpha1.ClusterSpec{}
						err = pwh.validatePravegaManifest(context.TODO(), p)
						Ω(err).Should(BeNil())
					})
				})

				Context("Pravega tag field specified", func() {
					It("should pass", func() {
						p.Spec = v1alpha1.ClusterSpec{
							Pravega: &v1alpha1.PravegaSpec{
								Image: &v1alpha1.PravegaImageSpec{
									ImageSpec: v1alpha1.ImageSpec{
										Repository: "pravega/pravega",
										Tag:        "0.4.0",
									},
								},
							},
						}
						err = pwh.validatePravegaManifest(context.TODO(), p)
						Ω(err).Should(BeNil())
					})
				})
			})
		})

		Context("Invalid version", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				client = fake.NewFakeClient()
				pwh = &pravegaWebhookHandler{client: client}
			})

			Context("Version not compatible", func() {
				It("should not pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "1.0.0",
					}
					err = pwh.validatePravegaManifest(context.TODO(), p)
					Ω(err).ShouldNot(BeNil())
				})
			})

			Context("Version meaningless", func() {
				It("should not pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "hahahaha",
					}
					err := pwh.validatePravegaManifest(context.TODO(), p)
					Ω(err).ShouldNot(BeNil())
				})
			})
		})

		Context("Invalid upgrade version", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				p.Spec = v1alpha1.ClusterSpec{
					Version: "0.3.0",
				}
				client = fake.NewFakeClient(p)
				pwh = &pravegaWebhookHandler{client: client}
			})

			Context("Not in upgrade path", func() {
				It("should not pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "0.5.0",
					}
					err = pwh.validatePravegaManifest(context.TODO(), p)
					Ω(err).ShouldNot(BeNil())
				})
			})
		})
	})
})
