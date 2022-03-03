/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1beta1_test

import (
	"os"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestV1beta1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PravegaCluster API")
}

var _ = Describe("PravegaCluster Types Spec", func() {

	var (
		p v1beta1.PravegaCluster
	)

	BeforeEach(func() {
		p = v1beta1.PravegaCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		}
	})

	Context("WithDefaults", func() {
		var changed bool

		BeforeEach(func() {
			changed = p.WithDefaults()
			p.Spec.ExternalAccess.Type = "LoadBalancer"
			p.Spec.ExternalAccess.DomainName = "example.com"
			p.WithDefaults()
		})

		It("should return as changed", func() {
			Ω(changed).Should(BeTrue())
		})

		It("should set zookeeper uri", func() {
			Ω(p.Spec.ZookeeperUri).Should(Equal("zookeeper-client:2181"))
		})

		It("should set external access", func() {
			Ω(p.Spec.ExternalAccess).ShouldNot(BeNil())
		})

		It("should set version to 0.9.0", func() {
			Ω(p.Spec.Version).Should(Equal("0.9.0"))
		})

		It("should set pravega spec", func() {
			Ω(p.Spec.Pravega).ShouldNot(BeNil())
		})

		It("should set bookkeeper uri", func() {
			Ω(p.Spec.BookkeeperUri).ShouldNot(BeNil())
		})
		It("IssecureController should return false", func() {
			Ω(p.Spec.TLS.IsSecureController()).To(Equal(false))
		})

		It("IsSecureSegmentStore should return false", func() {
			Ω(p.Spec.TLS.IsSecureSegmentStore()).To(Equal(false))
		})

		It("IsCaBundlePresent should return false", func() {
			Ω(p.Spec.TLS.IsCaBundlePresent()).To(Equal(false))
		})

		It("Autentication Enabled should return false", func() {
			Ω(p.Spec.Authentication.IsEnabled()).To(Equal(false))
		})

		It("should set external access type and domain name to empty", func() {
			Ω(string(p.Spec.ExternalAccess.Type)).Should(Equal(""))
			Ω(p.Spec.ExternalAccess.DomainName).Should(Equal(""))
		})
	})

	Context("ValidatePravegaVersion", func() {
		var (
			p     *v1beta1.PravegaCluster
			file1 *os.File
		)

		BeforeEach(func() {
			p = &v1beta1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			p.WithDefaults()
		})

		Context("Spec version empty", func() {
			var (
				err error
			)
			BeforeEach(func() {
				p.Spec.Version = ""
				err = p.ValidatePravegaVersion()
			})
			It("should return nil", func() {
				Ω(err).To(BeNil())
			})
		})

		Context("Version not in valid format", func() {
			var (
				err error
			)
			BeforeEach(func() {
				p.Spec.Version = "999"
				err = p.ValidatePravegaVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "request version is not in valid format")).Should(Equal(true))
			})
		})

		Context("Spec version and current version same", func() {
			var (
				err error
			)
			BeforeEach(func() {
				p.Spec.Version = "0.7.0"
				p.Status.CurrentVersion = "0.7.0"
				err = p.ValidatePravegaVersion()
			})
			It("should return nil", func() {
				Ω(err).To(BeNil())
			})
		})

		Context("current version not in correct format", func() {
			var (
				err error
			)
			BeforeEach(func() {
				p.Spec.Version = "0.7.0"
				p.Status.CurrentVersion = "999"
				err = p.ValidatePravegaVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "found version is not in valid format")).Should(Equal(true))
			})

		})

		Context("unsupported upgrade to a version", func() {
			var (
				err error
			)
			BeforeEach(func() {
				p.Status.CurrentVersion = "0.7.2"
				p.Spec.Version = "0.7.0"
				err = p.ValidatePravegaVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "downgrading the cluster from version 0.7.2 to 0.7.0 is not supported")).Should(Equal(true))
			})
		})

		Context("supported upgrade to a version", func() {
			var (
				err error
			)
			BeforeEach(func() {
				p.Status.CurrentVersion = "0.7.0"
				p.Spec.Version = "0.7.1"
				err = p.ValidatePravegaVersion()
			})
			It("should return nil", func() {
				Ω(err).To(BeNil())
			})
		})

		Context("validation while cluster upgrade in progress", func() {
			var (
				err error
			)
			BeforeEach(func() {
				p.Status.SetUpgradingConditionTrue(" ", " ")
				p.Spec.Version = "0.7.1"
				p.Status.TargetVersion = "0.7.0"
				err = p.ValidatePravegaVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "failed to process the request, cluster is upgrading")).Should(Equal(true))
			})
		})

		Context("validation while cluster rollback in progress", func() {
			var (
				err error
			)
			BeforeEach(func() {
				p.Status.CurrentVersion = "0.7.0"
				p.Status.Init()
				p.Status.AddToVersionHistory("0.6.0")
				p.Status.SetRollbackConditionTrue(" ", " ")
				p.Spec.Version = "0.7.0"
				err = p.ValidatePravegaVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "failed to process the request, rollback in progress")).Should(Equal(true))
			})
		})

		Context("validation while cluster in error state", func() {
			var (
				err error
			)
			BeforeEach(func() {
				p.Status.SetErrorConditionTrue("some err", " ")
				p.Spec.Version = "0.7.0"
				err = p.ValidatePravegaVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "failed to process the request, cluster is in error state")).Should(Equal(true))
			})
		})

		Context("validation while cluster in UpgradeFailed state", func() {
			var (
				err error
			)
			BeforeEach(func() {
				p.Status.CurrentVersion = "0.7.0"
				p.Status.Init()
				p.Status.AddToVersionHistory("0.6.0")
				p.Status.SetErrorConditionTrue("UpgradeFailed", " ")
				p.Spec.Version = "0.7.0"
				err = p.ValidatePravegaVersion()
			})
			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Rollback to version 0.7.0 not supported")).Should(Equal(true))
			})
		})

		Context("validation while cluster in UpgradeFailed state and supported rollback version", func() {
			var (
				err error
			)
			BeforeEach(func() {
				p.Status.CurrentVersion = "0.6.0"
				p.Status.Init()
				p.Status.AddToVersionHistory("0.6.0")
				p.Status.SetErrorConditionTrue("UpgradeFailed", " ")
				p.Spec.Version = "0.6.0"
				err = p.ValidatePravegaVersion()
			})
			It("should return nil", func() {
				Ω(err).To(BeNil())
			})
		})

		AfterEach(func() {
			file1.Close()
			os.Remove("filename")
		})
	})

	Context("Setting TLS and Autentication to nil", func() {
		BeforeEach(func() {
			p.Spec.Version = "0.6.0"
			p.WithDefaults()
			p.Spec.Authentication = nil
			p.Spec.TLS = nil
		})
		It("Autentication Enabled should return false", func() {
			Ω(p.Spec.Authentication.IsEnabled()).To(Equal(false))
		})
		It("IsCaBundlePresent should return false", func() {
			Ω(p.Spec.TLS.IsCaBundlePresent()).To(Equal(false))
		})
		It("IsSecureSegmentStore should return false", func() {
			Ω(p.Spec.TLS.IsSecureSegmentStore()).To(Equal(false))
		})
		It("IssecureController should return false", func() {
			Ω(p.Spec.TLS.IsSecureController()).To(Equal(false))
		})
	})

	Context("Checking various parameters", func() {
		BeforeEach(func() {
			p.Spec.Version = "0.6.0"
			p.WithDefaults()
		})
		annotationsMap := map[string]string{
			"annotations": "test",
			"labels":      "test",
		}
		It("should set volume claim template", func() {
			Ω(p.Spec.Pravega.CacheVolumeClaimTemplate).ShouldNot(BeNil())
		})

		name := p.StatefulSetNameForSegmentstore()
		It("Should return segmentstore sts name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		p.WithDefaults()
		name = p.StatefulSetNameForSegmentstoreAbove07()
		It("Should return segmentstore sts name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.StatefulSetNameForSegmentstoreBelow07()
		It("Should return segmentstore sts name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		p.WithDefaults()
		name = p.PravegaControllerServiceURL()
		It("Should return controller service url", func() {
			Ω(name).ShouldNot(BeNil())
		})

		p.Spec.Pravega.ControllerPodLabels = annotationsMap
		labels := p.LabelsForController()

		It("Should return controller labels", func() {
			Ω(labels).ShouldNot(BeNil())
		})

		p.Spec.Pravega.SegmentStorePodLabels = annotationsMap
		labels = p.LabelsForSegmentStore()
		It("Should return segmentstore labels", func() {
			Ω(labels).ShouldNot(BeNil())
		})

		p.Spec.Pravega.ControllerPodAnnotations = annotationsMap
		annotations := p.AnnotationsForController()
		It("Should return controller annotations", func() {
			Ω(annotations).ShouldNot(BeNil())
		})

		p.Spec.Pravega.SegmentStorePodAnnotations = annotationsMap
		annotations = p.AnnotationsForSegmentStore()
		It("Should return segmentstore annotations", func() {
			Ω(annotations).ShouldNot(BeNil())
		})

		labels = p.LabelsForPravegaCluster()
		It("Should return pravega labels", func() {
			Ω(labels).ShouldNot(BeNil())
		})

		name = p.PdbNameForController()
		It("Event size should not be zero", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ConfigMapNameForController()
		It("Should return controller configmap name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ServiceNameForController()
		It("Should return controller service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ServiceNameForSegmentStore(0)
		It("Should return segmentstore service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ServiceNameForSegmentStoreBelow07(0)
		It("Should return segmentstore service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ServiceNameForSegmentStoreAbove07(0)
		It("Should return segmentstore service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.HeadlessServiceNameForSegmentStore()
		It("Should return segmentstore headless service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.HeadlessServiceNameForBookie()
		It("Should return bookie headless service name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.DeploymentNameForController()
		It("Should return controller deployment name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.PdbNameForSegmentstore()
		It("Should return segmentstore pdb name", func() {
			Ω(name).ShouldNot(BeNil())
		})

		name = p.ConfigMapNameForSegmentstore()
		It("Should return segmentstore configmap name", func() {
			Ω(name).ShouldNot(BeNil())
		})
	})

	Context("checking event generation utility", func() {
		BeforeEach(func() {
			p.WithDefaults()
		})
		message := "upgrade failed"
		event := p.NewEvent("UPGRADE_ERROR", v1beta1.UpgradeErrorReason, message, "Error")
		It("Event size should not be zero", func() {
			Ω(event.Size()).ShouldNot(Equal(0))
		})

		event = p.NewApplicationEvent("UPGRADE_ERROR", v1beta1.UpgradeErrorReason, message, "Error")
		It("Event size should not be zero", func() {
			Ω(event.Size()).ShouldNot(Equal(0))
		})
	})

	Context("WaitForClusterToTerminate", func() {
		var (
			client client.Client
			err    error
			p1     *v1beta1.PravegaCluster
		)

		BeforeEach(func() {
			p1 = &v1beta1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			p1.WithDefaults()
			s := scheme.Scheme
			s.AddKnownTypes(v1beta1.SchemeGroupVersion, p1)
			client = fake.NewFakeClient(p1)
			err = p1.WaitForClusterToTerminate(client)
		})
		It("should  be nil", func() {
			Ω(err).Should(BeNil())
		})
	})

	Context("Validate Segment Store Memory Settings", func() {
		var (
			p *v1beta1.PravegaCluster
		)

		BeforeEach(func() {
			p = &v1beta1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			p.WithDefaults()
		})

		Context("validating with the correct spec", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "1610612736"
				p.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m"}
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return nil", func() {
				Ω(err).Should(BeNil())
			})
		})

		Context("empty segmentStoreResources object", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.SegmentStoreResources = nil
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "spec.pravega.segmentStoreResources cannot be empty")).Should(Equal(true))
			})
		})

		Context("empty segmentStoreResources.limits object", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("5Gi"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "spec.pravega.segmentStoreResources.limits cannot be empty")).Should(Equal(true))
			})
		})

		Context("empty segmentStoreResources.requests object", func() {
			var (
				changed bool
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "1610612736"
				p.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m"}
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
				changed = p.WithDefaults()
			})

			It("Should set memory and cpu requests to memory and cpu limits respectively", func() {
				Ω(changed).Should(Equal(true))
			})
		})

		Context("empty segmentStoreResources.requests object validation check", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "1610612736"
				p.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m"}
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("should return nil", func() {
				Ω(err).Should(BeNil())
			})
		})

		Context("memory limits and requests are not set", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("1000m"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("2000m"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Missing required value for field spec.pravega.segmentStoreResources.limits.memory")).Should(Equal(true))
			})
		})

		Context("Only memory limits is set", func() {
			var (
				changed bool
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "1610612736"
				p.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m"}
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("1000m"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
				changed = p.WithDefaults()
			})

			It("Should set memory requests to memory limits", func() {
				Ω(changed).Should(Equal(true))
			})
		})

		Context("CPU limits and requests are not set", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Missing required value for field spec.pravega.segmentStoreResources.limits.cpu")).Should(Equal(true))
			})
		})

		Context("Only cpu limits is set", func() {
			var (
				changed bool
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "1610612736"
				p.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m"}
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("4Gi"),
						corev1.ResourceCPU:    resource.MustParse("2000m"),
					},
				}
				changed = p.WithDefaults()
			})

			It("Should set cpu requests to cpu limits", func() {
				Ω(changed).Should(Equal(true))
			})
		})

		Context("memory requests is greater than memory limits", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("5Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "spec.pravega.segmentStoreResources.requests.memory value must be less than or equal to spec.pravega.segmentStoreResources.limits.memory")).Should(Equal(true))
			})
		})

		Context("CPU requests is greater than CPU limits", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "spec.pravega.segmentStoreResources.requests.cpu value must be less than or equal to spec.pravega.segmentStoreResources.limits.cpu")).Should(Equal(true))
			})
		})

		Context("pravegaservice.cache.size.max is not set", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["pravegaservice.cache.size.max"] = ""
				p.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m"}
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Missing required value for option pravegaservice.cache.size.max")).Should(Equal(true))
			})
		})

		Context("JVM option -Xmx is not set", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "1610612736"
				p.Spec.Pravega.SegmentStoreJVMOptions = []string{"-XX:MaxDirectMemorySize=2560m"}
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Missing required value for Segment Store JVM Option -Xmx")).Should(Equal(true))
			})
		})

		Context("JVM option -XX:MaxDirectMemorySize is not set", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "1610612736"
				p.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g"}
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Missing required value for Segment Store JVM option -XX:MaxDirectMemorySize")).Should(Equal(true))
			})
		})

		Context("sum of MaxDirectMemorySize and Xmx is greater than total memory limit", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "1610612736"
				p.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m"}
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("3Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("3Gi"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "MaxDirectMemorySize along with JVM Xmx value should be less than the total available memory!")).Should(Equal(true))
			})
		})

		Context("pravegaservice.cache.size.max is greater than MaxDirectMemorySize", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["pravegaservice.cache.size.max"] = "3221225472"
				p.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx1g", "-XX:MaxDirectMemorySize=2560m"}
				p.Spec.Pravega.SegmentStoreResources = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2000m"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
				}
				err = p.ValidateSegmentStoreMemorySettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Cache size configured should be less than the JVM MaxDirectMemorySize value")).Should(Equal(true))
			})
		})
	})

	Context("Validate Bookie Settings", func() {
		var (
			p *v1beta1.PravegaCluster
		)

		BeforeEach(func() {
			p = &v1beta1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			p.WithDefaults()
		})

		Context("Validating with correct values for Ensemble Size, Write Quorum size and Ack quorum size", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
				p.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "3"
				p.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
				err = p.ValidateBookkeperSettings()
			})

			It("Should return nil", func() {
				Ω(err).Should(BeNil())
			})
		})

		Context("Invalid Value for Enseble size", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3.4"
				err = p.ValidateBookkeperSettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Cannot convert ensemble size from string to integer")).Should(Equal(true))
			})
		})

		Context("Invalid Value for Write Quorum size", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "3##4"
				err = p.ValidateBookkeperSettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Cannot convert write quorum size from string to integer")).Should(Equal(true))
			})
		})

		Context("Invalid Value for Ack Quorum size", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "2!342"
				err = p.ValidateBookkeperSettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "Cannot convert ack quorum size from string to integer")).Should(Equal(true))
			})
		})

		Context("Validating with Ensemble Size < Write Quorum Size", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
				p.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "4"
				p.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
				err = p.ValidateBookkeperSettings()
			})

			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "The value provided for the option bookkeeper.write.quorum.size should be less than or equal to the value of option bookkeeper.ensemble.size")).Should(Equal(true))
			})
		})

		Context("Validating with Ensemble size <=2 and Write quorum size is set to default", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "2"
				p.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = ""
				p.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
				err = p.ValidateBookkeperSettings()
			})

			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "The value provided for the option bookkeeper.ensemble.size should be greater than or equal to the value of option bookkeeper.write.quorum.size (default is 3)")).Should(Equal(true))
			})
		})

		Context("Validating with Write quorum size > 3 and Ensemble size is set to default", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.ensemble.size"] = ""
				p.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "4"
				p.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "3"
				err = p.ValidateBookkeperSettings()
			})

			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "The value provided for the option bookkeeper.write.quorum.size should be less than or equal to the value of option bookkeeper.ensemble.size (default is 3)")).Should(Equal(true))
			})
		})

		Context("Validating whether minimum racks count is set to true false or \"\"", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.write.quorum.racks.minimumCount.enable"] = "True"
				err = p.ValidateBookkeperSettings()
			})

			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "bookkeeper.write.quorum.racks.minimumCount.enable can be only set to \"true\" \"false\" or \"\"")).Should(Equal(true))
			})
		})

		Context("Validating with Enseble Size set to 1 and minimum count enabled is set to true", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "1"
				p.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = ""
				p.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = ""
				p.Spec.Pravega.Options["bookkeeper.write.quorum.racks.minimumCount.enable"] = "true"
				err = p.ValidateBookkeperSettings()
			})

			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "bookkeeper.write.quorum.racks.minimumCount.enable should be set to false if bookkeeper.ensemble.size is 1")).Should(Equal(true))
			})
		})

		Context("Validating with Write quorum size < Acq quorum size", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "4"
				p.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "4"
				p.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "5"
				err = p.ValidateBookkeperSettings()
			})

			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "The value provided for the option bookkeeper.ack.quorum.size should be less than or equal to the value of option bookkeeper.write.quorum.size")).Should(Equal(true))
			})
		})

		Context("Validating with Write quorum size <=2 and Acq quorum size is set to default", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
				p.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = "2"
				p.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = ""
				err = p.ValidateBookkeperSettings()
			})

			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "The value provided for the option bookkeeper.write.quorum.size should be greater than or equal to the value of option bookkeeper.ack.quorum.size (default is 3)")).Should(Equal(true))
			})
		})

		Context("Validating with Ack quorum size > 3 and Write quorum size is set to default", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Pravega.Options["bookkeeper.ensemble.size"] = "3"
				p.Spec.Pravega.Options["bookkeeper.write.quorum.size"] = ""
				p.Spec.Pravega.Options["bookkeeper.ack.quorum.size"] = "4"
				err = p.ValidateBookkeperSettings()
			})

			It("should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "The value provided for the option bookkeeper.ack.quorum.size should be less than or equal to the value of option bookkeeper.write.quorum.size (default is 3)")).Should(Equal(true))
			})
		})
	})
	Context("Validate Authentication Settings", func() {
		var (
			p *v1beta1.PravegaCluster
		)

		BeforeEach(func() {
			p = &v1beta1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			p.WithDefaults()
		})

		Context("Validating with authentication enabled and correct options", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Authentication.Enabled = true
				p.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "true"
				p.Spec.Pravega.Options["controller.security.auth.delegationToken.signingKey.basis"] = "secret"
				p.Spec.Pravega.Options["autoScale.security.auth.token.signingKey.basis"] = "secret"
				err = p.ValidateAuthenticationSettings()
			})

			It("Should return nil", func() {
				Ω(err).Should(BeNil())
			})
		})
		Context("Validating with authentication disabled and correct options", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Authentication.Enabled = false
				p.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "false"
				err = p.ValidateAuthenticationSettings()
			})

			It("Should return nil", func() {
				Ω(err).Should(BeNil())
			})

		})
		Context("Validating with authentication disabled and enabling authentication from segment store", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Authentication.Enabled = false
				p.Spec.Pravega.Options["autoScale.authEnabled"] = "true"
				err = p.ValidateAuthenticationSettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "autoScale.controller.connect.security.auth.enable/autoScale.authEnabled should not be set to true")).Should(Equal(true))
			})
		})

		Context("Validating with authentication enabled from controller and disabled from segmentstore", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Authentication.Enabled = true
				p.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "false"
				err = p.ValidateAuthenticationSettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "autoScale.controller.connect.security.auth.enable/autoScale.authEnabled should be set to true")).Should(Equal(true))
			})
		})
		Context("Validating with authentication enabled from controller and not providing option in segmentstore", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Authentication.Enabled = true
				err = p.ValidateAuthenticationSettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "autoScale.controller.connect.security.auth.enable field is not present")).Should(Equal(true))
			})
		})
		Context("Validating with authentication enabled from controller and not providing controller token signing key", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Authentication.Enabled = true
				p.Spec.Pravega.Options["autoScale.authEnabled"] = "true"
				err = p.ValidateAuthenticationSettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "controller.security.auth.delegationToken.signingKey.basis field is not present")).Should(Equal(true))
			})
		})
		Context("Validating with authentication enabled from controller and not providing segment store token signing key", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Authentication.Enabled = true
				p.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "true"
				p.Spec.Pravega.Options["controller.auth.tokenSigningKey"] = "secret"
				err = p.ValidateAuthenticationSettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "autoScale.security.auth.token.signingKey.basis field is not present")).Should(Equal(true))
			})
		})
		Context("Validating with authentication enabled from controller and providing different sigining key for controller and segmentstore", func() {
			var (
				err error
			)

			BeforeEach(func() {
				p.Spec.Authentication.Enabled = true
				p.Spec.Pravega.Options["autoScale.controller.connect.security.auth.enable"] = "true"
				p.Spec.Pravega.Options["controller.auth.tokenSigningKey"] = "secret"
				p.Spec.Pravega.Options["autoScale.security.auth.token.signingKey.basis"] = "secret1"
				err = p.ValidateAuthenticationSettings()
			})

			It("Should return error", func() {
				Ω(strings.ContainsAny(err.Error(), "controller and segmentstore token signing key should have same value")).Should(Equal(true))
			})
		})
	})
})
