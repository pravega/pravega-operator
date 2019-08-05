/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package pravegacluster

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBookie(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pravega cluster")
}

var _ = Describe("PravegaCluster Controller", func() {
	const (
		Name      = "example"
		Namespace = "default"
	)

	var (
		s = scheme.Scheme
		r *ReconcilePravegaCluster
	)

	Context("Reconcile", func() {
		var (
			req reconcile.Request
			res reconcile.Result
			p   *v1alpha1.PravegaCluster
		)

		BeforeEach(func() {
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			p = &v1alpha1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			s.AddKnownTypes(v1alpha1.SchemeGroupVersion, p)
		})

		Context("Without spec", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				res, err = r.Reconcile(req)
			})

			It("shouldn't error", func() {
				Ω(err).Should(BeNil())
			})

			Context("Before defaults are applied", func() {
				It("should requeue the request", func() {
					Ω(res.Requeue).Should(BeTrue())
				})

				It("should set the default cluster spec options", func() {
					foundPravega := &v1alpha1.PravegaCluster{}
					err = client.Get(context.TODO(), req.NamespacedName, foundPravega)
					Ω(err).Should(BeNil())
					Ω(foundPravega.Spec.Version).Should(Equal("0.4.0"))
					Ω(foundPravega.Spec.ZookeeperUri).Should(Equal("zk-client:2181"))
					Ω(foundPravega.Spec.ExternalAccess).ShouldNot(BeNil())
					Ω(foundPravega.Spec.Pravega).ShouldNot(BeNil())
					Ω(foundPravega.Spec.Bookkeeper).ShouldNot(BeNil())
				})
			})

			Context("After defaults are applied", func() {

				BeforeEach(func() {
					// 2nd reconcile iteration
					res, err = r.Reconcile(req)
				})

				It("should requeue after ReconfileTime delay", func() {
					Ω(res.RequeueAfter).To(Equal(ReconcileTime))
				})

				Context("Bookkeeper", func() {
					It("should create a statefulset", func() {
						foundBk := &appsv1.StatefulSet{}
						nn := types.NamespacedName{
							Name:      util.StatefulSetNameForBookie(p.Name),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundBk)
						Ω(err).Should(BeNil())
					})

					It("should create a config-map", func() {
						foundCm := &corev1.ConfigMap{}
						nn := types.NamespacedName{
							Name:      util.ConfigMapNameForBookie(p.Name),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundCm)
						Ω(err).Should(BeNil())
					})

					It("should create a headless-service", func() {
						foundSvc := &corev1.Service{}
						nn := types.NamespacedName{
							Name:      util.HeadlessServiceNameForBookie(p.Name),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundSvc)
						Ω(err).Should(BeNil())
					})
				})

				Context("Controller", func() {
					It("should create a deployment", func() {
						foundController := &appsv1.Deployment{}
						nn := types.NamespacedName{
							Name:      util.DeploymentNameForController(p.Name),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundController)
						Ω(err).Should(BeNil())
					})

					It("should create a config-map", func() {
						foundCm := &corev1.ConfigMap{}
						nn := types.NamespacedName{
							Name:      util.ConfigMapNameForController(p.Name),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundCm)
						Ω(err).Should(BeNil())
					})

					It("should create a client-service", func() {
						foundSvc := &corev1.Service{}
						nn := types.NamespacedName{
							Name:      util.ServiceNameForController(p.Name),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundSvc)
						Ω(err).Should(BeNil())
					})
				})

				Context("SegmentStore", func() {
					It("should create a statefulset", func() {
						foundSS := &appsv1.StatefulSet{}
						nn := types.NamespacedName{
							Name:      util.StatefulSetNameForSegmentstore(p.Name),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundSS)
						Ω(err).Should(BeNil())
					})

					It("should create a config-map", func() {
						foundCm := &corev1.ConfigMap{}
						nn := types.NamespacedName{
							Name:      util.ConfigMapNameForSegmentstore(p.Name),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundCm)
						Ω(err).Should(BeNil())
					})

					It("should create a headless-service", func() {
						foundSvc := &corev1.Service{}
						nn := types.NamespacedName{
							Name:      util.HeadlessServiceNameForSegmentStore(p.Name),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundSvc)
						Ω(err).Should(BeNil())
					})

					It("should not create a client-services", func() {
						// By default, external access is not enabled, hence, there
						// should not be any client service
						foundSvc := &corev1.Service{}
						nn := types.NamespacedName{
							Name:      util.ServiceNameForSegmentStore(p.Name, 0),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundSvc)
						Ω(err).Should(MatchError("services \"example-pravega-segmentstore-0\" not found"))
					})
				})
			})
		})

		Context("Custom spec", func() {
			var (
				client    client.Client
				err       error
				customReq *corev1.ResourceRequirements
			)

			BeforeEach(func() {
				customReq = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("4"),
						corev1.ResourceMemory: resource.MustParse("6Gi"),
					},
				}
				p.Spec = v1alpha1.ClusterSpec{
					Version: "0.3.2-rc2",
					Bookkeeper: &v1alpha1.BookkeeperSpec{
						Replicas:  5,
						Resources: customReq,
						Image: &v1alpha1.BookkeeperImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Repository: "foo/bookkeeper",
							},
						},
					},
					Pravega: &v1alpha1.PravegaSpec{
						ControllerReplicas:    2,
						SegmentStoreReplicas:  4,
						ControllerResources:   customReq,
						SegmentStoreResources: customReq,
						Image: &v1alpha1.PravegaImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Repository: "bar/pravega",
							},
						},
					},
					TLS: &v1alpha1.TLSPolicy{
						Static: &v1alpha1.StaticTLS{
							ControllerSecret:   "controller-secret",
							SegmentStoreSecret: "segmentstore-secret",
						},
					},
				}
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				res, err = r.Reconcile(req)
			})

			It("shouldn't error", func() {
				Ω(err).Should(BeNil())
			})

			It("should requeue after ReconfileTime delay", func() {
				Ω(res.RequeueAfter).To(Equal(ReconcileTime))
			})

			Context("Cluster", func() {
				It("should have a custom version", func() {
					Ω(p.Spec.Version).Should(Equal("0.3.2-rc2"))
				})
			})

			Context("Bookkeeper", func() {
				var foundBk *appsv1.StatefulSet

				BeforeEach(func() {
					foundBk = &appsv1.StatefulSet{}
					nn := types.NamespacedName{
						Name:      util.StatefulSetNameForBookie(p.Name),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn, foundBk)
				})

				It("should create a bookie statefulset", func() {
					Ω(err).Should(BeNil())
				})

				It("should set number of replicas", func() {
					Ω(*foundBk.Spec.Replicas).Should(BeEquivalentTo(5))
				})

				It("should set container image", func() {
					Ω(foundBk.Spec.Template.Spec.Containers[0].Image).Should(Equal("foo/bookkeeper:0.3.2-rc2"))
				})

				It("should set pod resource requests and limits", func() {
					Ω(foundBk.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String()).Should(Equal("2"))
					Ω(foundBk.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()).Should(Equal("4Gi"))
					Ω(foundBk.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String()).Should(Equal("4"))
					Ω(foundBk.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()).Should(Equal("6Gi"))
				})
			})

			Context("Pravega Controller", func() {
				var foundController *appsv1.Deployment

				BeforeEach(func() {
					foundController = &appsv1.Deployment{}
					nn := types.NamespacedName{
						Name:      util.DeploymentNameForController(p.Name),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn, foundController)
				})

				It("should create a controller deployment", func() {
					Ω(err).Should(BeNil())
				})

				It("should set number of replicas", func() {
					Ω(*foundController.Spec.Replicas).Should(BeEquivalentTo(2))
				})

				It("should set container image", func() {
					Ω(foundController.Spec.Template.Spec.Containers[0].Image).Should(Equal("bar/pravega:0.3.2-rc2"))
				})

				It("should set pod resource requests and limits", func() {
					Ω(foundController.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String()).Should(Equal("2"))
					Ω(foundController.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()).Should(Equal("4Gi"))
					Ω(foundController.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String()).Should(Equal("4"))
					Ω(foundController.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()).Should(Equal("6Gi"))
				})

				It("should set secret volume", func() {
					Ω(foundController.Spec.Template.Spec.Volumes[0].Name).Should(Equal("heap-dump"))
					Ω(foundController.Spec.Template.Spec.Volumes[1].Name).Should(Equal("tls-secret"))
					Ω(foundController.Spec.Template.Spec.Volumes[1].VolumeSource.Secret.SecretName).Should(Equal("controller-secret"))
					Ω(foundController.Spec.Template.Spec.Containers[0].VolumeMounts[1].Name).Should(Equal("tls-secret"))
					Ω(foundController.Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath).Should(Equal("/etc/secret-volume"))
				})
			})

			Context("Pravega SegmentStore", func() {
				var foundSS *appsv1.StatefulSet

				BeforeEach(func() {
					foundSS = &appsv1.StatefulSet{}
					nn := types.NamespacedName{
						Name:      util.StatefulSetNameForSegmentstore(p.Name),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn, foundSS)
				})

				It("should create a segmentstore statefulset", func() {
					Ω(err).Should(BeNil())
				})

				It("should set number of replicas", func() {
					Ω(*foundSS.Spec.Replicas).Should(BeEquivalentTo(4))
				})

				It("should set container image", func() {
					Ω(foundSS.Spec.Template.Spec.Containers[0].Image).Should(Equal("bar/pravega:0.3.2-rc2"))
				})

				It("should set pod resource requests and limits", func() {
					Ω(foundSS.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String()).Should(Equal("2"))
					Ω(foundSS.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()).Should(Equal("4Gi"))
					Ω(foundSS.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String()).Should(Equal("4"))
					Ω(foundSS.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()).Should(Equal("6Gi"))
				})

				It("should set secret volume", func() {
					Ω(foundSS.Spec.Template.Spec.Volumes[0].Name).Should(Equal("heap-dump"))
					Ω(foundSS.Spec.Template.Spec.Volumes[1].Name).Should(Equal("tls-secret"))
					Ω(foundSS.Spec.Template.Spec.Volumes[1].VolumeSource.Secret.SecretName).Should(Equal("segmentstore-secret"))
					Ω(foundSS.Spec.Template.Spec.Containers[0].VolumeMounts[2].Name).Should(Equal("tls-secret"))
					Ω(foundSS.Spec.Template.Spec.Containers[0].VolumeMounts[2].MountPath).Should(Equal("/etc/secret-volume"))
				})
			})
		})
	})
})
