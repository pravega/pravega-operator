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
	"github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"k8s.io/api/core/v1"
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

func ConfigMapNameForBookieVersions(clusterName string) string {
	return fmt.Sprintf("%s-supported-upgrade-paths", clusterName)
}

func PdbNameForBookie(clusterName string) string {
	return fmt.Sprintf("%s-bookie", clusterName)
}

func ConfigMapNameForBookie(clusterName string) string {
	return fmt.Sprintf("%s-bookie", clusterName)
}

func StatefulSetNameForBookie(clusterName string) string {
	return fmt.Sprintf("%s-bookie", clusterName)
}

func HeadlessServiceNameForBookie(clusterName string) string {
	return fmt.Sprintf("%s-bookie-headless", clusterName)
}

func LabelsForBookie(bookkeeperCluster *v1alpha1.BookkeeperCluster) map[string]string {
	labels := LabelsForBookkeeperCluster(bookkeeperCluster)
	labels["component"] = "bookie"
	return labels
}

func LabelsForBookkeeperCluster(bookkeeperCluster *v1alpha1.BookkeeperCluster) map[string]string {
	return map[string]string{
		"app":                "bookkeeper-cluster",
		"bookkeeper_cluster": bookkeeperCluster.Name,
	}
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

func GetClusterExpectedSize(p *v1alpha1.BookkeeperCluster) (size int) {
	return int(p.Spec.Replicas)
}

func BookkeeperImage(p *v1alpha1.BookkeeperCluster) (image string) {
	return fmt.Sprintf("%s:%s", p.Spec.Image.Repository, p.Spec.Version)
}

func BookkeeperTargetImage(p *v1alpha1.BookkeeperCluster) (string, error) {
	if p.Status.TargetVersion == "" {
		return "", fmt.Errorf("target version is not set")
	}
	return fmt.Sprintf("%s:%s", p.Spec.Image.Repository, p.Status.TargetVersion), nil
}

func GetPodVersion(pod *v1.Pod) string {
	return pod.GetAnnotations()["bookkeeper.version"]
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
