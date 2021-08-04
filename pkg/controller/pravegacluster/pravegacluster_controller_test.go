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
	"fmt"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	policyv1beta1 "k8s.io/api/policy/v1beta1"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	"github.com/pravega/pravega-operator/pkg/controller/pravega"

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

func TestPravega(t *testing.T) {
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
			s.AddKnownTypes(v1beta1.SchemeGroupVersion, p)
		})
		Context("Should have Reconcile Result false when request namespace does not contain pravega cluster", func() {
			var (
				client client.Client
				err    error
			)
			BeforeEach(func() {
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				req.NamespacedName.Namespace = "temp"
				res, err = r.Reconcile(req)
			})
			It("should have false in reconcile result", func() {
				Ω(res.Requeue).To(Equal(false))
				Ω(err).To(BeNil())
			})
		})

		Context("Delete STS with sts not present", func() {
			var (
				client       client.Client
				err          error
				foundPravega *v1beta1.PravegaCluster
			)

			BeforeEach(func() {
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				//1st reconcile
				foundPravega = &v1beta1.PravegaCluster{}
				_, _ = r.Reconcile(req)
				_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				err = r.deleteSTS(foundPravega)
			})
			It("shouldn't error", func() {
				Ω(err).Should(BeNil())
			})
		})

		Context("deleteOldSegmentStoreIfExist ", func() {
			var (
				client       client.Client
				err, err1    error
				foundPravega *v1beta1.PravegaCluster
			)

			BeforeEach(func() {
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				//1st reconcile
				foundPravega = &v1beta1.PravegaCluster{}
				_, _ = r.Reconcile(req)
				_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				pvc := &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cache-example-pravega-segmentstore-0",
						Namespace: p.Namespace,
					},
				}
				r.client.Create(context.TODO(), pvc)
				svc := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example-pravega-segmentstore-0",
						Namespace: p.Namespace,
					},
				}
				r.client.Create(context.TODO(), svc)
				sts := &appsv1.StatefulSet{}
				foundPravega.Spec.Version = "0.5.0"
				foundPravega.Spec.Pravega.SegmentStoreReplicas = 2
				foundPravega.Spec.ExternalAccess.Enabled = true
				foundPravega.WithDefaults()
				sts = pravega.MakeSegmentStoreStatefulSet(foundPravega)
				extService := &corev1.Service{}
				pravega.MakeSegmentStoreExternalServices(foundPravega)
				extService.Name = p.ServiceNameForSegmentStoreBelow07(int32(0))
				r.client.Create(context.TODO(), extService)
				svcName := p.ServiceNameForSegmentStoreBelow07(int32(0))
				_ = r.client.Get(context.TODO(), types.NamespacedName{Name: svcName, Namespace: foundPravega.Namespace}, extService)
				r.client.Create(context.TODO(), sts)
				_ = r.client.Get(context.TODO(), types.NamespacedName{Name: sts.Name, Namespace: foundPravega.Namespace}, sts)
				err = r.deleteOldSegmentStoreIfExists(foundPravega)
				foundPravega.Spec.ExternalAccess.Enabled = false
				err1 = r.deleteOldSegmentStoreIfExists(foundPravega)
			})
			It("shouldn't error", func() {
				Ω(err).Should(BeNil())
			})
			It("shouldn't error", func() {
				Ω(err1).Should(BeNil())
			})
		})

		Context("syncControllerSize", func() {
			var (
				client       client.Client
				err, err1    error
				foundPravega *v1beta1.PravegaCluster
			)

			BeforeEach(func() {
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				//1st reconcile
				foundPravega = &v1beta1.PravegaCluster{}
				_, _ = r.Reconcile(req)

				_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				deploy := &appsv1.Deployment{}

				r.client.Create(context.TODO(), deploy)
				name := p.DeploymentNameForController()
				_ = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)

				foundPravega.Spec.Pravega.ControllerReplicas = 2
				err1 = r.syncControllerSize(foundPravega)
				_, _ = r.Reconcile(req)
				err = r.syncControllerSize(foundPravega)

			})
			It("should not give error", func() {
				Ω(err).Should(BeNil())
			})
			It("should give error", func() {
				Ω(strings.ContainsAny(err1.Error(), "failed to get deployment")).Should(Equal(true))
			})
		})

		Context("Without spec", func() {
			var (
				client       client.Client
				err          error
				foundPravega *v1beta1.PravegaCluster
			)

			BeforeEach(func() {
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				//1st reconcile
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
					foundPravega = &v1beta1.PravegaCluster{}
					err = client.Get(context.TODO(), req.NamespacedName, foundPravega)
					Ω(err).Should(BeNil())
					Ω(foundPravega.Spec.Version).Should(Equal(v1beta1.DefaultPravegaVersion))
					Ω(foundPravega.Spec.ZookeeperUri).Should(Equal(v1beta1.DefaultZookeeperUri))
					Ω(foundPravega.Spec.BookkeeperUri).Should(Equal(v1beta1.DefaultBookkeeperUri))
					Ω(foundPravega.Spec.ExternalAccess).ShouldNot(BeNil())
					Ω(foundPravega.Spec.ExternalAccess.Enabled).Should(Equal(false))
					Ω(foundPravega.Spec.ExternalAccess.DomainName).Should(Equal(""))
					Ω(foundPravega.Spec.Pravega).ShouldNot(BeNil())
				})
			})

			Context("After defaults are applied", func() {
				BeforeEach(func() {
					// 2nd reconcile
					res, err = r.Reconcile(req)
				})

				It("should requeue after ReconfileTime delay", func() {
					Ω(res.RequeueAfter).To(Equal(ReconcileTime))
				})

				It("should set current version on 2nd reconcile ", func() {
					res, err = r.Reconcile(req)
					foundPravega := &v1beta1.PravegaCluster{}
					err = client.Get(context.TODO(), req.NamespacedName, foundPravega)
					Ω(err).Should(BeNil())
					Ω(foundPravega.Spec.Version).Should(Equal(v1beta1.DefaultPravegaVersion))
					Ω(foundPravega.Status.CurrentVersion).Should(Equal(v1beta1.DefaultPravegaVersion))
				})
			})

			Context("Cluster deployment", func() {
				BeforeEach(func() {
					// 2nd reconcile
					res, err = r.Reconcile(req)
					foundPravega = &v1beta1.PravegaCluster{}
					err = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				})

				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})

				Context("Controller", func() {
					It("should create a deployment", func() {
						foundController := &appsv1.Deployment{}
						nn := types.NamespacedName{
							Name:      foundPravega.DeploymentNameForController(),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundController)
						Ω(err).Should(BeNil())
					})

					It("should create a config-map", func() {
						foundCm := &corev1.ConfigMap{}
						nn := types.NamespacedName{
							Name:      foundPravega.ConfigMapNameForController(),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundCm)
						Ω(err).Should(BeNil())
					})

					It("should create a client-service", func() {
						foundSvc := &corev1.Service{}
						nn := types.NamespacedName{
							Name:      foundPravega.ServiceNameForController(),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundSvc)
						Ω(err).Should(BeNil())
					})
				})

				Context("SegmentStore", func() {
					BeforeEach(func() {
						// 3rd reconcile
						res, err = r.Reconcile(req)
						foundPravega = &v1beta1.PravegaCluster{}
						err = client.Get(context.TODO(), req.NamespacedName, foundPravega)
					})

					It("shouldn't error", func() {
						Ω(err).Should(BeNil())
					})

					It("should create a statefulset", func() {
						foundSS := &appsv1.StatefulSet{}
						nn := types.NamespacedName{
							Name:      foundPravega.StatefulSetNameForSegmentstore(),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundSS)
						Ω(err).Should(BeNil())
					})

					It("should create a config-map", func() {
						foundCm := &corev1.ConfigMap{}
						nn := types.NamespacedName{
							Name:      foundPravega.ConfigMapNameForSegmentstore(),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundCm)
						Ω(err).Should(BeNil())
					})

					It("should create a headless-service", func() {
						foundSvc := &corev1.Service{}
						nn := types.NamespacedName{
							Name:      foundPravega.HeadlessServiceNameForSegmentStore(),
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
							Name:      foundPravega.ServiceNameForSegmentStore(0),
							Namespace: Namespace,
						}
						err = client.Get(context.TODO(), nn, foundSvc)
						fmt.Printf("client-services error: %v", err)
						Ω(err).Should(MatchError("services \"example-pravega-segment-store-0\" not found"))
					})
				})

				Context("checking updatePDB", func() {
					var (
						err1 error
						str1 string
					)
					BeforeEach(func() {
						res, err = r.Reconcile(req)
						p.WithDefaults()
						currentpdb := &policyv1beta1.PodDisruptionBudget{}
						r.client.Get(context.TODO(), types.NamespacedName{Name: p.PdbNameForSegmentstore(), Namespace: p.Namespace}, currentpdb)
						maxUnavailable := intstr.FromInt(3)
						newpdb := &policyv1beta1.PodDisruptionBudget{
							TypeMeta: metav1.TypeMeta{
								Kind:       "PodDisruptionBudget",
								APIVersion: "policy/v1beta1",
							},
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-name",
								Namespace: p.Namespace,
							},
							Spec: policyv1beta1.PodDisruptionBudgetSpec{
								MaxUnavailable: &maxUnavailable,
								Selector: &metav1.LabelSelector{
									MatchLabels: p.LabelsForController(),
								},
							},
						}
						err1 = r.updatePdb(currentpdb, newpdb)
						str1 = fmt.Sprintf("%s", currentpdb.Spec.MaxUnavailable)
					})
					It("should not give error", func() {
						Ω(err1).Should(BeNil())
					})
					It("unavailable replicas should change to 3", func() {
						Ω(str1).To(Equal("3"))
					})
				})

				Context("checking updateStatefulset", func() {
					var (
						err1 error
						str1 string
					)
					BeforeEach(func() {

						p.WithDefaults()
						sts := &appsv1.StatefulSet{}
						name := p.StatefulSetNameForSegmentstore()
						foundPravega := &v1beta1.PravegaCluster{}
						_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
						foundPravega.Spec.Pravega.SegmentStoreServiceAccountName = "testsa"
						client.Update(context.TODO(), foundPravega)
						_, _ = r.Reconcile(req)
						err1 = r.client.Get(context.TODO(),
							types.NamespacedName{Name: name, Namespace: p.Namespace}, sts)

						str1 = fmt.Sprintf("%s", sts.Spec.Template.Spec.ServiceAccountName)
					})
					It("should not give error", func() {
						Ω(err1).Should(BeNil())
					})
					It("service account name should get updated correctly", func() {
						Ω(str1).To(Equal("testsa"))
					})
				})

				Context("checking updateDeployment", func() {
					var (
						err1 error
						str1 string
					)
					BeforeEach(func() {

						p.WithDefaults()
						deploy := &appsv1.Deployment{}
						name := p.DeploymentNameForController()
						foundPravega := &v1beta1.PravegaCluster{}
						_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
						foundPravega.Spec.Pravega.ControllerServiceAccountName = "testsa"
						client.Update(context.TODO(), foundPravega)
						_, _ = r.Reconcile(req)
						err1 = r.client.Get(context.TODO(),
							types.NamespacedName{Name: name, Namespace: p.Namespace}, deploy)

						str1 = fmt.Sprintf("%s", deploy.Spec.Template.Spec.ServiceAccountName)
					})
					It("should not give error", func() {
						Ω(err1).Should(BeNil())
					})
					It("service account name should get updated correctly", func() {
						Ω(str1).To(Equal("testsa"))
					})
				})

				Context("checking checkVersionUpgradeTriggered function", func() {
					var (
						ans1, ans2 bool
					)
					BeforeEach(func() {
						client = fake.NewFakeClient(p)
						r = &ReconcilePravegaCluster{client: client, scheme: s}
						res, err = r.Reconcile(req)

						ans1 = r.checkVersionUpgradeTriggered(p)
						p.Spec.Version = "0.8.0"
						ans2 = r.checkVersionUpgradeTriggered(p)
					})
					It("ans1 should be false", func() {
						Ω(ans1).To(Equal(false))
					})
					It("ans2 should be true", func() {
						Ω(ans2).To(Equal(true))
					})

				})

				Context("reconcileFinalizers", func() {
					BeforeEach(func() {
						p.WithDefaults()
						client.Update(context.TODO(), p)
						err = r.reconcileFinalizers(p)
						now := metav1.Now()
						p.SetDeletionTimestamp(&now)
						client.Update(context.TODO(), p)
						err = r.reconcileFinalizers(p)
					})
					It("should give error due to failure in connecting to zookeeper", func() {
						Expect(err).To(HaveOccurred())
					})
				})

				Context("cleanUpZookeeperMeta", func() {
					BeforeEach(func() {
						p.WithDefaults()
						err = r.cleanUpZookeeperMeta(p)
					})
					It("should give error", func() {
						Ω(err).ShouldNot(BeNil())
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
				p.Spec = v1beta1.ClusterSpec{
					Version:       "0.3.2-rc2",
					BookkeeperUri: v1beta1.DefaultBookkeeperUri,
					Pravega: &v1beta1.PravegaSpec{
						ControllerReplicas:    2,
						SegmentStoreReplicas:  4,
						ControllerResources:   customReq,
						SegmentStoreResources: customReq,
						Image: &v1beta1.ImageSpec{
							Repository: "bar/pravega",
						},
						ControllerJvmOptions:   []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMFraction=1"},
						SegmentStoreJVMOptions: []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMFraction=1"},
					},
					TLS: &v1beta1.TLSPolicy{
						Static: &v1beta1.StaticTLS{
							ControllerSecret:   "controller-secret",
							SegmentStoreSecret: "segmentstore-secret",
						},
					},
				}
				//equivalent of 1st reconcile
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				// 2nd reconcile
				res, err = r.Reconcile(req)
			})

			It("shouldn't error", func() {
				Ω(err).Should(BeNil())
			})

			It("should requeue after ReconfileTime delay", func() {
				Ω(res.RequeueAfter).To(Equal(ReconcileTime))
			})

			It("should have a custom version", func() {
				foundPravega := &v1beta1.PravegaCluster{}
				err = client.Get(context.TODO(), req.NamespacedName, foundPravega)
				Ω(err).Should(BeNil())
				Ω(foundPravega.Spec.Version).Should(Equal("0.3.2-rc2"))
				Ω(foundPravega.Status.CurrentVersion).Should(Equal("0.3.2-rc2"))
			})

			Context("Pravega Controller", func() {
				var foundController *appsv1.Deployment
				BeforeEach(func() {
					foundController = &appsv1.Deployment{}
					nn := types.NamespacedName{
						Name:      p.DeploymentNameForController(),
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
					Ω(foundController.Spec.Template.Spec.Volumes[0].Name).Should(Equal("tls-secret"))
					Ω(foundController.Spec.Template.Spec.Volumes[0].VolumeSource.Secret.SecretName).Should(Equal("controller-secret"))
					Ω(foundController.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).Should(Equal("tls-secret"))
					Ω(foundController.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).Should(Equal("/etc/secret-volume"))
				})

				It("shoud overide pravega controller jvm options", func() {
					foundCm := &corev1.ConfigMap{}
					nn := types.NamespacedName{
						Name:      p.ConfigMapNameForController(),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn, foundCm)
					Ω(err).Should(BeNil())
					Ω(strings.Contains(foundCm.Data["JAVA_OPTS"], "-XX:MaxDirectMemorySize=1g")).Should(BeTrue())
					Ω(strings.Contains(foundCm.Data["JAVA_OPTS"], "-XX:MaxRAMFraction=1")).Should(BeTrue())
					Ω(strings.Contains(foundCm.Data["JAVA_OPTS"], "-XX:MaxRAMFraction=2")).Should(BeFalse())
				})
			})

			Context("Pravega SegmentStore", func() {
				var foundSS *appsv1.StatefulSet

				BeforeEach(func() {
					// 3rd reconcile
					res, err = r.Reconcile(req)
					foundSS = &appsv1.StatefulSet{}
					nn := types.NamespacedName{
						Name:      p.StatefulSetNameForSegmentstore(),
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
					Ω(foundSS.Spec.Template.Spec.Volumes[0].Name).Should(Equal("tls-secret"))
					Ω(foundSS.Spec.Template.Spec.Volumes[0].VolumeSource.Secret.SecretName).Should(Equal("segmentstore-secret"))
					Ω(foundSS.Spec.Template.Spec.Containers[0].VolumeMounts[1].Name).Should(Equal("tls-secret"))
					Ω(foundSS.Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath).Should(Equal("/etc/secret-volume"))
				})

				It("should overide pravega segmentstore jvm options", func() {
					foundCm := &corev1.ConfigMap{}
					nn := types.NamespacedName{
						Name:      p.ConfigMapNameForSegmentstore(),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn, foundCm)
					Ω(err).Should(BeNil())
					Ω(strings.Contains(foundCm.Data["JAVA_OPTS"], "-XX:MaxDirectMemorySize=1g")).Should(BeTrue())
					Ω(strings.Contains(foundCm.Data["JAVA_OPTS"], "-XX:MaxRAMFraction=1")).Should(BeTrue())
					Ω(strings.Contains(foundCm.Data["JAVA_OPTS"], "-XX:MaxRAMFraction=2")).Should(BeFalse())
				})
			})
		})

		Context("Custom spec with ExternalAccess", func() {
			var (
				client     client.Client
				err        error
				domainName string
			)

			BeforeEach(func() {
				domainName = "pravega.com."
				p.Spec = v1beta1.ClusterSpec{
					Version: "0.3.2-rc2",
					ExternalAccess: &v1beta1.ExternalAccess{
						Enabled:    true,
						Type:       corev1.ServiceTypeClusterIP,
						DomainName: domainName,
					},
					Pravega: &v1beta1.PravegaSpec{
						ControllerReplicas:   2,
						SegmentStoreReplicas: 3,
					},
				}
				// equivalent of 1st reconcile
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				// 2nd reconcile
				res, err = r.Reconcile(req)
			})

			It("shouldn't error", func() {
				Ω(err).Should(BeNil())
			})

			Context("Pravega Controller External Access", func() {
				var foundControllerSvc *corev1.Service

				BeforeEach(func() {
					foundControllerSvc = &corev1.Service{}
					nn := types.NamespacedName{
						Name:      p.ServiceNameForController(),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn, foundControllerSvc)
				})

				It("should create a controller service", func() {
					Ω(err).Should(BeNil())
				})

				It("should set external access service type to LoadBalancer", func() {
					Ω(p.Spec.ExternalAccess.Type).Should(Equal(corev1.ServiceTypeClusterIP))
					Ω(foundControllerSvc.Spec.Type).Should(Equal(corev1.ServiceTypeClusterIP))
				})

				It("should not set any annotations", func() {
					mapLength := len(foundControllerSvc.GetAnnotations())
					Ω(mapLength).To(Equal(0))
				})
			})

			Context("Pravega SegmentStore External Access", func() {
				var foundSegmentStoreSvc1 *corev1.Service
				var foundSegmentStoreSvc2 *corev1.Service
				var foundSegmentStoreSvc3 *corev1.Service

				BeforeEach(func() {
					// 2nd reconcile
					res, err = r.Reconcile(req)
					// 3rd reconcile
					res, err = r.Reconcile(req)

					foundSegmentStoreSvc1 = &corev1.Service{}
					nn1 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(0),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn1, foundSegmentStoreSvc1)

					foundSegmentStoreSvc2 = &corev1.Service{}
					nn2 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(1),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn2, foundSegmentStoreSvc2)

					foundSegmentStoreSvc3 = &corev1.Service{}
					nn3 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(2),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn3, foundSegmentStoreSvc3)

				})

				It("should create all segmentstore services", func() {
					Ω(err).Should(BeNil())
				})

				It("should set external access service type to ClusterIP for each service", func() {
					Ω(p.Spec.ExternalAccess.Type).Should(Equal(corev1.ServiceTypeClusterIP))
					Ω(foundSegmentStoreSvc1.Spec.Type).Should(Equal(corev1.ServiceTypeClusterIP))
					Ω(foundSegmentStoreSvc2.Spec.Type).Should(Equal(corev1.ServiceTypeClusterIP))
					Ω(foundSegmentStoreSvc3.Spec.Type).Should(Equal(corev1.ServiceTypeClusterIP))
				})

				It("should set only DNS name annotation", func() {
					mapLength := len(foundSegmentStoreSvc1.GetAnnotations())
					Ω(mapLength).To(Equal(1))

					svcName1 := p.ServiceNameForSegmentStore(0) + "." + domainName
					Expect(foundSegmentStoreSvc1.GetAnnotations()).To(HaveKeyWithValue(
						"external-dns.alpha.kubernetes.io/hostname", svcName1))

					svcName2 := p.ServiceNameForSegmentStore(1) + "." + domainName
					Expect(foundSegmentStoreSvc2.GetAnnotations()).To(HaveKeyWithValue(
						"external-dns.alpha.kubernetes.io/hostname", svcName2))

					svcName3 := p.ServiceNameForSegmentStore(2) + "." + domainName
					Expect(foundSegmentStoreSvc3.GetAnnotations()).To(HaveKeyWithValue(
						"external-dns.alpha.kubernetes.io/hostname", svcName3))
				})
			})
		})

		Context("Custom spec with ExternalAccess and changing the domainName", func() {
			var (
				client     client.Client
				err        error
				domainName string
			)

			BeforeEach(func() {
				domainName = "pravega.com."
				p.Spec = v1beta1.ClusterSpec{
					Version: "0.3.2-rc2",
					ExternalAccess: &v1beta1.ExternalAccess{
						Enabled:    true,
						Type:       corev1.ServiceTypeClusterIP,
						DomainName: domainName,
					},
					Pravega: &v1beta1.PravegaSpec{
						ControllerReplicas:   2,
						SegmentStoreReplicas: 3,
					},
				}
				// equivalent of 1st reconcile
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				// 2nd reconcile
				res, err = r.Reconcile(req)
			})

			Context("Pravega SegmentStore External Access", func() {
				var foundSegmentStoreSvc1 *corev1.Service

				BeforeEach(func() {
					// 2nd reconcile
					res, err = r.Reconcile(req)
					// 3rd reconcile
					res, err = r.Reconcile(req)
					domainName = "pravega1.com."
					foundPravega := &v1beta1.PravegaCluster{}
					_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
					foundPravega.Spec.ExternalAccess.DomainName = domainName
					client.Update(context.TODO(), foundPravega)
					// 4th reconcile
					_, _ = r.Reconcile(req)

					foundSegmentStoreSvc1 = &corev1.Service{}
					nn1 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(0),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn1, foundSegmentStoreSvc1)
				})

				It("should create all segmentstore services", func() {
					Ω(err).Should(BeNil())
				})

				It("should set only DNS name annotation", func() {
					mapLength := len(foundSegmentStoreSvc1.GetAnnotations())
					Ω(mapLength).To(Equal(1))

					svcName1 := p.ServiceNameForSegmentStore(0) + "." + domainName
					Expect(foundSegmentStoreSvc1.GetAnnotations()).To(HaveKeyWithValue(
						"external-dns.alpha.kubernetes.io/hostname", svcName1))
				})
			})
		})

		Context("Custom spec with ExternalAccess with annotations and overridden Service Type", func() {
			var (
				client     client.Client
				err        error
				domainName string
			)

			BeforeEach(func() {
				annotationsMap := map[string]string{
					"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
				}
				domainName = "pravega.com."
				p.Spec = v1beta1.ClusterSpec{
					Version: "0.3.2-rc2",
					ExternalAccess: &v1beta1.ExternalAccess{
						Enabled:    true,
						Type:       corev1.ServiceTypeClusterIP,
						DomainName: domainName,
					},
					Pravega: &v1beta1.PravegaSpec{
						ControllerReplicas:              2,
						SegmentStoreReplicas:            3,
						ControllerExternalServiceType:   corev1.ServiceTypeLoadBalancer,
						SegmentStoreExternalServiceType: corev1.ServiceTypeNodePort,
						ControllerServiceAnnotations:    annotationsMap,
						ControllerPodLabels:             annotationsMap,
						SegmentStoreServiceAnnotations:  annotationsMap,
						SegmentStorePodLabels:           annotationsMap,
					},
				}
				// equivalent of 1st reconcile
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				// 2nd reconcile
				res, err = r.Reconcile(req)
			})

			It("shouldn't error", func() {
				Ω(err).Should(BeNil())
			})

			Context("Pravega Controller External Access", func() {
				var foundControllerSvc *corev1.Service

				BeforeEach(func() {
					foundControllerSvc = &corev1.Service{}
					nn := types.NamespacedName{
						Name:      p.ServiceNameForController(),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn, foundControllerSvc)
				})

				It("should create a controller service", func() {
					Ω(err).Should(BeNil())
				})

				It("should set external access service type to LoadBalancer", func() {
					Ω(foundControllerSvc.Spec.Type).Should(Equal(corev1.ServiceTypeLoadBalancer))
				})

				It("should set provided annotations", func() {
					Expect(foundControllerSvc.GetAnnotations()).To(HaveKeyWithValue(
						"service.beta.kubernetes.io/aws-load-balancer-type",
						"nlb"))
				})
			})

			Context("Pravega SegmentStore External Access", func() {
				var foundSegmentStoreSvc1 *corev1.Service
				var foundSegmentStoreSvc2 *corev1.Service
				var foundSegmentStoreSvc3 *corev1.Service

				BeforeEach(func() {
					// 3rd reconcile
					res, err = r.Reconcile(req)

					foundSegmentStoreSvc1 = &corev1.Service{}
					nn1 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(0),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn1, foundSegmentStoreSvc1)

					foundSegmentStoreSvc2 = &corev1.Service{}
					nn2 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(1),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn2, foundSegmentStoreSvc2)

					foundSegmentStoreSvc3 = &corev1.Service{}
					nn3 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(2),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn3, foundSegmentStoreSvc3)

				})

				It("should create all segmentstore services", func() {
					Ω(err).Should(BeNil())
				})

				It("should set external access service type to NodePort for each service", func() {
					Ω(p.Spec.Pravega.SegmentStoreExternalServiceType).Should(Equal(corev1.ServiceTypeNodePort))
					Ω(p.Spec.ExternalAccess.Type).Should(Equal(corev1.ServiceTypeClusterIP))
					Ω(foundSegmentStoreSvc1.Spec.Type).Should(Equal(corev1.ServiceTypeNodePort))
					Ω(foundSegmentStoreSvc2.Spec.Type).Should(Equal(corev1.ServiceTypeNodePort))
					Ω(foundSegmentStoreSvc3.Spec.Type).Should(Equal(corev1.ServiceTypeNodePort))
				})

				It("should set provided annotations and DNS annotation", func() {
					mapLength := len(foundSegmentStoreSvc1.GetAnnotations())
					Ω(mapLength).To(Equal(2))
					Expect(foundSegmentStoreSvc1.GetAnnotations()).To(HaveKeyWithValue(
						"service.beta.kubernetes.io/aws-load-balancer-type",
						"nlb"))
					Expect(foundSegmentStoreSvc2.GetAnnotations()).To(HaveKeyWithValue(
						"service.beta.kubernetes.io/aws-load-balancer-type",
						"nlb"))
					Expect(foundSegmentStoreSvc3.GetAnnotations()).To(HaveKeyWithValue(
						"service.beta.kubernetes.io/aws-load-balancer-type",
						"nlb"))
					svcName1 := p.ServiceNameForSegmentStore(0) + "." + domainName
					Expect(foundSegmentStoreSvc1.GetAnnotations()).To(HaveKeyWithValue(
						"external-dns.alpha.kubernetes.io/hostname", svcName1))

					svcName2 := p.ServiceNameForSegmentStore(1) + "." + domainName
					Expect(foundSegmentStoreSvc2.GetAnnotations()).To(HaveKeyWithValue(
						"external-dns.alpha.kubernetes.io/hostname", svcName2))

					svcName3 := p.ServiceNameForSegmentStore(2) + "." + domainName
					Expect(foundSegmentStoreSvc3.GetAnnotations()).To(HaveKeyWithValue(
						"external-dns.alpha.kubernetes.io/hostname", svcName3))
				})
			})
		})

		Context("Custom spec with ExternalAccess with annotations and empty domain name", func() {
			var (
				client client.Client
				err    error
			)

			BeforeEach(func() {
				annotationsMap := map[string]string{
					"service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-prefix": "abc",
				}
				p.Spec = v1beta1.ClusterSpec{
					Version: "0.3.2-rc2",
					ExternalAccess: &v1beta1.ExternalAccess{
						Enabled: true,
						Type:    corev1.ServiceTypeClusterIP,
					},
					Pravega: &v1beta1.PravegaSpec{
						SegmentStoreReplicas:           3,
						SegmentStoreServiceAnnotations: annotationsMap,
					},
				}
				//equivalent 1st reconcile
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				// 2nd reconcile
				res, err = r.Reconcile(req)
			})

			It("shouldn't error", func() {
				Ω(err).Should(BeNil())
			})

			Context("Pravega SegmentStore External Access", func() {
				var foundSegmentStoreSvc1 *corev1.Service
				var foundSegmentStoreSvc2 *corev1.Service
				var foundSegmentStoreSvc3 *corev1.Service

				BeforeEach(func() {
					// 3rd reconcile
					res, err = r.Reconcile(req)
					foundSegmentStoreSvc1 = &corev1.Service{}
					nn1 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(0),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn1, foundSegmentStoreSvc1)

					foundSegmentStoreSvc2 = &corev1.Service{}
					nn2 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(1),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn2, foundSegmentStoreSvc2)

					foundSegmentStoreSvc3 = &corev1.Service{}
					nn3 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(2),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn3, foundSegmentStoreSvc3)

				})

				It("should create all segmentstore services", func() {
					Ω(err).Should(BeNil())
				})

				It("should set provided annotations", func() {
					mapLength := len(foundSegmentStoreSvc1.GetAnnotations())
					Ω(mapLength).To(Equal(1))
					Expect(foundSegmentStoreSvc1.GetAnnotations()).To(HaveKeyWithValue(
						"service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-prefix",
						"abc"))
					Expect(foundSegmentStoreSvc2.GetAnnotations()).To(HaveKeyWithValue(
						"service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-prefix",
						"abc"))
					Expect(foundSegmentStoreSvc3.GetAnnotations()).To(HaveKeyWithValue(
						"service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-prefix",
						"abc"))
				})
			})
		})

		Context("Custom spec with ExternalAccess with domain name and without other annotations", func() {
			var (
				client     client.Client
				err        error
				domainName string
			)

			BeforeEach(func() {
				domainName = "pravega.com."
				p.Spec = v1beta1.ClusterSpec{
					Version: "0.3.2-rc2",
					ExternalAccess: &v1beta1.ExternalAccess{
						Enabled:    true,
						Type:       corev1.ServiceTypeClusterIP,
						DomainName: domainName,
					},
					Pravega: &v1beta1.PravegaSpec{
						SegmentStoreReplicas: 3,
					},
				}
				// 1st reconcile
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				// 2nd reconcile
				res, err = r.Reconcile(req)
			})

			It("shouldn't error", func() {
				Ω(err).Should(BeNil())
			})

			Context("Pravega SegmentStore External Access", func() {
				var foundSegmentStoreSvc1 *corev1.Service
				var foundSegmentStoreSvc2 *corev1.Service
				var foundSegmentStoreSvc3 *corev1.Service

				BeforeEach(func() {
					// 3rd reconcile
					res, err = r.Reconcile(req)
					foundSegmentStoreSvc1 = &corev1.Service{}
					nn1 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(0),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn1, foundSegmentStoreSvc1)

					foundSegmentStoreSvc2 = &corev1.Service{}
					nn2 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(1),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn2, foundSegmentStoreSvc2)

					foundSegmentStoreSvc3 = &corev1.Service{}
					nn3 := types.NamespacedName{
						Name:      p.ServiceNameForSegmentStore(2),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn3, foundSegmentStoreSvc3)

				})

				It("should create all segmentstore services", func() {
					Ω(err).Should(BeNil())
				})

				It("should set provided domain name as annotation", func() {
					mapLength := len(foundSegmentStoreSvc1.GetAnnotations())
					Ω(mapLength).To(Equal(1))
					svcName1 := p.ServiceNameForSegmentStore(0) + "." + domainName
					Expect(foundSegmentStoreSvc1.GetAnnotations()).To(HaveKeyWithValue(
						"external-dns.alpha.kubernetes.io/hostname", svcName1))

					svcName2 := p.ServiceNameForSegmentStore(1) + "." + domainName
					Expect(foundSegmentStoreSvc2.GetAnnotations()).To(HaveKeyWithValue(
						"external-dns.alpha.kubernetes.io/hostname", svcName2))

					svcName3 := p.ServiceNameForSegmentStore(2) + "." + domainName
					Expect(foundSegmentStoreSvc3.GetAnnotations()).To(HaveKeyWithValue(
						"external-dns.alpha.kubernetes.io/hostname", svcName3))
				})
			})
		})

		Context("Custom spec with ExternalAccess with Segmentstore Scaledown", func() {
			var (
				client       client.Client
				err          error
				foundPravega *v1beta1.PravegaCluster
			)

			BeforeEach(func() {
				p.Spec = v1beta1.ClusterSpec{
					Version: v1beta1.DefaultPravegaVersion,
					ExternalAccess: &v1beta1.ExternalAccess{
						Enabled: true,
						Type:    corev1.ServiceTypeLoadBalancer,
					},
					Pravega: &v1beta1.PravegaSpec{
						SegmentStoreReplicas: 3,
					},
				}
				// 1st reconcile
				p.WithDefaults()
				client = fake.NewFakeClient(p)
				r = &ReconcilePravegaCluster{client: client, scheme: s}
				// 2nd reconcile
				res, _ = r.Reconcile(req)
				// 3rd reconcile
				_, _ = r.Reconcile(req)
				foundPravega = &v1beta1.PravegaCluster{}
				_ = client.Get(context.TODO(), req.NamespacedName, foundPravega)
			})

			Context("Scaledown Segmentstore Services", func() {

				var (
					foundSegmentStoreSvc1 *corev1.Service
					foundSegmentStoreSvc2 *corev1.Service
					foundSegmentStoreSvc3 *corev1.Service
				)

				BeforeEach(func() {
					foundPravega.Spec.Pravega.SegmentStoreReplicas = 1
					client.Update(context.TODO(), foundPravega)
					// 4th reconcile
					_, _ = r.Reconcile(req)
				})

				It("sts should be present", func() {
					foundSS := &appsv1.StatefulSet{}
					nn := types.NamespacedName{
						Name:      foundPravega.StatefulSetNameForSegmentstore(),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn, foundSS)
					Ω(err).Should(BeNil())
				})

				It("svc 1 should be found", func() {
					foundSegmentStoreSvc1 = &corev1.Service{}
					nn1 := types.NamespacedName{
						Name:      foundPravega.ServiceNameForSegmentStore(0),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn1, foundSegmentStoreSvc1)
					Ω(err).Should(BeNil())
				})

				It("svc2 should not be found", func() {
					foundSegmentStoreSvc2 = &corev1.Service{}
					nn2 := types.NamespacedName{
						Name:      foundPravega.ServiceNameForSegmentStore(2),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn2, foundSegmentStoreSvc2)
					Ω(err).ShouldNot(BeNil())
				})

				It("svc3 should not be found", func() {
					foundSegmentStoreSvc3 = &corev1.Service{}
					nn3 := types.NamespacedName{
						Name:      foundPravega.ServiceNameForSegmentStore(3),
						Namespace: Namespace,
					}
					err = client.Get(context.TODO(), nn3, foundSegmentStoreSvc3)
					Ω(err).ShouldNot(BeNil())
				})
			})
		})
	})
})
