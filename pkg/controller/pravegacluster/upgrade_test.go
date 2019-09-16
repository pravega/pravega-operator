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

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/util"

	pravegav1alpha1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
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

var _ = Describe("Pravega Cluster", func() {
	const (
		Name      = "example"
		Namespace = "default"
	)

	var (
		s = scheme.Scheme
		r *ReconcilePravegaCluster
	)

	Context("Upgrade", func() {
		var (
			req reconcile.Request
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
			p.Spec.Version = "0.5.0"
			s.AddKnownTypes(v1alpha1.SchemeGroupVersion, p)
		})

		Context("Pravega condition", func() {
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
					foundPravega *v1alpha1.PravegaCluster
				)
				BeforeEach(func() {
					_, err = r.Reconcile(req)
					foundPravega = &v1alpha1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should have current version set to spec version", func() {
					Ω(foundPravega.Status.CurrentVersion).Should(Equal(foundPravega.Spec.Version))
				})

				It("should set upgrade condition and status to be false", func() {
					_, upgradeCondition := foundPravega.Status.GetClusterCondition(pravegav1alpha1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Status).Should(Equal(corev1.ConditionFalse))
				})
			})
		})

		Context("Upgrade to new version", func() {
			var (
				client client.Client
			)

			BeforeEach(func() {
				p.Spec = v1alpha1.ClusterSpec{
					Version: "0.5.0",
				}
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				_, _ = r.Reconcile(req)
				foundPravega := &v1alpha1.PravegaCluster{}
				_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				foundPravega.Spec.Version = "0.6.0"
				// bypass the pods ready check in the upgrade logic
				foundPravega.Status.SetPodsReadyConditionTrue()
				client.Update(context.TODO(), foundPravega)
				_, _ = r.Reconcile(req)
			})

			Context("Condition", func() {
				var (
					foundPravega *v1alpha1.PravegaCluster
				)
				BeforeEach(func() {
					foundPravega = &v1alpha1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set target version to 0.6.0", func() {
					Ω(foundPravega.Status.TargetVersion).Should(Equal("0.6.0"))
				})

				It("should set upgrade condition to be true", func() {
					_, upgradeCondition := foundPravega.Status.GetClusterCondition(pravegav1alpha1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Status).Should(Equal(corev1.ConditionTrue))
				})
			})

			Context("Upgrade Bookkeeper", func() {
				var (
					foundPravega *v1alpha1.PravegaCluster
					sts          *appsv1.StatefulSet
				)
				BeforeEach(func() {
					sts = &appsv1.StatefulSet{}
					name := util.StatefulSetNameForBookie(p.Name)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					sts.Status.ReadyReplicas = 1
					r.client.Update(context.TODO(), sts)

					_, _ = r.Reconcile(req)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					foundPravega = &v1alpha1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set upgrade condition reason to UpgradingBookkeeperReason and message to 0", func() {
					_, upgradeCondition := foundPravega.Status.GetClusterCondition(pravegav1alpha1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Reason).Should(Equal(pravegav1alpha1.UpdatingBookkeeperReason))
					Ω(upgradeCondition.Message).Should(Equal("0"))
				})
			})

			Context("Upgrade Segmentstore", func() {
				var (
					foundPravega *v1alpha1.PravegaCluster
					sts          *appsv1.StatefulSet
				)
				BeforeEach(func() {
					foundPravega = &v1alpha1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)

					// Bookkeeper
					sts = &appsv1.StatefulSet{}
					name := util.StatefulSetNameForBookie(p.Name)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					targetImage, _ := util.BookkeeperTargetImage(foundPravega)
					sts.Spec.Template.Spec.Containers[0].Image = targetImage
					r.client.Update(context.TODO(), sts)

					_, _ = r.Reconcile(req)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					foundPravega = &v1alpha1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set upgrade condition reason to UpgradingSegmentstoreReason and message to 0", func() {
					_, upgradeCondition := foundPravega.Status.GetClusterCondition(pravegav1alpha1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Reason).Should(Equal(pravegav1alpha1.UpdatingSegmentstoreReason))
					Ω(upgradeCondition.Message).Should(Equal("0"))
				})
			})

			Context("Upgrade Controller", func() {
				var (
					foundPravega *v1alpha1.PravegaCluster
					sts          *appsv1.StatefulSet
				)
				BeforeEach(func() {
					foundPravega = &v1alpha1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)

					// Bookkeeper
					sts = &appsv1.StatefulSet{}
					name := util.StatefulSetNameForBookie(p.Name)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					targetImage, _ := util.BookkeeperTargetImage(foundPravega)
					sts.Spec.Template.Spec.Containers[0].Image = targetImage
					r.client.Update(context.TODO(), sts)

					// Segmentstore
					name = util.StatefulSetNameForSegmentstore(p.Name)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					targetImage, _ = util.PravegaTargetImage(foundPravega)
					sts.Spec.Template.Spec.Containers[0].Image = targetImage
					r.client.Update(context.TODO(), sts)

					_, _ = r.Reconcile(req)
					_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)
					foundPravega = &v1alpha1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("should set upgrade condition reason to UpgradingControllerReason and message to 0", func() {
					_, upgradeCondition := foundPravega.Status.GetClusterCondition(pravegav1alpha1.ClusterConditionUpgrading)
					Ω(upgradeCondition.Reason).Should(Equal(pravegav1alpha1.UpdatingControllerReason))
					Ω(upgradeCondition.Message).Should(Equal("0"))
				})
			})
		})
	})
})
