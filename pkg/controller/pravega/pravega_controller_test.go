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
	"strings"
	"testing"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
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
			p *v1beta1.PravegaCluster
		)

		BeforeEach(func() {
			p = &v1beta1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			p.Spec.Version = "0.5.0"
		})

		Context("Empty Controller Service Type", func() {
			var (
				customReq *corev1.ResourceRequirements
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
					Version:      "0.5.0",
					ZookeeperUri: "example.com",
					ExternalAccess: &v1beta1.ExternalAccess{
						Enabled:    true,
						Type:       corev1.ServiceTypeClusterIP,
						DomainName: "pravega.com.",
					},
					BookkeeperUri: v1beta1.DefaultBookkeeperUri,
					Pravega: &v1beta1.PravegaSpec{
						ControllerReplicas:             2,
						SegmentStoreReplicas:           4,
						ControllerServiceAccountName:   "pravega-components",
						SegmentStoreServiceAccountName: "pravega-components",
						ControllerResources:            customReq,
						SegmentStoreResources:          customReq,
						ControllerServiceAnnotations:   annotationsMap,
						ControllerPodLabels:            annotationsMap,
						ControllerPodAnnotations:       annotationsMap,
						SegmentStoreServiceAnnotations: annotationsMap,
						SegmentStorePodLabels:          annotationsMap,
						InfluxDBSecret: &v1beta1.InfluxDBSecret{
							Secret:    "influxdb-secret",
							MountPath: "",
						},
						Image: &v1beta1.ImageSpec{
							Repository: "bar/pravega",
						},
						ControllerJvmOptions:   []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50.0"},
						SegmentStoreJVMOptions: []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50.0"},
						Options: map[string]string{
							"dummy-key":             "dummy-value",
							"configMapVolumeMounts": "prvg-logback:logback.xml=/opt/pravega/conf/logback.xml",
							"emptyDirVolumeMounts":  "heap-dump=/tmp/dumpfile/heap,log=/opt/pravega/logs",
							"hostPathVolumeMounts":  "heap-dump=/tmp/dumpfile/heap,log=/opt/pravega/logs",
						},
						CacheVolumeClaimTemplate: &corev1.PersistentVolumeClaimSpec{
							VolumeName: "abc",
						},
						DebugLogging: true,
						AuthImplementations: &v1beta1.AuthImplementationSpec{
							MountPath: "/ifs/data",
							AuthHandlers: []v1beta1.AuthHandlerSpec{
								{
									Image:  "testimage1",
									Source: "source1",
								},
								{
									Image:  "testimage2",
									Source: "source2",
								},
							},
						},
					},
					TLS: &v1beta1.TLSPolicy{
						Static: &v1beta1.StaticTLS{
							ControllerSecret:   "controller-secret",
							SegmentStoreSecret: "segmentstore-secret",
						},
					},
					Authentication: &v1beta1.AuthenticationParameters{
						Enabled:               true,
						PasswordAuthSecret:    "authentication-secret",
						ControllerTokenSecret: "controllerauth-secret",
					},
				}
				p.WithDefaults()
				no := int64(0)
				securitycontext := corev1.PodSecurityContext{
					RunAsUser: &no,
				}

				p.Spec.Pravega.ControllerSecurityContext = &securitycontext

				p.Spec.Pravega.ControllerInitContainers = []corev1.Container{
					corev1.Container{
						Name:    "testing",
						Image:   "dummy-image",
						Command: []string{"sh", "-c", "ls;pwd"},
					},
				}
			})

			Context("Controller", func() {

				It("should create a pod disruption budget", func() {
					pdb := pravega.MakeControllerPodDisruptionBudget(p)
					Ω(pdb.Name).Should(Equal(p.PdbNameForController()))
				})

				It("should create a config-map and set the JVM options given by user", func() {
					cm := pravega.MakeControllerConfigMap(p)
					javaOpts := cm.Data["JAVA_OPTS"]
					Ω(cm.Data["log.level"]).Should(Equal("DEBUG"))
					Ω(strings.Contains(javaOpts, "-Dpravegaservice.clusterName=default")).Should(BeTrue())
					Ω(strings.Contains(javaOpts, "-XX:MaxDirectMemorySize=1g")).Should(BeTrue())
					Ω(strings.Contains(javaOpts, "-XX:MaxRAMPercentage=50.0")).Should(BeTrue())
				})

				It("should create the deployment", func() {
					deploy := pravega.MakeControllerDeployment(p)
					Ω(*deploy.Spec.Replicas).Should(Equal(int32(2)))
				})

				It("should have configMap/emptyDir/hostPath VolumeMounts set to the values given by user", func() {
					deploy := pravega.MakeControllerPodTemplate(p)
					mounthostpath0 := deploy.Spec.Containers[0].VolumeMounts[4].MountPath
					Ω(mounthostpath0).Should(Equal("/opt/pravega/conf/logback.xml"))
					mounthostpath1 := deploy.Spec.Containers[0].VolumeMounts[0].MountPath
					Ω(mounthostpath1).Should(Equal("/tmp/dumpfile/heap"))
					mounthostpath2 := deploy.Spec.Containers[0].VolumeMounts[1].MountPath
					Ω(mounthostpath2).Should(Equal("/opt/pravega/logs"))
					mounthostpath3 := deploy.Spec.Containers[0].VolumeMounts[2].MountPath
					Ω(mounthostpath3).Should(Equal("/tmp/dumpfile/heap"))
					mounthostpath4 := deploy.Spec.Containers[0].VolumeMounts[3].MountPath
					Ω(mounthostpath4).Should(Equal("/opt/pravega/logs"))

				})

				It("should have VolumeMounts created for influxdb secret", func() {
					deploy := pravega.MakeControllerPodTemplate(p)
					mounthostpath := deploy.Spec.Containers[0].VolumeMounts[9].MountPath
					Ω(mounthostpath).Should(Equal("/etc/influxdb-secret-volume"))
				})
				It("should create the service", func() {
					svc := pravega.MakeControllerService(p)
					Ω(svc.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
				})
				It("should have initcontainers ", func() {
					podTemplate := pravega.MakeControllerPodTemplate(p)
					Ω(podTemplate.Spec.InitContainers[0].Name).To(Equal("testing"))
					Ω(podTemplate.Spec.InitContainers[0].Image).To(Equal("dummy-image"))
					Ω(strings.Contains(podTemplate.Spec.InitContainers[0].Command[2], "ls;pwd")).Should(BeTrue())
					Ω(podTemplate.Spec.InitContainers[1].Image).To(Equal("testimage1"))
					Ω(podTemplate.Spec.InitContainers[1].Name).To(Equal("authplugin0"))
					Ω(podTemplate.Spec.InitContainers[1].VolumeMounts[0].Name).To(Equal("authplugin"))
					Ω(podTemplate.Spec.InitContainers[1].VolumeMounts[0].MountPath).To(Equal("/ifs/data"))
					Ω(podTemplate.Spec.InitContainers[2].Name).To(Equal("authplugin1"))
					Ω(podTemplate.Spec.InitContainers[2].Image).To(Equal("testimage2"))
					Ω(podTemplate.Spec.InitContainers[2].VolumeMounts[0].Name).To(Equal("authplugin"))
					Ω(podTemplate.Spec.InitContainers[2].VolumeMounts[0].MountPath).To(Equal("/ifs/data"))
				})
			})

			Context("Controller with external service type and external access type empty", func() {
				BeforeEach(func() {
					p.Spec.Pravega.ControllerExternalServiceType = ""
					p.Spec.ExternalAccess.Type = ""
				})
				It("should create the service with external access type loadbalancer", func() {
					svc := pravega.MakeControllerService(p)
					Ω(svc.Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
				})
				It("should have runAsUser value as 0", func() {
					podTemplate := pravega.MakeControllerPodTemplate(p)
					Ω(fmt.Sprintf("%v", *podTemplate.Spec.SecurityContext.RunAsUser)).To(Equal("0"))
					Ω(podTemplate.Annotations["service.beta.kubernetes.io/aws-load-balancer-type"]).To(Equal("nlb"))
					Ω(podTemplate.Labels["service.beta.kubernetes.io/aws-load-balancer-type"]).To(Equal("nlb"))
				})
			})
			Context("Controller with external service type empty", func() {
				BeforeEach(func() {
					p.Spec.Pravega.ControllerExternalServiceType = ""
				})
				It("should create the service with external access type clusterIP", func() {
					svc := pravega.MakeControllerService(p)
					Ω(svc.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
				})
			})
		})

		Context("Controller Svc Type Load Balancer", func() {
			var (
				customReq *corev1.ResourceRequirements
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
					Version:      "0.5.0",
					ZookeeperUri: "example.com",
					ExternalAccess: &v1beta1.ExternalAccess{
						Enabled:    true,
						Type:       corev1.ServiceTypeClusterIP,
						DomainName: "pravega.com.",
					},
					BookkeeperUri: v1beta1.DefaultBookkeeperUri,
					Pravega: &v1beta1.PravegaSpec{
						ControllerReplicas:             2,
						SegmentStoreReplicas:           4,
						ControllerServiceAccountName:   "pravega-components",
						SegmentStoreServiceAccountName: "pravega-components",
						ControllerResources:            customReq,
						SegmentStoreResources:          customReq,
						ControllerExternalServiceType:  corev1.ServiceTypeLoadBalancer,
						ControllerServiceAnnotations:   annotationsMap,
						ControllerPodLabels:            annotationsMap,
						SegmentStoreServiceAnnotations: annotationsMap,
						SegmentStorePodLabels:          annotationsMap,
						Image: &v1beta1.ImageSpec{
							Repository: "bar/pravega",
						},
						ControllerJvmOptions:   []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50.0"},
						SegmentStoreJVMOptions: []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMPercentage=50.0"},
						Options: map[string]string{
							"dummy-key":             "dummy-value",
							"configMapVolumeMounts": "prvg-logback:logback.xml=/opt/pravega/conf/logback.xml",
							"emptyDirVolumeMounts":  "heap-dump=/tmp/dumpfile/heap,log=/opt/pravega/logs",
						},
						CacheVolumeClaimTemplate: &corev1.PersistentVolumeClaimSpec{
							VolumeName: "abc",
						},
						AuthImplementations: &v1beta1.AuthImplementationSpec{
							AuthHandlers: []v1beta1.AuthHandlerSpec{
								{
									Image: "testimage1",
								},
							},
						},
					},
					TLS: &v1beta1.TLSPolicy{
						Static: &v1beta1.StaticTLS{
							ControllerSecret:   "controller-secret",
							SegmentStoreSecret: "segmentstore-secret",
						},
					},
					Authentication: &v1beta1.AuthenticationParameters{
						Enabled:               true,
						PasswordAuthSecret:    "authentication-secret",
						ControllerTokenSecret: "controllerauth-secret",
					},
				}
				p.WithDefaults()
			})

			Context("Controller", func() {

				It("should create a pod disruption budget", func() {
					pdb := pravega.MakeControllerPodDisruptionBudget(p)
					Ω(pdb.Spec.Selector.MatchLabels).To(Equal(p.LabelsForController()))
				})

				It("should create a config-map and set the JVM options given by user", func() {
					cm := pravega.MakeControllerConfigMap(p)
					javaOpts := cm.Data["JAVA_OPTS"]
					Ω(cm.Data["ZK_URL"]).To(Equal(p.Spec.ZookeeperUri))
					Ω(strings.Contains(javaOpts, "-Dpravegaservice.clusterName=default")).Should(BeTrue())
					Ω(strings.Contains(javaOpts, "-XX:MaxDirectMemorySize=1g")).Should(BeTrue())
					Ω(strings.Contains(javaOpts, "-XX:MaxRAMPercentage=50.0")).Should(BeTrue())
				})

				It("should create the deployment", func() {
					deploy := pravega.MakeControllerDeployment(p)
					Ω(deploy.Name).To(Equal(p.DeploymentNameForController()))
				})

				It("should have configMap/emptyDir VolumeMounts set to the values given by user", func() {
					deploy := pravega.MakeControllerPodTemplate(p)
					mounthostpath0 := deploy.Spec.Containers[0].VolumeMounts[2].MountPath
					Ω(mounthostpath0).Should(Equal("/opt/pravega/conf/logback.xml"))
					mounthostpath1 := deploy.Spec.Containers[0].VolumeMounts[0].MountPath
					Ω(mounthostpath1).Should(Equal("/tmp/dumpfile/heap"))
					mounthostpath2 := deploy.Spec.Containers[0].VolumeMounts[1].MountPath
					Ω(mounthostpath2).Should(Equal("/opt/pravega/logs"))
				})

				It("should have initcontainers created for auth handelers", func() {
					podTemplate := pravega.MakeControllerPodTemplate(p)
					Ω(podTemplate.Spec.InitContainers[0].Image).To(Equal("testimage1"))
					Ω(podTemplate.Spec.InitContainers[0].Name).To(Equal("authplugin0"))
					Ω(podTemplate.Spec.InitContainers[0].VolumeMounts[0].Name).To(Equal("authplugin"))
					Ω(podTemplate.Spec.InitContainers[0].VolumeMounts[0].MountPath).To(Equal("/opt/pravega/pluginlib"))

				})
				It("should create the service", func() {
					svc := pravega.MakeControllerService(p)
					Ω(svc.Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
				})
			})
		})
	})
})
