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
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
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

func PdbNameForBookie(clusterName string) string {
	return fmt.Sprintf("%s-bookie", clusterName)
}

func ConfigMapNameForBookie(clusterName string) string {
	return fmt.Sprintf("%s-bookie", clusterName)
}

func StatefulSetNameForBookie(clusterName string) string {
	return fmt.Sprintf("%s-bookie", clusterName)
}

func PdbNameForController(clusterName string) string {
	return fmt.Sprintf("%s-pravega-controller", clusterName)
}

func ConfigMapNameForController(clusterName string) string {
	return fmt.Sprintf("%s-pravega-controller", clusterName)
}

func ServiceNameForController(clusterName string) string {
	return fmt.Sprintf("%s-pravega-controller", clusterName)
}

func ServiceNameForSegmentStore(clusterName string, index int32) string {
	return fmt.Sprintf("%s-pravega-segmentstore-%d", clusterName, index)
}

func HeadlessServiceNameForSegmentStore(clusterName string) string {
	return fmt.Sprintf("%s-pravega-segmentstore-headless", clusterName)
}

func HeadlessServiceNameForBookie(clusterName string) string {
	return fmt.Sprintf("%s-bookie-headless", clusterName)
}

func DeploymentNameForController(clusterName string) string {
	return fmt.Sprintf("%s-pravega-controller", clusterName)
}

func PdbNameForSegmentstore(clusterName string) string {
	return fmt.Sprintf("%s-segmentstore", clusterName)
}

func ConfigMapNameForSegmentstore(clusterName string) string {
	return fmt.Sprintf("%s-pravega-segmentstore", clusterName)
}

//function to check if the version is below 0.7 or not
func IsVersionBelow07(ver string) bool {
	first3 := strings.Trim(ver, "\t \n")[0:3]
	if first3 == "0.6" || first3 == "0.5" || first3 == "0.4" || first3 == "0.3" || first3 == "0.2" || first3 == "0.1" {
		return true
	}
	return false
}

func IsClusterUpgradingTo07(p *api.PravegaCluster) bool {
	/*this check is to avoid creation of a new segmentstore when the CurrentVersionis 06 and target version is 07
	  as we are doing it in the upgrade path*/
	if !IsVersionBelow07(p.Spec.Version) && IsVersionBelow07(p.Status.CurrentVersion) {
		return true
	}
	return false
}

//if version is above or equals to 0.7 this name will be assigned
func StatefulSetNameForSegmentstoreAbove07(name string) string {
	return fmt.Sprintf("%s-pravega-above-version-07-segmentstore", name)
}

//if version is below 0.7 this name will be assigned
func StatefulSetNameForSegmentstoreBelow07(name string) string {
	return fmt.Sprintf("%s-pravega-segmentstore", name)
}

//to return name of segmentstore based on the version
func StatefulSetNameForSegmentstore(p *api.PravegaCluster) string {
	if p.Spec.Version == "" || IsVersionBelow07(p.Spec.Version) {
		StatefulSetNameForSegmentstoreBelow07(p.Name)
	}
	return StatefulSetNameForSegmentstoreAbove07(p.Name)
}

func LabelsForBookie(pravegaCluster *v1alpha1.PravegaCluster) map[string]string {
	labels := LabelsForPravegaCluster(pravegaCluster)
	labels["component"] = "bookie"
	return labels
}

func LabelsForController(pravegaCluster *v1alpha1.PravegaCluster) map[string]string {
	labels := LabelsForPravegaCluster(pravegaCluster)
	labels["component"] = "pravega-controller"
	return labels
}

func LabelsForSegmentStore(pravegaCluster *v1alpha1.PravegaCluster) map[string]string {
	labels := LabelsForPravegaCluster(pravegaCluster)
	labels["component"] = "pravega-segmentstore"
	return labels
}

func LabelsForPravegaCluster(pravegaCluster *v1alpha1.PravegaCluster) map[string]string {
	return map[string]string{
		"app":             "pravega-cluster",
		"pravega_cluster": pravegaCluster.Name,
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

func PravegaControllerServiceURL(pravegaCluster v1alpha1.PravegaCluster) string {
	return fmt.Sprintf("tcp://%v.%v:%v", ServiceNameForController(pravegaCluster.Name), pravegaCluster.Namespace, "9090")
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

func GetClusterExpectedSize(p *v1alpha1.PravegaCluster) (size int) {
	return int(p.Spec.Pravega.ControllerReplicas + p.Spec.Pravega.SegmentStoreReplicas + p.Spec.Bookkeeper.Replicas)
}

func PravegaImage(p *v1alpha1.PravegaCluster) (image string) {
	return fmt.Sprintf("%s:%s", p.Spec.Pravega.Image.Repository, p.Spec.Version)
}

func BookkeeperImage(p *v1alpha1.PravegaCluster) (image string) {
	return fmt.Sprintf("%s:%s", p.Spec.Bookkeeper.Image.Repository, p.Spec.Version)
}

func PravegaTargetImage(p *v1alpha1.PravegaCluster) (string, error) {
	if p.Status.TargetVersion == "" {
		return "", fmt.Errorf("target version is not set")
	}
	return fmt.Sprintf("%s:%s", p.Spec.Pravega.Image.Repository, p.Status.TargetVersion), nil
}

func BookkeeperTargetImage(p *v1alpha1.PravegaCluster) (string, error) {
	if p.Status.TargetVersion == "" {
		return "", fmt.Errorf("target version is not set")
	}
	return fmt.Sprintf("%s:%s", p.Spec.Bookkeeper.Image.Repository, p.Status.TargetVersion), nil
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
