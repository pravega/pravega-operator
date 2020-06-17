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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPravegacluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pravegacluster")
}

var _ = Describe("pravegacluster", func() {

	Context("IsVersionBelow07", func() {
		var result1, result2, result3, result4 bool
		BeforeEach(func() {
			result1 = IsVersionBelow07("0.7.1")
			result2 = IsVersionBelow07("0.6.0")
			result3 = IsVersionBelow07("666")
			result4 = IsVersionBelow07("")
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

		out := HealthcheckCommand(1234)
		It("should not be nil", func() {
			Ω(len(out)).ShouldNot(Equal(0))
		})

	})
	Context("ReadinessHealthcheckCommand()", func() {
		out := ReadinessHealthcheckCommand(1234)
		It("Should not be Empty", func() {
			Ω(len(out)).ShouldNot(Equal(0))
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
							Name: "test",
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
							Name:  "test",
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
