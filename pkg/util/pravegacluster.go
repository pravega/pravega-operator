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
	"k8s.io/api/core/v1"
)

var (
	versionRegexp *regexp.Regexp
)

const (
	MajorMinorVersionRegexp string = `^v?(?P<Version>[0-9]+\.[0-9]+)`
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

func StatefulSetNameForSegmentstore(clusterName string) string {
	return fmt.Sprintf("%s-pravega-segmentstore", clusterName)
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

func PvcIsOrphan(stsPvcName string, replicas int32) bool {
	index := strings.LastIndexAny(stsPvcName, "-")
	if index == -1 {
		return false
	}

	ordinal, err := strconv.Atoi(stsPvcName[index+1:])
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

func GetPodVersion(pod *v1.Pod) string {
	return pod.GetAnnotations()["pravega.version"]
}

func CompareVersions(v1, v2, operator string) (bool, error) {
	clusterVersion, _ := v.NewSemver(normalizeVersion(v1))
	constraints, err := v.NewConstraint(fmt.Sprintf("%s %s", operator, v2))
	if err != nil {
		return false, err
	}
	return constraints.Check(clusterVersion), nil
}

func normalizeVersion(version string) string {
	matches := versionRegexp.FindStringSubmatch(version)
	if matches == nil || len(matches) <= 1 {
		// Assume that version is the latest release
		return "0.5"
	}
	return matches[1]
}
