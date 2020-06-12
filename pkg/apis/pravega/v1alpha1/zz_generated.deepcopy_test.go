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
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PravegaCluster DeepCopy", func() {
	Context("with defaults", func() {
		var (
			str1, str2 string
			str3, str4 v1.PullPolicy
			p1, p2     *v1alpha1.PravegaCluster
		)
		BeforeEach(func() {
			p1 = &v1alpha1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			p1.WithDefaults()
			p1.Status.Init()
			temp := *p1.DeepCopy()
			p2 = &temp
			str1 = p1.Spec.Pravega.Image.Repository
			str2 = p2.Spec.Pravega.Image.Repository
			p1.Spec.Pravega.Image.Repository = "pravega/exmple"
			p1.Spec.Pravega.DeepCopyInto(p2.Spec.Pravega)
			str3 = p1.Spec.Pravega.Image.PullPolicy
			str4 = p2.Spec.Pravega.Image.PullPolicy
			p1.Spec.Pravega.Image.PullPolicy = "PullIfNotPresent"
			p1.Spec.Pravega.Image.DeepCopyInto(p2.Spec.Pravega.Image)
			p2.Spec.Authentication = p1.Spec.Authentication.DeepCopy()
			p1.Spec.Authentication.Enabled = true
			temp1 := *p1.Spec.Authentication.DeepCopy()
			p2.Spec.Authentication = &temp1
			p1.Spec.Bookkeeper.Image.Repository = "pravega/bookie"
			p2.Spec.Bookkeeper.Image = p1.Spec.Bookkeeper.Image.DeepCopy()
			p1.Spec.Bookkeeper.BookkeeperJVMOptions.MemoryOpts = []string{"1g"}
			p2.Spec.Bookkeeper.BookkeeperJVMOptions = p1.Spec.Bookkeeper.BookkeeperJVMOptions.DeepCopy()
			p2.Spec.Bookkeeper.Storage = p1.Spec.Bookkeeper.Storage.DeepCopy()
			p1.Spec.Bookkeeper.Options["ledgers"] = "l1"
			p1.Spec.Bookkeeper.Replicas = 4
			p2.Spec.Bookkeeper = p1.Spec.Bookkeeper.DeepCopy()
			p1.Spec.Version = "0.4.0"
			p2.Spec = *p1.Spec.DeepCopy()
			p1.Status.SetPodsReadyConditionTrue()
			p2.Status.Conditions[0] = *p1.Status.Conditions[0].DeepCopy()
			p1.Status.VersionHistory = []string{"0.6.0", "0.5.0"}
			p2.Status = *p1.Status.DeepCopy()
			p1.Status.Members.Ready = []string{"bookie-0", "bookie-1"}
			p1.Status.Members.Unready = []string{"bookie-3", "bookie-2"}
			p2.Status.Members = *p1.Status.Members.DeepCopy()
			p1.Spec.ExternalAccess.DomainName = "example.com"
			p2.Spec.ExternalAccess = p1.Spec.ExternalAccess.DeepCopy()
			p1.Spec.TLS.Static.ControllerSecret = "controller-secret"
			p2.Spec.TLS = p1.Spec.TLS.DeepCopy()
			p1.Spec.TLS.Static.SegmentStoreSecret = "segmentstore-secret"
			p2.Spec.TLS.Static = p1.Spec.TLS.Static.DeepCopy()
			p1.Spec.Pravega.Image.Repository = "pravega/exmple"
			p2.Spec.Pravega.Image = p1.Spec.Pravega.Image.DeepCopy()
			p1.Spec.Pravega.Image.ImageSpec.Tag = "test"
			p2.Spec.Pravega.Image.ImageSpec = *p1.Spec.Pravega.Image.ImageSpec.DeepCopy()
			p1.Spec.Pravega.Tier2.FileSystem.PersistentVolumeClaim.ClaimName = "fs"
			p2.Spec.Pravega.Tier2.FileSystem = p1.Spec.Pravega.Tier2.FileSystem.DeepCopy()
			p1.Spec.Pravega.Options["key"] = "value"
			p1.Spec.Pravega.SegmentStoreServiceAnnotations["user"] = "test"
			p1.Spec.Pravega.ControllerServiceAnnotations["user"] = "test1"
			p2.Spec.Pravega = p1.Spec.Pravega.DeepCopy()
			p2.Spec.Pravega.Tier2 = p1.Spec.Pravega.Tier2.DeepCopy()
			p1.Spec.Pravega.Tier2 = &v1alpha1.Tier2Spec{
				Ecs: &v1alpha1.ECSSpec{
					ConfigUri:   "configUri",
					Bucket:      "bucket",
					Prefix:      "prefix",
					Credentials: "credentials",
				},
			}
			p2.Spec.Pravega.Tier2.Ecs = p1.Spec.Pravega.Tier2.Ecs.DeepCopy()
			p1.Spec.Pravega.Tier2 = &v1alpha1.Tier2Spec{
				FileSystem: &v1alpha1.FileSystemSpec{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "claim",
						ReadOnly:  true,
					},
				},
			}
			p2.Spec.Pravega.Tier2.FileSystem = p1.Spec.Pravega.Tier2.FileSystem.DeepCopy()

			p1.Spec.Pravega.Tier2 = &v1alpha1.Tier2Spec{
				Hdfs: &v1alpha1.HDFSSpec{
					Uri:               "uri",
					Root:              "root",
					ReplicationFactor: 1,
				},
			}
			p2.Spec.Pravega.Tier2.Hdfs = p1.Spec.Pravega.Tier2.Hdfs.DeepCopy()
		})
		It("value of str1 and str2 should be equal", func() {
			Ω(str2).To(Equal(str1))
		})
		It("value of str3 and str4 should be equal", func() {
			Ω(str3).To(Equal(str4))
		})
		It("Authentication enabled should be true", func() {
			Ω(p2.Spec.Authentication.Enabled).To(Equal(true))
		})
		It("bookie repository should match", func() {
			Ω(p2.Spec.Bookkeeper.Image.Repository).To(Equal("pravega/bookie"))
		})
		It("checking bookie jvm option as 1g", func() {
			Ω(p2.Spec.Bookkeeper.BookkeeperJVMOptions.MemoryOpts[0]).To(Equal("1g"))
		})
		It("checking bookie options ledger field", func() {
			Ω(p2.Spec.Bookkeeper.Options["ledgers"]).To(Equal("l1"))
		})
		It("checking bookie options ledger field", func() {
			Ω(p2.Spec.Bookkeeper.Options["ledgers"]).To(Equal("l1"))
		})
		It("checking bookiespec replicas", func() {
			Ω(p2.Spec.Bookkeeper.Replicas).To(Equal(int32(4)))
		})
		It("checking spec version", func() {
			Ω(p2.Spec.Version).To(Equal("0.4.0"))
		})
		It("checking status conditions", func() {
			Ω(p2.Status.Conditions[0].Reason).To(Equal(p1.Status.Conditions[0].Reason))
		})
		It("checking  spec storage", func() {
			Ω(p2.Spec.Bookkeeper.Storage).To(Equal(p1.Spec.Bookkeeper.Storage))
		})
		It("checking  version history", func() {
			Ω(p2.Status.VersionHistory[0]).To(Equal("0.6.0"))
		})
		It("checking ready members", func() {
			Ω(p2.Status.Members.Ready[0]).To(Equal("bookie-0"))
		})
		It("checking  unready members", func() {
			Ω(p2.Status.Members.Unready[0]).To(Equal("bookie-3"))
		})
		It("checking  external access domain name", func() {
			Ω(p2.Spec.ExternalAccess.DomainName).To(Equal("example.com"))
		})
		It("checking  controller secret", func() {
			Ω(p2.Spec.TLS.Static.ControllerSecret).To(Equal("controller-secret"))
		})
		It("checking  segmentstore secret", func() {
			Ω(p2.Spec.TLS.Static.SegmentStoreSecret).To(Equal("segmentstore-secret"))
		})
		It("checking  image  repository", func() {
			Ω(p2.Spec.Pravega.Image.Repository).To(Equal("pravega/exmple"))
		})
		It("checking  image  tag", func() {
			Ω(p2.Spec.Pravega.Image.ImageSpec.Tag).To(Equal("test"))
		})
		It("checking  pravega options", func() {
			Ω(p2.Spec.Pravega.Options["key"]).To(Equal("value"))
		})
		It("checking tier2 ECS", func() {
			Ω(p2.Spec.Pravega.Tier2.Ecs.ConfigUri).To(Equal("configUri"))
		})
		It("checking tier2 Hdfs", func() {
			Ω(p2.Spec.Pravega.Tier2.Hdfs.Uri).To(Equal("uri"))
		})
		It("checking tier2 Filesystem", func() {
			Ω(p2.Spec.Pravega.Tier2.FileSystem.PersistentVolumeClaim.ClaimName).To(Equal("claim"))
		})
		It("checking for nil authentication", func() {
			p1.Spec.Authentication = nil
			Ω(p1.Spec.Authentication.DeepCopy()).Should(BeNil())
		})
		It("checking for nil Bookkeeper", func() {
			p1.Spec.Bookkeeper = nil
			Ω(p1.Spec.Bookkeeper.DeepCopy()).Should(BeNil())
		})
		It("checking for nil TLS", func() {
			p1.Spec.TLS = nil
			Ω(p1.Spec.TLS.DeepCopy()).Should(BeNil())
		})
		It("checking for nil Pravega", func() {
			p1.Spec.Pravega = nil
			Ω(p1.Spec.Pravega.DeepCopy()).Should(BeNil())
		})
		It("checking for nil BookkeeperJVMOptions", func() {
			p1.Spec.Bookkeeper.BookkeeperJVMOptions = nil
			Ω(p1.Spec.Bookkeeper.BookkeeperJVMOptions.DeepCopy()).Should(BeNil())
		})
		It("checking for nil BookkeeperStorage", func() {
			p1.Spec.Bookkeeper.Storage = nil
			Ω(p1.Spec.Bookkeeper.Storage.DeepCopy()).Should(BeNil())
		})
		It("checking for nil Bookkeeper Image", func() {
			p1.Spec.Bookkeeper.Image = nil
			Ω(p1.Spec.Bookkeeper.Image.DeepCopy()).Should(BeNil())
		})
		It("checking for nil External access", func() {
			p1.Spec.ExternalAccess = nil
			Ω(p1.Spec.ExternalAccess.DeepCopy()).Should(BeNil())
		})
		It("checking for nil TLS Sttic", func() {
			p1.Spec.TLS.Static = nil
			Ω(p1.Spec.TLS.Static.DeepCopy()).Should(BeNil())
		})
		It("checking for nil Pravega Tier2", func() {
			p1.Spec.Pravega.Tier2 = nil
			Ω(p1.Spec.Pravega.Tier2.DeepCopy()).Should(BeNil())
		})
		It("checking for nil Pravega Image", func() {
			p1.Spec.Pravega.Image = nil
			Ω(p1.Spec.Pravega.Image.DeepCopy()).Should(BeNil())
		})
		It("checking for nil ECS", func() {
			p1.Spec.Pravega.Tier2.Ecs = nil
			Ω(p1.Spec.Pravega.Tier2.Ecs.DeepCopy()).Should(BeNil())
		})
		It("checking for nil Hdfs", func() {
			p1.Spec.Pravega.Tier2.Hdfs = nil
			Ω(p1.Spec.Pravega.Tier2.Hdfs.DeepCopy()).Should(BeNil())
		})
		It("checking for nil filesystem", func() {
			p1.Spec.Pravega.Tier2.FileSystem = nil
			Ω(p1.Spec.Pravega.Tier2.FileSystem.DeepCopy()).Should(BeNil())
		})
		It("checking for nil tier2", func() {
			p1.Spec.Pravega.Tier2 = nil
			Ω(p1.Spec.Pravega.Tier2.DeepCopy()).Should(BeNil())
		})
		It("checking for nil member status", func() {
			var memberstatus *v1alpha1.MembersStatus
			memberstatus2 := memberstatus.DeepCopy()
			Ω(memberstatus2).To(BeNil())
		})
		It("checking for nil cluster status", func() {
			var clusterstatus *v1alpha1.ClusterStatus
			clusterstatus2 := clusterstatus.DeepCopy()
			Ω(clusterstatus2).To(BeNil())
		})
		It("checking for nil cluster spec", func() {
			var clusterspec *v1alpha1.ClusterSpec
			clusterspec2 := clusterspec.DeepCopy()
			Ω(clusterspec2).To(BeNil())
		})
		It("checking for nil cluster condition", func() {
			var clustercond *v1alpha1.ClusterCondition
			clustercond2 := clustercond.DeepCopy()
			Ω(clustercond2).To(BeNil())
		})
		It("checking for nil pravega cluster", func() {
			var cluster *v1alpha1.PravegaCluster
			cluster2 := cluster.DeepCopy()
			Ω(cluster2).To(BeNil())
		})
		It("checking for nil imagespec", func() {
			var imagespec *v1alpha1.ImageSpec
			imagespec2 := imagespec.DeepCopy()
			Ω(imagespec2).To(BeNil())
		})
		It("checking for nil clusterlist", func() {
			var clusterlist *v1alpha1.PravegaClusterList
			clusterlist2 := clusterlist.DeepCopy()
			Ω(clusterlist2).To(BeNil())
		})
		It("checking for deepcopy for clusterlist", func() {
			var clusterlist v1alpha1.PravegaClusterList
			clusterlist.ResourceVersion = "v1alpha1"
			clusterlist2 := clusterlist.DeepCopy()
			Ω(clusterlist2.ResourceVersion).To(Equal("v1alpha1"))
		})
		It("checking for Deepcopy object", func() {
			pravega := p2.DeepCopyObject()
			Ω(pravega.GetObjectKind().GroupVersionKind().Version).To(Equal(""))
		})
		It("checking for nil pravega cluster deepcopyobject", func() {
			var cluster *v1alpha1.PravegaCluster
			cluster2 := cluster.DeepCopyObject()
			Ω(cluster2).To(BeNil())
		})
		It("checking for deepcopyobject for clusterlist", func() {
			var clusterlist v1alpha1.PravegaClusterList
			clusterlist.ResourceVersion = "v1alpha1"
			clusterlist2 := clusterlist.DeepCopyObject()
			Ω(clusterlist2).ShouldNot(BeNil())
		})
		It("checking for nil pravega clusterlist deepcopyobject", func() {
			var clusterlist *v1alpha1.PravegaClusterList
			clusterlist2 := clusterlist.DeepCopyObject()
			Ω(clusterlist2).To(BeNil())
		})
		It("checking for deepcopyobject for clusterlist with items", func() {
			var clusterlist v1alpha1.PravegaClusterList
			clusterlist.ResourceVersion = "v1alpha1"
			clusterlist.Items = []v1alpha1.PravegaCluster{
				{
					Spec: v1alpha1.ClusterSpec{},
				},
			}
			clusterlist2 := clusterlist.DeepCopyObject()
			Ω(clusterlist2).ShouldNot(BeNil())
		})
	})
})
