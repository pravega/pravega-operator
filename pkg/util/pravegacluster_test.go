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
	"log"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFinalizer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pravegacluster")
}

var _ = Describe("pravegacluster", func() {

	Context("IsVersionBelow07", func() {
		var result1, result2 bool
		BeforeEach(func() {

			result1 = IsVersionBelow07("0.7.1")
			result2 = IsVersionBelow07("0.6.0")
			IsVersionBelow07("666")
			IsVersionBelow07("")
			input := []string{"0.4.0", "0.5.0"}
			ContainsVersion(input, "0.4.0")

		})
		It("should return true for result2", func() {
			Ω(result2).To(Equal(true))
		})
		It("should return false for result1", func() {
			Ω(result1).To(Equal(false))
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
		var result1, result2 bool
		BeforeEach(func() {

			result1 = IsOrphan("segment-store-4", 3)
			result2 = IsOrphan("segment-store-2", 3)
			IsOrphan("segmentstore", 1)
			IsOrphan("segment-store-1ab", 1)
		})
		It("should return true for result2", func() {
			Ω(result1).To(Equal(true))
		})
		It("should return false for result1", func() {
			Ω(result2).To(Equal(false))
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
			log.Printf("result is %v", result)
			result = RemoveString(result, "-XX:+ExitOnOutOfMemoryError")

		})
		It("should return true for result", func() {
			Ω(len(result)).ShouldNot(Equal(0))
			Ω(ContainsString(result, "-Xms1024m")).To(Equal(true))

		})
		It("should contain string ", func() {
			Ω(ContainsString(result, "-Xms512m")).To(Equal(false))
		})
		It("should contain string ", func() {
			Ω(ContainsString(result1, "-Xms512m")).To(Equal(true))
		})
		It("should contain string ", func() {
			Ω(ContainsString(result, "-yy:mem")).To(Equal(true))
		})
		It("should contain string ", func() {
			Ω(ContainsString(result, "-XX:+ExitOnOutOfMemoryError")).To(Equal(false))
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
	Context("Min()", func() {

		It("Min should be 10", func() {
			Ω(Min(10, 20)).Should(Equal(int32(10)))
			Ω(Min(30, 20)).Should(Equal(int32(20)))
		})

	})

	Context("podReady", func() {

		testpod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}},
			Status: v1.PodStatus{
				Conditions: []v1.PodCondition{
					{
						Type:   v1.PodReady,
						Status: v1.ConditionTrue,
					},
				}},
		}
		testpod2 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}},
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
		testpod3 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}},
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:  "test",
						State: v1.ContainerState{},
					},
				}},
		}
		testpod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}, Spec: v1.PodSpec{Containers: []v1.Container{{Image: "testimage"}}}}
		fault, _ := IsPodFaulty(testpod2)
		fault1, _ := IsPodFaulty(testpod3)
		It("pod ready should be true", func() {
			Ω(IsPodReady(testpod)).To(Equal(true))
			Ω(IsPodReady(testpod1)).To(Equal(false))
			Ω(GetPodVersion(testpod)).To(Equal(""))
			Ω(fault).To(Equal(true))
			Ω(fault1).To(Equal(false))

		})
	})

})
