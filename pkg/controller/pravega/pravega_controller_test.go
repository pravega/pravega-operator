/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package pravega_test

import (
	"testing"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/controller/pravega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pravega")
}

var _ = Describe("Controller", func() {

	var _ = Describe("Controller Test", func() {
		var (
			p *v1alpha1.PravegaCluster
		)

		BeforeEach(func() {
			p = &v1alpha1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			p.Spec.Version = "0.5.0"
		})

		Context("Empty Controller Service Type", func() {
			var (
				customReq *corev1.ResourceRequirements
				err       error
			)

			BeforeEach(func() {
				annotationsMap := map[string]string{
					"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
				}
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
					Version:      "0.5.0",
					ZookeeperUri: "example.com",
					ExternalAccess: &v1alpha1.ExternalAccess{
						Enabled:    true,
						Type:       corev1.ServiceTypeClusterIP,
						DomainName: "pravega.com.",
					},
					Bookkeeper: &v1alpha1.BookkeeperSpec{
						Replicas:  5,
						Resources: customReq,
						Image: &v1alpha1.BookkeeperImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Repository: "foo/bookkeeper",
							},
						},
						BookkeeperJVMOptions: &v1alpha1.BookkeeperJVMOptions{
							MemoryOpts:    []string{"-Xms2g", "-XX:MaxDirectMemorySize=2g"},
							GcOpts:        []string{"-XX:MaxGCPauseMillis=20", "-XX:-UseG1GC"},
							GcLoggingOpts: []string{"-XX:NumberOfGCLogFiles=10"},
						},
					},
					Pravega: &v1alpha1.PravegaSpec{
						ControllerReplicas:             2,
						SegmentStoreReplicas:           4,
						ControllerServiceAccountName:   "pravega-components",
						SegmentStoreServiceAccountName: "pravega-components",
						ControllerResources:            customReq,
						SegmentStoreResources:          customReq,
						ControllerServiceAnnotations:   annotationsMap,
						SegmentStoreServiceAnnotations: annotationsMap,
						Image: &v1alpha1.PravegaImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Repository: "bar/pravega",
							},
						},
						ControllerJvmOptions:   []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMFraction=1"},
						SegmentStoreJVMOptions: []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMFraction=1"},
						Options: map[string]string{
							"dummy-key": "dummy-value",
						},
						CacheVolumeClaimTemplate: &corev1.PersistentVolumeClaimSpec{
							VolumeName: "abc",
						},
					},
					TLS: &v1alpha1.TLSPolicy{
						Static: &v1alpha1.StaticTLS{
							ControllerSecret:   "controller-secret",
							SegmentStoreSecret: "segmentstore-secret",
						},
					},
					Authentication: &v1alpha1.AuthenticationParameters{
						Enabled:            true,
						PasswordAuthSecret: "authentication-secret",
					},
				}
				p.WithDefaults()
			})

			Context("First reconcile", func() {
				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})
			})

			Context("Controller", func() {

				It("should create a pod disruption budget", func() {
					_ = pravega.MakeControllerPodDisruptionBudget(p)
					Ω(err).Should(BeNil())
				})

				It("should create a config-map", func() {
					_ = pravega.MakeControllerConfigMap(p)
					Ω(err).Should(BeNil())
				})

				It("should create the deployment", func() {
					_ = pravega.MakeControllerDeployment(p)
					Ω(err).Should(BeNil())
				})

				It("should create the service", func() {
					_ = pravega.MakeControllerService(p)
					Ω(err).Should(BeNil())
				})

			})

		})

		Context("Controller Svc Type Load Balancer", func() {
			var (
				customReq *corev1.ResourceRequirements
				err       error
			)

			BeforeEach(func() {
				annotationsMap := map[string]string{
					"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
				}
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
					Version:      "0.5.0",
					ZookeeperUri: "example.com",
					ExternalAccess: &v1alpha1.ExternalAccess{
						Enabled:    true,
						Type:       corev1.ServiceTypeClusterIP,
						DomainName: "pravega.com.",
					},
					Bookkeeper: &v1alpha1.BookkeeperSpec{
						Replicas:  5,
						Resources: customReq,
						Image: &v1alpha1.BookkeeperImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Repository: "foo/bookkeeper",
							},
						},
						BookkeeperJVMOptions: &v1alpha1.BookkeeperJVMOptions{
							MemoryOpts:    []string{"-Xms2g", "-XX:MaxDirectMemorySize=2g"},
							GcOpts:        []string{"-XX:MaxGCPauseMillis=20", "-XX:-UseG1GC"},
							GcLoggingOpts: []string{"-XX:NumberOfGCLogFiles=10"},
						},
					},
					Pravega: &v1alpha1.PravegaSpec{
						ControllerReplicas:             2,
						SegmentStoreReplicas:           4,
						ControllerServiceAccountName:   "pravega-components",
						SegmentStoreServiceAccountName: "pravega-components",
						ControllerResources:            customReq,
						SegmentStoreResources:          customReq,
						ControllerExternalServiceType:  corev1.ServiceTypeLoadBalancer,
						ControllerServiceAnnotations:   annotationsMap,
						SegmentStoreServiceAnnotations: annotationsMap,
						Image: &v1alpha1.PravegaImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Repository: "bar/pravega",
							},
						},
						ControllerJvmOptions:   []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMFraction=1"},
						SegmentStoreJVMOptions: []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMFraction=1"},
						Options: map[string]string{
							"dummy-key": "dummy-value",
						},
						CacheVolumeClaimTemplate: &corev1.PersistentVolumeClaimSpec{
							VolumeName: "abc",
						},
					},
					TLS: &v1alpha1.TLSPolicy{
						Static: &v1alpha1.StaticTLS{
							ControllerSecret:   "controller-secret",
							SegmentStoreSecret: "segmentstore-secret",
						},
					},
					Authentication: &v1alpha1.AuthenticationParameters{
						Enabled:            true,
						PasswordAuthSecret: "authentication-secret",
					},
				}
				p.WithDefaults()
			})

			Context("First reconcile", func() {
				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})
			})

			Context("Controller", func() {

				It("should create a pod disruption budget", func() {
					_ = pravega.MakeControllerPodDisruptionBudget(p)
					Ω(err).Should(BeNil())
				})

				It("should create a config-map", func() {
					_ = pravega.MakeControllerConfigMap(p)
					Ω(err).Should(BeNil())
				})

				It("should create the deployment", func() {
					_ = pravega.MakeControllerDeployment(p)
					Ω(err).Should(BeNil())
				})

				It("should create the service", func() {
					_ = pravega.MakeControllerService(p)
					Ω(err).Should(BeNil())
				})

			})

		})

	})
})
