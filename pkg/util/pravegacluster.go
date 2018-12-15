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
	"strconv"
	"strings"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
)

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
