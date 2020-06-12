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

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	"github.com/pravega/pravega-operator/pkg/controller/pravega"

	pravegav1beta1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
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

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pravega cluster")
}

var _ = Describe("Pravega Cluster Version Sync", func() {
	const (
		Name      = "example"
		Namespace = "default"
	)

	var (
		s = scheme.Scheme
		r *ReconcilePravegaCluster
	)

	var _ = Describe("Upgrade Test", func() {
		var (
			req reconcile.Request
			p   *v1beta1.PravegaCluster
		)

		BeforeEach(func() {
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			p = &v1beta1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			p.Spec.Version = "0.5.0"
			s.AddKnownTypes(v1beta1.SchemeGroupVersion, p)
		})

		Context("Cluster condition prior to Upgrade", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				_, err = r.Reconcile(req)
			})

			Context("First reconcile", func() {
				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})
			})

			Context("Initial status", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
				)
				BeforeEach(func() {
					_, err = r.Reconcile(req)
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should have current version set to spec version", func() {
					Ω(foundPravega.Status.CurrentVersion).Should(Equal(foundPravega.Spec.Version))
				})

				It("should set upgrade condition and status to be false", func() {
					_, upgradeCondition := foundPravega.Status.GetClusterCondition(pravegav1beta1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Status).Should(Equal(corev1.ConditionFalse))
				})
			})
		})

		Context("Upgrade to new version", func() {
			var (
				client client.Client
			)

			BeforeEach(func() {
				p.Spec = v1beta1.ClusterSpec{
					Version: "0.5.0",
				}
				p.WithDefaults()
				p.Spec.ExternalAccess.Enabled = true
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				_, _ = r.Reconcile(req)
				foundPravega := &v1beta1.PravegaCluster{}
				_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				foundPravega.Spec.Version = "0.6.0"
				// bypass the pods ready check in the upgrade logic
				foundPravega.Status.SetPodsReadyConditionTrue()
				client.Update(context.TODO(), foundPravega)
				_, _ = r.Reconcile(req)
			})

			Context("Upgrading Condition", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
				)
				BeforeEach(func() {
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set target version to 0.6.0", func() {
					Ω(foundPravega.Status.TargetVersion).Should(Equal("0.6.0"))
				})

				It("should set upgrade condition to be true", func() {
					_, upgradeCondition := foundPravega.Status.GetClusterCondition(pravegav1beta1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Status).Should(Equal(corev1.ConditionTrue))
				})
			})

			Context("Upgrade Segmentstore", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
				)
				BeforeEach(func() {
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)

					_, _ = r.Reconcile(req)
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set upgrade condition reason to UpgradingSegmentstoreReason and message to 0", func() {
					_, upgradeCondition := foundPravega.Status.GetClusterCondition(pravegav1beta1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Reason).Should(Equal(pravegav1beta1.UpdatingSegmentstoreReason))
					Ω(upgradeCondition.Message).Should(Equal("0"))
				})
			})

			Context("Upgrade Controller", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
					sts          *appsv1.StatefulSet
				)
				BeforeEach(func() {
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)

					// Segmentstore
					sts = &appsv1.StatefulSet{}
					name := p.StatefulSetNameForSegmentstore()
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					targetImage, _ := foundPravega.PravegaTargetImage()
					sts.Spec.Template.Spec.Containers[0].Image = targetImage
					r.client.Update(context.TODO(), sts)

					_, _ = r.Reconcile(req)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)

					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set upgrade condition reason to UpgradingControllerReason and message to 0", func() {
					_, upgradeCondition := foundPravega.Status.GetClusterCondition(pravegav1beta1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Reason).Should(Equal(pravegav1beta1.UpdatingControllerReason))
					Ω(upgradeCondition.Message).Should(Equal("0"))
				})
			})
			Context("Upgrade Segmentstore to 0.7 from version below 0.7", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
				)
				BeforeEach(func() {

					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
					foundPravega.Spec.Version = "0.7.0"
					foundPravega.Spec.Pravega.SegmentStoreReplicas = 4
					foundPravega.Status.SetPodsReadyConditionTrue()
					client.Update(context.TODO(), foundPravega)
					_, _ = r.Reconcile(req)
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)

				})

				It("should set upgrade condition reason to UpgradingSegmentstoreReason and message to 0", func() {
					_, upgradeCondition := foundPravega.Status.GetClusterCondition(pravegav1beta1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Reason).Should(Equal(pravegav1beta1.UpdatingSegmentstoreReason))
					Ω(upgradeCondition.Message).Should(Equal("0"))
				})
			})
			Context("Upgrade Segmentstore to empty version", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
				)
				BeforeEach(func() {
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
					foundPravega.Status.TargetVersion = ""

					foundPravega.Status.SetPodsReadyConditionTrue()
					client.Update(context.TODO(), foundPravega)
					_, _ = r.Reconcile(req)
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)

				})
				It("should set upgrade condition reason and message to empty", func() {
					_, upgradeCondition := foundPravega.Status.GetClusterCondition(pravegav1beta1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Reason).Should(Equal(""))
					Ω(upgradeCondition.Message).Should(Equal(""))
				})
				It("should set the upgrade condition to false", func() {
					Ω(foundPravega.Status.IsClusterInUpgradingState()).To(Equal(false))
				})
			})
			Context("checking upgrade completion", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
				)
				BeforeEach(func() {
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
					foundPravega.Status.TargetVersion = "0.7.0"
					foundPravega.Status.SetUpgradingConditionTrue("", "")
					foundPravega.Status.SetPodsReadyConditionTrue()
					foundPravega.Status.CurrentVersion = "0.7.0"
					client.Update(context.TODO(), foundPravega)
					_, _ = r.Reconcile(req)
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set the target version to empty", func() {
					Ω(foundPravega.Status.TargetVersion).To(BeEquivalentTo(""))
				})
				It("should set the upgrade condition to false", func() {
					Ω(foundPravega.Status.IsClusterInUpgradingState()).To(Equal(false))
				})
			})
		})
	})

	var _ = Describe("Rollback Test", func() {
		var (
			req reconcile.Request
			p   *v1beta1.PravegaCluster
		)

		BeforeEach(func() {
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			p = &v1beta1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			p.Spec.Version = "0.5.0"
			s.AddKnownTypes(v1beta1.SchemeGroupVersion, p)
		})

		Context("Cluster Condition before Rollback", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				_, err = r.Reconcile(req)
			})

			Context("First reconcile", func() {
				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})
			})

			Context("Initial status", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
				)
				BeforeEach(func() {
					_, err = r.Reconcile(req)
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should have current version set to spec version", func() {
					Ω(foundPravega.Status.CurrentVersion).Should(Equal(foundPravega.Spec.Version))
				})

				It("should not have rollback condition set", func() {
					_, rollbackCondition := foundPravega.Status.GetClusterCondition(v1beta1.ClusterConditionRollback)
					Ω(rollbackCondition).Should(BeNil())
				})

				It("should have version history set", func() {
					history := foundPravega.Status.VersionHistory
					Ω(history[0]).Should(Equal("0.5.0"))
				})

			})
		})

		Context("Rollback to previous version", func() {
			var (
				client client.Client
			)

			BeforeEach(func() {
				p.Spec = v1beta1.ClusterSpec{
					Version: "0.6.0",
				}
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				_, _ = r.Reconcile(req)
				foundPravega := &v1beta1.PravegaCluster{}
				_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				foundPravega.Spec.Version = "0.5.0"
				foundPravega.Status.VersionHistory = []string{"0.5.0"}
				// bypass the pods ready check in the upgrade logic
				foundPravega.Status.SetPodsReadyConditionFalse()
				foundPravega.Status.SetErrorConditionTrue("UpgradeFailed", "some error")
				client.Update(context.TODO(), foundPravega)
				_, _ = r.Reconcile(req)
			})

			Context("Rollback Triggered", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
				)
				BeforeEach(func() {
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set Rollback condition status to be true", func() {
					_, rollbackCondition := foundPravega.Status.GetClusterCondition(v1beta1.ClusterConditionRollback)
					Ω(rollbackCondition.Status).To(Equal(corev1.ConditionTrue))
				})

				It("should set target version to previous version", func() {
					Ω(foundPravega.Status.TargetVersion).To(Equal(foundPravega.Spec.Version))
				})
			})

			Context("Rollback Controller", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
				)
				BeforeEach(func() {
					_, _ = r.Reconcile(req)
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set rollback condition reason to UpdatingController and message to 0", func() {
					_, rollbackCondition := foundPravega.Status.GetClusterCondition(pravegav1beta1.ClusterConditionRollback)
					Ω(rollbackCondition.Reason).Should(Equal(pravegav1beta1.UpdatingControllerReason))
					Ω(rollbackCondition.Message).Should(Equal("0"))
				})
			})

			Context("Rollback SegmentStore", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
				)
				BeforeEach(func() {
					_, _ = r.Reconcile(req)
					_, _ = r.Reconcile(req)
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set rollback condition reason to UpdatingSegmentStore and message to 0", func() {
					_, rollbackCondition := foundPravega.Status.GetClusterCondition(pravegav1beta1.ClusterConditionRollback)
					Ω(rollbackCondition.Reason).Should(Equal(pravegav1beta1.UpdatingSegmentstoreReason))
					Ω(rollbackCondition.Message).Should(Equal("0"))
				})
			})

			Context("Rollback Completed", func() {
				var (
					foundPravega *v1beta1.PravegaCluster
				)
				BeforeEach(func() {
					_, _ = r.Reconcile(req)
					_, _ = r.Reconcile(req)
					_, _ = r.Reconcile(req)
					_, _ = r.Reconcile(req)
					foundPravega = &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set currentversion equal to target version", func() {
					Ω(foundPravega.Status.CurrentVersion).Should(Equal("0.5.0"))
				})
				It("should set TargetVersoin to empty", func() {
					Ω(foundPravega.Status.TargetVersion).Should(Equal(""))
				})
				It("should set rollback condition to false", func() {
					_, rollbackCondition := foundPravega.Status.GetClusterCondition(pravegav1beta1.ClusterConditionRollback)
					Ω(rollbackCondition.Status).To(Equal(corev1.ConditionFalse))
				})
				It("should set error condition to false", func() {
					_, errorCondition := foundPravega.Status.GetClusterCondition(pravegav1beta1.ClusterConditionError)
					Ω(errorCondition.Status).To(Equal(corev1.ConditionFalse))
				})
			})
			Context("Rollback to version below 0.7 from above 0.7", func() {
				var (
					p1 *v1beta1.PravegaCluster
					p2 *v1beta1.PravegaCluster
				)
				BeforeEach(func() {
					p1 = &v1beta1.PravegaCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      Name,
							Namespace: Namespace,
						},
					}
					p1.Spec.Version = "0.6.0"
					s.AddKnownTypes(v1beta1.SchemeGroupVersion, p1)

					p2 = &v1beta1.PravegaCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      Name,
							Namespace: Namespace,
						},
					}
					p2.Spec.Version = "0.7.0"
					s.AddKnownTypes(v1beta1.SchemeGroupVersion, p2)
				})

				Context("Rollback SegmentStore to version below 0.7", func() {
					var (
						foundPravega *v1beta1.PravegaCluster
					)
					BeforeEach(func() {
						p1.WithDefaults()
						p2.WithDefaults()
						r = &ReconcilePravegaCluster{client: client, scheme: s}
						_, _ = r.Reconcile(req)
						foundPravega = &v1beta1.PravegaCluster{}
						_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
						foundPravega.Spec.Version = "0.6.0"
						foundPravega.Status.VersionHistory = []string{"0.6.0"}
						// bypass the pods ready check in the upgrade logic
						foundPravega.Status.SetPodsReadyConditionFalse()
						foundPravega.Status.SetErrorConditionTrue("UpgradeFailed", "some error")
						client.Update(context.TODO(), foundPravega)
						//creating the new sts so that rollback gets triggered as it needs both sts
						newsts := pravega.MakeSegmentStoreStatefulSet(p2)
						client.Create(context.TODO(), newsts)
						_, _ = r.Reconcile(req)
						_, _ = r.Reconcile(req)
						_, _ = r.Reconcile(req)
						foundPravega = &v1beta1.PravegaCluster{}
						_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
					})
					It("should set rollback condition reason to UpdatingSegmentStore and message to 0", func() {
						_, rollbackCondition := foundPravega.Status.GetClusterCondition(pravegav1beta1.ClusterConditionRollback)
						Ω(rollbackCondition.Reason).Should(Equal(pravegav1beta1.UpdatingSegmentstoreReason))
						Ω(rollbackCondition.Message).Should(Equal("0"))
					})
				})
			})
		})
	})
})
