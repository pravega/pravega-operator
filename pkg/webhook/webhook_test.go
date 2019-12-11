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

		Context("Mutate version", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				client = fake.NewFakeClient()
				pwh = &pravegaWebhookHandler{client: client}
			})
			Context("Version only in .spec", func() {
				BeforeEach(func() {
					p.Spec.Version = "0.3.2-rc3"
					err = pwh.mutatePravegaManifest(context.TODO(), p)
				})

				It("Shoud not have error", func() {
					Ω(err).Should(BeNil())
				})

				It("should use version to 0.3.2-rc3", func() {
					Ω(p.Spec.Version).Should(Equal("0.3.2-rc3"))
				})
			})

			Context("Version only in .spec.Pravega.Image.Tag", func() {

				BeforeEach(func() {
					p.Spec.Pravega = &v1alpha1.PravegaSpec{
						Image: &v1alpha1.PravegaImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Tag: "0.3.2-rc3",
							},
						},
					}
					err = pwh.mutatePravegaManifest(context.TODO(), p)
				})

				It("Shoud not have error", func() {
					Ω(err).Should(BeNil())
				})

				It("should set .spec.version to 0.3.2-rc3", func() {
					Ω(p.Spec.Version).Should(Equal("0.3.2-rc3"))
				})

				It("Image tags should be nil", func() {
					Ω(p.Spec.Pravega.Image.Tag).Should(Equal(""))
				})
			})

			Context("Version in .spec.Version and .spec.Pravega.Image.Tag", func() {

				BeforeEach(func() {
					p.Spec.Pravega = &v1alpha1.PravegaSpec{
						Image: &v1alpha1.PravegaImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Tag: "0.3.2-rc3",
							},
						},
					}
					p.Spec.Version = "0.1.0"
					err = pwh.mutatePravegaManifest(context.TODO(), p)
				})

				It("Shoud not have error", func() {
					Ω(err).Should(BeNil())
				})

				It("Version on .spec.Version should prevail", func() {
					Ω(p.Spec.Version).Should(Equal("0.1.0"))
					Ω(p.Spec.Pravega.Image.Tag).Should(Equal(""))
				})
			})
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
						Version: "0.4.0",
					}
					err = pwh.mutatePravegaManifest(context.TODO(), p)
					Ω(err).Should(BeNil())
				})
			})

			Context("Version with release candidate", func() {
				It("should pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "0.3.2-rc2",
					}
					err = pwh.mutatePravegaManifest(context.TODO(), p)
					Ω(err).Should(BeNil())
				})
			})

			Context("Empty version field", func() {
				Context("Empty pravega tag field", func() {
					It("should pass", func() {
						p.Spec = v1alpha1.ClusterSpec{}
						err = pwh.mutatePravegaManifest(context.TODO(), p)
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
						err = pwh.mutatePravegaManifest(context.TODO(), p)
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
						Version: "99.0.0",
					}
					err = pwh.mutatePravegaManifest(context.TODO(), p)
					Ω(err).ShouldNot(BeNil())
					Ω(err.Error()).To(Equal("unsupported Pravega cluster version 99.0.0"))
				})
			})

			Context("Version meaningless", func() {
				It("should not pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "hahahaha",
					}
					err := pwh.mutatePravegaManifest(context.TODO(), p)
					Ω(err).ShouldNot(BeNil())
					Ω(err.Error()).To(Equal("request version is not in valid format: failed to parse version hahahaha"))
				})
			})
		})

		Context("Upgrade version", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				p.Spec = v1alpha1.ClusterSpec{
					Version: "0.5.0-001",
				}
				client = fake.NewFakeClient(p)
				pwh = &pravegaWebhookHandler{client: client}
			})

			Context("Supported and in upgrade path", func() {
				It("should pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "0.5.0-002",
					}
					err = pwh.mutatePravegaManifest(context.TODO(), p)
					Ω(err).Should(BeNil())
				})
			})

			Context("Not supported", func() {
				It("should not pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "99.0.0-001",
					}
					err = pwh.mutatePravegaManifest(context.TODO(), p)
					Ω(err).ShouldNot(BeNil())
					Ω(err.Error()).To(Equal("unsupported Pravega cluster version 99.0.0-001"))
				})
			})

			Context("Not in upgrade path", func() {
				It("should not pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "0.4.0-001",
					}
					err = pwh.mutatePravegaManifest(context.TODO(), p)
					Ω(err).ShouldNot(BeNil())
					Ω(err.Error()).To(Equal("unsupported upgrade from version 0.5.0-001 to 0.4.0-001"))
				})
			})
		})

		Context("Reject request when upgrading", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				p.Spec = v1alpha1.ClusterSpec{
					Version: "0.5.0-001",
				}
				p.Status.SetUpgradingConditionTrue("", "")
				client = fake.NewFakeClient(p)
				pwh = &pravegaWebhookHandler{client: client}
			})

			Context("Sending request when upgrading", func() {
				It("should not pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "0.5.0-002",
					}
					err = pwh.clusterIsAvailable(context.TODO(), p)
					Ω(err).ShouldNot(BeNil())
				})
			})
		})

		Context("Reject request when rolling back", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				p.Spec = v1alpha1.ClusterSpec{
					Version: "0.5.0-001",
				}
				p.Status.SetPodsReadyConditionFalse()
				p.Status.SetRollbackConditionTrue("", "")
				client = fake.NewFakeClient(p)
				pwh = &pravegaWebhookHandler{client: client}
			})

			Context("Sending request when rolling back", func() {
				It("should not pass", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "0.5.0-002",
					}
					err = pwh.clusterIsAvailable(context.TODO(), p)
					Ω(err).Should(MatchError("failed to process the request, cluster is in rollback"))
				})
			})
		})

		Context("Rollback version", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				p.Spec = v1alpha1.ClusterSpec{
					Version: "0.5.0-002",
				}
				p.Status.CurrentVersion = "0.5.0-001"
				p.Status.Init()
				p.Status.SetErrorConditionTrue("UpgradeFailed", "some error message")

				client = fake.NewFakeClient(p)
				pwh = &pravegaWebhookHandler{client: client}
			})

			Context("Sending request when upgrade failed", func() {
				It("should not pass if version is different from previous stable version", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "0.5.0-003",
					}
					err = pwh.clusterIsAvailable(context.TODO(), p)
					Ω(err).Should(BeNil())
					err = pwh.mutatePravegaManifest(context.TODO(), p)
					Ω(err).Should(MatchError("Rollback to version 0.5.0-003 not supported. Only rollback to version 0.5.0-001 is supported."))
				})

				It("should pass if version is same as previous stable version", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "0.5.0-001",
					}
					err = pwh.clusterIsAvailable(context.TODO(), p)
					Ω(err).Should(BeNil())
					err = pwh.mutatePravegaManifest(context.TODO(), p)
					Ω(err).Should(BeNil())
				})
			})
		})

		Context("Version edit when cluster in Error state ", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				p.Spec = v1alpha1.ClusterSpec{
					Version: "0.5.0-002",
				}
				p.Status.CurrentVersion = "0.5.0-001"
				p.Status.Init()
				p.Status.SetErrorConditionTrue("Some strange reason", "some error message")

				client = fake.NewFakeClient(p)
				pwh = &pravegaWebhookHandler{client: client}
			})

			Context("Sending request when cluster in error state", func() {
				It("should not pass if cluster is in error state", func() {
					p.Spec = v1alpha1.ClusterSpec{
						Version: "0.5.0-033",
					}
					err = pwh.clusterIsAvailable(context.TODO(), p)
					Ω(err).Should(MatchError("failed to process the request, cluster is in error state."))
				})
			})
		})
	})
})
