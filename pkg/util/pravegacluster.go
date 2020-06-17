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
	"fmt"
	"regexp"
	"strconv"
	"strings"

	v "github.com/hashicorp/go-version"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	versionRegexp *regexp.Regexp
)

const (
	MajorMinorVersionRegexp string = `^v?(?P<Version>[0-9]+\.[0-9]+\.[0-9]+)`
)

func init() {
	versionRegexp = regexp.MustCompile(MajorMinorVersionRegexp)
}

//function to check if the version is below 0.7 or not
func IsVersionBelow07(ver string) bool {
	if ver == "" {
		return true
	}
	result, _ := CompareVersions(ver, "0.7.0", "<")
	if result {
		return true
	}
	return false
}

func IsOrphan(k8sObjectName string, replicas int32) bool {
	index := strings.LastIndexAny(k8sObjectName, "-")
	if index == -1 {
		return false
	}

	ordinal, err := strconv.Atoi(k8sObjectName[index+1:])
	if err != nil {
		return false
	}

	return int32(ordinal) >= replicas
}

func HealthcheckCommand(port int32) []string {
	return []string{"/bin/sh", "-c", fmt.Sprintf("netstat -ltn 2> /dev/null | grep %d || ss -ltn 2> /dev/null | grep %d", port, port)}
}

func ControllerReadinessCheck(port int32) []string {
	return []string{"/bin/sh", "-c", fmt.Sprintf("curl -s -X GET 'http://localhost:%d/v1/scopes/' -H 'accept: application/json' | grep 'system'", port)}
}

// Min returns the smaller of x or y.
func Min(x, y int32) int32 {
	if x > y {
		return y
	}
	return x
}

func ContainsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, str string) (result []string) {
	for _, item := range slice {
		if item == str {
			continue
		}
		result = append(result, item)
	}
	return result
}

func GetPodVersion(pod *v1.Pod) string {
	return pod.GetAnnotations()["pravega.version"]
}

func CompareVersions(v1, v2, operator string) (bool, error) {
	normv1, err := NormalizeVersion(v1)
	if err != nil {
		return false, err
	}
	normv2, err := NormalizeVersion(v2)
	if err != nil {
		return false, err
	}
	clusterVersion, err := v.NewSemver(normv1)
	if err != nil {
		return false, err
	}
	constraints, err := v.NewConstraint(fmt.Sprintf("%s %s", operator, normv2))
	if err != nil {
		return false, err
	}
	return constraints.Check(clusterVersion), nil
}

func ContainsVersion(list []string, version string) bool {
	result := false
	for _, v := range list {
		if result, _ = CompareVersions(version, v, "="); result {
			break
		}
	}
	return result
}

func NormalizeVersion(version string) (string, error) {
	matches := versionRegexp.FindStringSubmatch(version)
	if matches == nil || len(matches) <= 1 {
		return "", fmt.Errorf("failed to parse version %s", version)
	}
	return matches[1], nil
}

// OrderedMap is a map that has insertion order when iterating. The iteration of
// map in GO is in random order by default.
type OrderedMap struct {
	m    map[string]string
	keys []string
}

// This method will parse the JVM options into a key value pair and store it
// in the OrderedMap
func UpdateOneJVMOption(arg string, om *OrderedMap) {
	// Parse "-Xms"
	if strings.HasPrefix(arg, "-Xms") {
		if _, ok := om.m["-Xms"]; !ok {
			om.keys = append(om.keys, "-Xms")
		}
		om.m["-Xms"] = arg[4:]
		return
	}

	// Parse option starting with "-XX"
	if strings.HasPrefix(arg, "-XX:") {
		if arg[4] == '+' || arg[4] == '-' {
			if _, ok := om.m[arg[5:]]; !ok {
				om.keys = append(om.keys, arg[5:])
			}
			om.m[arg[5:]] = string(arg[4])
			return
		}
		s := strings.Split(arg[4:], "=")
		if _, ok := om.m[s[0]]; !ok {
			om.keys = append(om.keys, s[0])
		}
		om.m[s[0]] = s[1]
		return
	}

	// Not in those formats, just keep the option as a key
	if _, ok := om.m[arg]; !ok {
		om.keys = append(om.keys, arg)
	}
	om.m[arg] = ""
	return
}

// Concatenate the key value pair to be a JVM option string.
func GenerateJVMOption(k, v string) string {
	if v == "" {
		return k
	}

	if k == "-Xms" {
		return fmt.Sprintf("%v%v", k, v)
	}

	if v == "+" || v == "-" {
		return fmt.Sprintf("-XX:%v%v", v, k)
	}

	return fmt.Sprintf("-XX:%v=%v", k, v)
}

// This method will override the default JVM options with user provided custom options
func OverrideDefaultJVMOptions(defaultOpts []string, customOpts []string) []string {

	// Nothing to be overriden, just return the default options
	if customOpts == nil {
		return defaultOpts
	}

	om := &OrderedMap{m: map[string]string{}, keys: []string{}}

	// Firstly, store the default options in an ordered map. The ordered map is a
	// map that has insertion order guarantee when iterating.
	for _, option := range defaultOpts {
		UpdateOneJVMOption(option, om)
	}

	// Secondly, update the ordered map with custom options. If the option has been
	// found in the map, its value will be updated by the custom options. If not, the
	// the map will just add a new key value pair.
	for _, option := range customOpts {
		UpdateOneJVMOption(option, om)
	}

	jvmOpts := []string{}
	// Iterate the ordered map in its insertion order.
	for _, key := range om.keys {
		jvmOpts = append(jvmOpts, GenerateJVMOption(key, om.m[key]))
	}

	return jvmOpts
}

func IsPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsPodFaulty(pod *corev1.Pod) (bool, error) {
	if pod.Status.ContainerStatuses[0].State.Waiting != nil && (pod.Status.ContainerStatuses[0].State.Waiting.Reason == "ImagePullBackOff" ||
		pod.Status.ContainerStatuses[0].State.Waiting.Reason == "CrashLoopBackOff") {
		return true, fmt.Errorf("pod %s update failed because of %s", pod.Name, pod.Status.ContainerStatuses[0].State.Waiting.Reason)
	}
	return false, nil
}

func DownwardAPIEnv() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
	}
}

func PodAntiAffinity(component string, clusterName string) *corev1.Affinity {
	return &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "component",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{component},
								},
								{
									Key:      "pravega_cluster",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{clusterName},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}
}
