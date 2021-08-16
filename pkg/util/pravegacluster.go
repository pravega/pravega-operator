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
	"reflect"
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
	healthcheckVersion      string = "0.10.0"
	MajorMinorVersionRegexp string = `^v?(?P<Version>[0-9]+\.[0-9]+\.[0-9]+)`
)

func init() {
	versionRegexp = regexp.MustCompile(MajorMinorVersionRegexp)
}

//function to check if v1 is below v2 or not
func IsVersionBelow(v1 string, v2 string) bool {
	if v1 == "" {
		return true
	}
	if v2 == "" {
		return false
	}
	result, _ := CompareVersions(v1, v2, "<")
	if result {
		return true
	}
	return false
}

func CompareConfigMap(cm1 *corev1.ConfigMap, cm2 *corev1.ConfigMap) bool {
	eq := reflect.DeepEqual(cm1.Data, cm2.Data)
	if eq {
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

func HealthcheckCommand(version string, port int32, restport int32) []string {
	command := ""
	if IsVersionBelow(version, healthcheckVersion) {
		command = fmt.Sprintf("netstat -ltn 2> /dev/null | grep %d || ss -ltn 2> /dev/null | grep %d", port, port)
	} else {
		command = fmt.Sprintf("curl -s -X GET 'http://localhost:%d/v1/health/liveness' || curl -s -k -X GET 'https://localhost:%d/v1/health/liveness'", restport, restport)
	}
	return []string{"/bin/sh", "-c", command}
}

func ControllerReadinessCheck(version string, port int32, authflag bool) []string {
	command := ""
	if IsVersionBelow(version, healthcheckVersion) {
		//This function check for the readiness of the controller in the following cases
		//1) Auth and TLS Enabled- in this case, we check if the controller is properly enabled with authentication or not and we do a get on controller and with dummy credentials(testtls:testtls) and the controller returns 401 error in this case if it's correctly configured
		//2) Auth Enabled and TLS Disabled- in this case, we check if the controller is properly enabled with authentication or not and we do a get on controller and with dummy credentials(testtls:testtls) and the controller returns 401 error in this case if it's correctly configured
		//3) Auth Disabled and TLS Enabled- in this case, we check if the controller can create scopes or not by checking if _system scope is present or not
		//4) Auth and TLS Disabled- in this case, we check if the controller can create scopes or not by checking if _system scope is present or not
		if authflag == true {
			// This is to check the readiness of controller in case auth is Enabled
			// here we are using login credential as testtls:testtls which should
			// not be used as auth credential and we depend on controller giving us
			// 401 error which means controller is properly configured with auth
			// it checks both cases when tls is enabled as well as tls disabled
			// with auth enabled
			command = fmt.Sprintf("echo $JAVA_OPTS | grep 'controller.auth.tlsEnabled=true' &&  curl -v -k -u testtls:testtls -s -X GET 'https://localhost:%d/v1/scopes/' 2>&1 -H 'accept: application/json' | grep 401 || (echo $JAVA_OPTS | grep 'controller.auth.tlsEnabled=false' && curl -v -k -u testtls:testtls -s -X GET 'http://localhost:%d/v1/scopes/' 2>&1 -H 'accept: application/json' | grep 401 ) ||  (echo $JAVA_OPTS | grep 'controller.security.tls.enable=true' && echo $JAVA_OPTS | grep -v 'controller.auth.tlsEnabled' && curl -v -k -u testtls:testtls -s -X GET 'https://localhost:%d/v1/scopes/' 2>&1 -H 'accept: application/json' | grep 401 ) || (curl -v -k -u testtls:testtls -s -X GET 'http://localhost:%d/v1/scopes/' 2>&1 -H 'accept: application/json' | grep 401 )", port, port, port, port)
		} else {
			// This is to check the readiness in case auth is not enabled
			// and it covers both the cases with tls enabled and tls disabled
			// along with auth disabled
			command = fmt.Sprintf("echo $JAVA_OPTS | grep 'controller.auth.tlsEnabled=true' &&  curl -s -X GET 'https://localhost:%d/v1/scopes/' -H 'accept: application/json' | grep '_system'|| (echo $JAVA_OPTS | grep 'controller.auth.tlsEnabled=false' && curl -s -X GET 'http://localhost:%d/v1/scopes/' -H 'accept: application/json' | grep '_system' ) || (echo $JAVA_OPTS | grep 'controller.security.tls.enable=true' && echo $JAVA_OPTS | grep -v 'controller.auth.tlsEnabled' && curl -s -X GET 'https://localhost:%d/v1/scopes/' -H 'accept: application/json' | grep '_system' ) || (curl -s -X GET 'http://localhost:%d/v1/scopes/' -H 'accept: application/json' | grep '_system') ", port, port, port, port)
		}
	} else {
		command = fmt.Sprintf("curl -s -X GET 'http://localhost:%d/v1/health/readiness' || curl -s -k -X GET 'https://localhost:%d/v1/health/readiness'", port, port)
	}
	return []string{"/bin/sh", "-c", command}
}

func SegmentStoreReadinessCheck(version string, port int32, restport int32) []string {
	command := ""
	if IsVersionBelow(version, healthcheckVersion) {
		command = fmt.Sprintf("netstat -ltn 2> /dev/null | grep %d || ss -ltn 2> /dev/null | grep %d", port, port)
	} else {
		command = fmt.Sprintf("curl -s -X GET 'http://localhost:%d/v1/health/readiness' || curl -s -k -X GET 'https://localhost:%d/v1/health/readiness'", restport, restport)
	}
	return []string{"/bin/sh", "-c", command}
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

func IsPravegaContainer(container corev1.ContainerStatus) bool {
	return container.Name == "pravega-controller" || container.Name == "pravega-segmentstore"
}

func IsPodFaulty(pod *corev1.Pod) (bool, error) {
	for _, container := range pod.Status.ContainerStatuses {
		if IsPravegaContainer(container) && container.State.Waiting != nil && (container.State.Waiting.Reason == "ImagePullBackOff" ||
			container.State.Waiting.Reason == "CrashLoopBackOff") {
			return true, fmt.Errorf("pod %s update failed because of %s", pod.Name, container.State.Waiting.Reason)
		}
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
