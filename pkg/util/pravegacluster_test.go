/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */
package util

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPravegacluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pravegacluster")
}

var _ = Describe("pravegacluster", func() {

	Context("IsVersionBelow", func() {
		var result1, result2, result3, result4, result5 bool
		BeforeEach(func() {
			result1 = IsVersionBelow("0.7.1", "0.7.0")
			result2 = IsVersionBelow("0.6.0", "0.7.0")
			result3 = IsVersionBelow("666", "0.7.0")
			result4 = IsVersionBelow("", "0.7.0")
			result5 = IsVersionBelow("0.6.5", "")
		})
		It("should return false for result1", func() {
			Ω(result1).To(Equal(false))
		})
		It("should return true for result2", func() {
			Ω(result2).To(Equal(true))
		})
		It("should return false for result3", func() {
			Ω(result3).To(Equal(false))
		})
		It("should return true for result4", func() {
			Ω(result4).To(Equal(true))
		})
		It("should return false for result5", func() {
			Ω(result5).To(Equal(false))
		})
	})

	Context("ContainsVersion fn", func() {
		var result1, result2 bool
		BeforeEach(func() {
			input := []string{"0.4.0", "0.5.0", "a.b.c"}
			result1 = ContainsVersion(input, "0.4.0")
			result2 = ContainsVersion(input, "0.7.0")
		})
		It("should return true for result", func() {
			Ω(result1).To(Equal(true))
		})
		It("should return true for result", func() {
			Ω(result2).To(Equal(false))
		})
	})

	Context("IsOrphan", func() {
		var result1, result2, result3, result4 bool
		BeforeEach(func() {
			result1 = IsOrphan("segment-store-4", 3)
			result2 = IsOrphan("segment-store-2", 3)
			result3 = IsOrphan("segmentstore", 1)
			result4 = IsOrphan("segment-store-1ab", 1)
		})
		It("should return true for result2", func() {
			Ω(result1).To(Equal(true))
		})
		It("should return false for result1", func() {
			Ω(result2).To(Equal(false))
		})
		It("should return false for result3", func() {
			Ω(result3).To(Equal(false))
		})
		It("should return false for result4", func() {
			Ω(result4).To(Equal(false))
		})
	})

	Context("OverrideDefaultJVMOptions", func() {
		var result, result1 []string
		BeforeEach(func() {
			jvmOpts := []string{
				"-Xms512m",
				"-XX:+ExitOnOutOfMemoryError",
				"-XX:+CrashOnOutOfMemoryError",
				"-XX:+HeapDumpOnOutOfMemoryError",
				"-XX:HeapDumpPath=/heap",
			}
			customOpts := []string{
				"-Xms1024m",
				"-XX:+ExitOnOutOfMemoryError",
				"-XX:+CrashOnOutOfMemoryError",
				"-XX:+HeapDumpOnOutOfMemoryError",
				"-XX:HeapDumpPath=/heap",
				"-yy:mem",
				"",
			}
			result = OverrideDefaultJVMOptions(jvmOpts, customOpts)
			result1 = OverrideDefaultJVMOptions(jvmOpts, result1)
		})
		It("should contain string", func() {
			Ω(len(result)).ShouldNot(Equal(0))
			Ω(result[0]).To(Equal("-Xms1024m"))
			Ω(result1[0]).To(Equal("-Xms512m"))
		})
	})

	Context("RemoveString", func() {
		var opts []string
		BeforeEach(func() {
			opts = []string{
				"abc-test",
				"test1",
			}
			opts = RemoveString(opts, "abc-test")
		})
		It("should return false for result", func() {
			Ω(opts[0]).To(Equal("test1"))
		})
	})

	Context("ContainsString", func() {
		var result, result1 bool
		BeforeEach(func() {
			opts := []string{
				"abc-test",
				"test1",
			}
			result = ContainsString(opts, "abc-test")
			result1 = ContainsString(opts, "abc-test1")
		})

		It("should return true", func() {
			Ω(result).To(Equal(true))
		})

		It("should return false", func() {
			Ω(result1).To(Equal(false))
		})
	})

	Context("PodAntiAffinity", func() {
		affinity := PodAntiAffinity("segstore", "pravega")
		It("should not be nil", func() {
			Ω(affinity).ShouldNot(BeNil())
		})
	})

	Context("DownwardAPIEnv()", func() {
		env := DownwardAPIEnv()
		It("should not be nil", func() {
			Ω(env).ShouldNot(BeNil())
		})
	})

	Context("HealthcheckCommand()", func() {
		var r1, r2 []string
		BeforeEach(func() {
			r1 = HealthcheckCommand("0.9.0", 1234, 6061)
			r2 = HealthcheckCommand("0.10.0", 1234, 6061)
		})
		It("should not be nil", func() {
			Ω(len(r1)).ShouldNot(Equal(0))
		})
		It("should not be nil", func() {
			Ω(len(r2)).ShouldNot(Equal(0))
		})
	})

	Context("ControllerReadinessCheck()", func() {
		var r1, r2, r3 []string
		BeforeEach(func() {
			r1 = ControllerReadinessCheck("0.9.0", 1234, true)
			r2 = ControllerReadinessCheck("0.9.0", 1234, false)
			r3 = ControllerReadinessCheck("0.10.0", 1234, true)
		})
		It("Should not be Empty", func() {
			Ω(len(r1)).ShouldNot(Equal(0))
		})
		It("Should not be Empty", func() {
			Ω(len(r2)).ShouldNot(Equal(0))
		})
		It("Should not be Empty", func() {
			Ω(len(r3)).ShouldNot(Equal(0))
		})
	})

	Context("SegmentStoreReadinessCheck()", func() {
		var r1, r2 []string
		BeforeEach(func() {
			r1 = SegmentStoreReadinessCheck("0.9.0", 1234, 6061)
			r2 = SegmentStoreReadinessCheck("0.10.0", 1234, 6061)
		})
		It("Should not be Empty", func() {
			Ω(len(r1)).ShouldNot(Equal(0))
		})
		It("Should not be Empty", func() {
			Ω(len(r2)).ShouldNot(Equal(0))
		})
	})

	Context("Min()", func() {
		It("Min should be 10", func() {
			Ω(Min(10, 20)).Should(Equal(int32(10)))
		})
		It("Min should be 20", func() {
			Ω(Min(30, 20)).Should(Equal(int32(20)))
		})
	})

	Context("podReady", func() {
		var result, result1 bool
		BeforeEach(func() {
			testpod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}},
				Status: v1.PodStatus{
					Conditions: []v1.PodCondition{
						{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						},
					}},
			}
			testpod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}}}
			result = IsPodReady(testpod)
			result1 = IsPodReady(testpod1)
		})
		It("pod ready should be true", func() {
			Ω(result).To(Equal(true))

		})
		It("pod ready should be false", func() {
			Ω(result1).To(Equal(false))

		})
	})

	Context("podFaulty", func() {
		var result, result1 bool
		BeforeEach(func() {
			testpod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}},
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "pravega-segmentstore",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason: "CrashLoopBackOff",
								},
							},
						},
					}},
			}
			testpod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}},
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:  "pravega-segmentstore",
							State: v1.ContainerState{},
						},
					}},
			}
			result, _ = IsPodFaulty(testpod)
			result1, _ = IsPodFaulty(testpod1)
		})
		It("pod faulty should be true", func() {
			Ω(result).To(Equal(true))

		})
		It("pod faulty should be false", func() {
			Ω(result1).To(Equal(false))

		})
	})

	Context("CompareConfigMap", func() {
		var output1, output2 bool
		BeforeEach(func() {
			configData1 := map[string]string{
				"TEST_DATA": "testdata",
			}
			configData2 := map[string]string{
				"TEST_DATA": "testdata1",
			}
			configMap1 := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				Data: configData1,
			}
			configMap2 := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				Data: configData1,
			}
			configMap3 := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				Data: configData2,
			}
			output1 = CompareConfigMap(configMap1, configMap2)
			output2 = CompareConfigMap(configMap1, configMap3)
		})

		It("output1 should be true", func() {
			Ω(output1).To(Equal(true))
		})
		It("output2 should be false", func() {
			Ω(output2).To(Equal(false))
		})
	})

	Context("GetPodVersion", func() {
		var out string
		BeforeEach(func() {
			annotationsMap := map[string]string{
				"pravega.version": "0.7.0",
			}

			testpod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Annotations: annotationsMap}}
			out = GetPodVersion(testpod)
		})
		It("should return correct version", func() {
			Ω(out).To(Equal("0.7.0"))
		})
	})
})
