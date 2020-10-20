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
	"fmt"
	"testing"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	"github.com/pravega/pravega-operator/pkg/controller/pravega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSegmentStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pravega")
}

var _ = Describe("PravegaSegmentstore", func() {

	var _ = Describe("SegmentStore Test", func() {
		var (
			p *v1beta1.PravegaCluster
		)

		BeforeEach(func() {
			p = &v1beta1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
		})

		Context("With one SegmentStore replica", func() {
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
				p.Spec = v1beta1.ClusterSpec{
					Version: "0.5.0",
					ExternalAccess: &v1beta1.ExternalAccess{
						Enabled:    true,
						Type:       corev1.ServiceTypeClusterIP,
						DomainName: "pravega.com.",
					},
					BookkeeperUri: v1beta1.DefaultBookkeeperUri,
					Pravega: &v1beta1.PravegaSpec{
						ControllerReplicas:             2,
						SegmentStoreReplicas:           1,
						ControllerServiceAccountName:   "pravega-components",
						SegmentStoreServiceAccountName: "pravega-components",
						ControllerResources:            customReq,
						SegmentStoreResources:          customReq,
						ControllerServiceAnnotations:   annotationsMap,
						SegmentStoreServiceAnnotations: annotationsMap,
						SegmentStoreEnvVars:            "SEG_CONFIG_MAP",
						SegmentStoreSecret: &v1beta1.SegmentStoreSecret{
							Secret:    "seg-secret",
							MountPath: "",
						},
						Image: &v1beta1.ImageSpec{
							Repository: "bar/pravega",
						},
						ControllerJvmOptions:   []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50"},
						SegmentStoreJVMOptions: []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50"},
						Options: map[string]string{
							"dummy-key": "dummy-value",
						},
						LongTermStorage: &v1beta1.LongTermStorageSpec{
							Ecs: &v1beta1.ECSSpec{
								ConfigUri:   "configUri",
								Bucket:      "bucket",
								Prefix:      "prefix",
								Credentials: "credentials",
							},
						},
						DebugLogging: true,
					},
					TLS: &v1beta1.TLSPolicy{
						Static: &v1beta1.StaticTLS{
							ControllerSecret:   "controller-secret",
							SegmentStoreSecret: "segmentstore-secret",
							CaBundle:           "ecs-cert",
						},
					},
					Authentication: &v1beta1.AuthenticationParameters{
						Enabled:            true,
						PasswordAuthSecret: "authentication-secret",
					},
				}
				p.WithDefaults()
				no := int64(0)
				securitycontext := corev1.PodSecurityContext{
					RunAsUser: &no,
				}
				p.Spec.Pravega.SegmentStoreSecurityContext = &securitycontext
			})

			Context("First reconcile", func() {
				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})
			})

			Context("SegmentStore", func() {

				It("should create a headless service", func() {
					_ = pravega.MakeSegmentStoreHeadlessService(p)
					Ω(err).Should(BeNil())
				})

				It("should create a pod disruption budget", func() {
					_ = pravega.MakeSegmentstorePodDisruptionBudget(p)
					Ω(err).Should(BeNil())
				})

				It("should create a config-map", func() {
					_ = pravega.MakeSegmentstoreConfigMap(p)
					Ω(err).Should(BeNil())
				})
				It("should create a config-map with empty tier2", func() {
					p.Spec.Pravega.LongTermStorage = &v1beta1.LongTermStorageSpec{}
					cm := pravega.MakeSegmentstoreConfigMap(p)
					Ω(cm.Data["TIER2_STORAGE"]).To(Equal(""))
					Ω(err).Should(BeNil())
				})
				It("should create a stateful set", func() {
					_ = pravega.MakeSegmentStoreStatefulSet(p)
					Ω(err).Should(BeNil())
				})
				It("should set external access service type to LoadBalancer", func() {
					Ω(p.Spec.ExternalAccess.Type).Should(Equal(corev1.ServiceTypeClusterIP))
				})
				It("should have runAsUser value as 0", func() {
					podTemplate := pravega.MakeSegmentStorePodTemplate(p)
					Ω(fmt.Sprintf("%v", *podTemplate.Spec.SecurityContext.RunAsUser)).To(Equal("0"))
				})
			})
		})

		Context("With more than one SegmentStore replica", func() {
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
				p.Spec = v1beta1.ClusterSpec{
					Version: "0.5.0",
					ExternalAccess: &v1beta1.ExternalAccess{
						Enabled:    true,
						Type:       corev1.ServiceTypeClusterIP,
						DomainName: "pravega.com.",
					},
					BookkeeperUri: v1beta1.DefaultBookkeeperUri,
					Pravega: &v1beta1.PravegaSpec{
						ControllerReplicas:              2,
						SegmentStoreReplicas:            4,
						ControllerServiceAccountName:    "pravega-components",
						SegmentStoreServiceAccountName:  "pravega-components",
						ControllerResources:             customReq,
						SegmentStoreResources:           customReq,
						ControllerServiceAnnotations:    annotationsMap,
						SegmentStoreServiceAnnotations:  annotationsMap,
						SegmentStoreExternalServiceType: corev1.ServiceTypeLoadBalancer,
						SegmentStoreSecret: &v1beta1.SegmentStoreSecret{
							Secret:    "seg-secret",
							MountPath: "/tmp/mount",
						},
						Image: &v1beta1.ImageSpec{
							Repository: "bar/pravega",
						},
						ControllerJvmOptions:   []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50"},
						SegmentStoreJVMOptions: []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50"},
						Options: map[string]string{
							"dummy-key": "dummy-value",
						},
						LongTermStorage: &v1beta1.LongTermStorageSpec{
							FileSystem: &v1beta1.FileSystemSpec{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "claim",
									ReadOnly:  true,
								},
							},
						},
					},
					TLS: &v1beta1.TLSPolicy{
						Static: &v1beta1.StaticTLS{
							ControllerSecret:   "controller-secret",
							SegmentStoreSecret: "segmentstore-secret",
							CaBundle:           "ecs-cert",
						},
					},
					Authentication: &v1beta1.AuthenticationParameters{
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

			Context("SegmentStore", func() {

				It("should create a headless service", func() {
					_ = pravega.MakeSegmentStoreHeadlessService(p)
					Ω(err).Should(BeNil())
				})

				It("should create a pod disruption budget", func() {
					_ = pravega.MakeSegmentstorePodDisruptionBudget(p)
					Ω(err).Should(BeNil())
				})

				It("should create a config-map", func() {
					_ = pravega.MakeSegmentstoreConfigMap(p)
					Ω(err).Should(BeNil())
				})

				It("should create a stateful set", func() {
					_ = pravega.MakeSegmentStoreStatefulSet(p)
					Ω(err).Should(BeNil())
				})

				It("should set external access service type to LoadBalancer", func() {
					Ω(p.Spec.ExternalAccess.Type).Should(Equal(corev1.ServiceTypeClusterIP))
				})

			})
		})

		Context("With HDFS as Tier2", func() {
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
				p.Spec = v1beta1.ClusterSpec{
					Version: "0.5.0",
					ExternalAccess: &v1beta1.ExternalAccess{
						Enabled:    true,
						Type:       corev1.ServiceTypeClusterIP,
						DomainName: "pravega.com.",
					},
					BookkeeperUri: v1beta1.DefaultBookkeeperUri,
					Pravega: &v1beta1.PravegaSpec{
						ControllerReplicas:              2,
						SegmentStoreReplicas:            4,
						ControllerServiceAccountName:    "pravega-components",
						SegmentStoreServiceAccountName:  "pravega-components",
						ControllerResources:             customReq,
						SegmentStoreResources:           customReq,
						ControllerServiceAnnotations:    annotationsMap,
						SegmentStoreServiceAnnotations:  annotationsMap,
						SegmentStoreExternalServiceType: corev1.ServiceTypeLoadBalancer,
						Image: &v1beta1.ImageSpec{
							Repository: "bar/pravega",
						},
						ControllerJvmOptions:   []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50"},
						SegmentStoreJVMOptions: []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50"},
						Options: map[string]string{
							"dummy-key": "dummy-value",
						},
						LongTermStorage: &v1beta1.LongTermStorageSpec{
							Hdfs: &v1beta1.HDFSSpec{
								Uri:               "uri",
								Root:              "root",
								ReplicationFactor: 1,
							},
						},
					},
					TLS: &v1beta1.TLSPolicy{
						Static: &v1beta1.StaticTLS{
							ControllerSecret:   "controller-secret",
							SegmentStoreSecret: "segmentstore-secret",
							CaBundle:           "ecs-cert",
						},
					},
					Authentication: &v1beta1.AuthenticationParameters{
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

			Context("SegmentStore", func() {

				It("should create a headless service", func() {
					_ = pravega.MakeSegmentStoreHeadlessService(p)
					Ω(err).Should(BeNil())
				})

				It("should create a pod disruption budget", func() {
					_ = pravega.MakeSegmentstorePodDisruptionBudget(p)
					Ω(err).Should(BeNil())
				})

				It("should create a config-map", func() {
					_ = pravega.MakeSegmentstoreConfigMap(p)
					Ω(err).Should(BeNil())
				})
				It("should create a stateful set", func() {
					_ = pravega.MakeSegmentStoreStatefulSet(p)
					Ω(err).Should(BeNil())
				})
				It("should set external access service type to LoadBalancer", func() {
					_ = pravega.MakeSegmentStoreExternalServices(p)
					Ω(err).Should(BeNil())
				})
			})
			Context("Create External service with external service type and access type empty", func() {
				BeforeEach(func() {
					p.Spec.Pravega.SegmentStoreExternalServiceType = ""
					p.Spec.ExternalAccess.Type = ""
					p.Spec.ExternalAccess.DomainName = "example"

				})
				It("should create external service with access type loadbalancer", func() {
					svc := pravega.MakeSegmentStoreExternalServices(p)
					Ω(svc[0].Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
					Ω(err).Should(BeNil())
				})
			})
			Context("Create External service with SegmentStoreExternalTrafficPolicy as cluster", func() {
				BeforeEach(func() {
					p.Spec.Pravega.SegmentStoreExternalTrafficPolicy = "cluster"
				})
				It("should create external service with ExternalTrafficPolicy type cluster", func() {
					svc := pravega.MakeSegmentStoreExternalServices(p)
					Ω(svc[0].Spec.ExternalTrafficPolicy).To(Equal(corev1.ServiceExternalTrafficPolicyTypeCluster))
				})
			})
			Context("Create External service with LoadBalancerIP", func() {
				BeforeEach(func() {
					p.Spec.Pravega.SegmentStoreLoadBalancerIP = "10.240.12.18"
				})
				It("should create external service with LoadBalancerIP", func() {
					svc := pravega.MakeSegmentStoreExternalServices(p)
					Ω(svc[0].Spec.LoadBalancerIP).To(Equal("10.240.12.18"))
				})
			})
			Context("Create External service with external service type empty", func() {
				BeforeEach(func() {
					m := make(map[string]string)
					p.Spec.Pravega.SegmentStoreServiceAnnotations = m
					p.Spec.Pravega.SegmentStoreExternalServiceType = ""
				})
				It("should create the service with external access type ClusterIP", func() {
					svc := pravega.MakeSegmentStoreExternalServices(p)
					Ω(svc[0].Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
					Ω(err).Should(BeNil())
				})
			})
		})
	})
})
