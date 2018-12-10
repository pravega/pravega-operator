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

func ConfigMapNameForBookie(clusterName string) string {
	return fmt.Sprintf("%s-bookie", clusterName)
}

func StatefulSetNameForBookie(clusterName string) string {
	return fmt.Sprintf("%s-bookie", clusterName)
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

func PvcNameForSts(pvcName string, stsName string) string {
	return fmt.Sprintf("%s-%s", pvcName, stsName)
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

func ConfigMapNameForSegmentstore(clusterName string) string {
	return fmt.Sprintf("%s-pravega-segmentstore", clusterName)
}

func StatefulSetNameForSegmentstore(clusterName string) string {
	return fmt.Sprintf("%s-pravega-segmentstore", clusterName)
}

func LabelsForBookie(pravegaCluster *v1alpha1.PravegaCluster) map[string]string {
	return LabelsForPravegaCluster(pravegaCluster, "bookie")
}

func LabelsForController(pravegaCluster *v1alpha1.PravegaCluster) map[string]string {
	return LabelsForPravegaCluster(pravegaCluster, "pravega-controller")
}

func LabelsForSegmentStore(pravegaCluster *v1alpha1.PravegaCluster) map[string]string {
	return LabelsForPravegaCluster(pravegaCluster, "pravega-segmentstore")
}

func LabelsForPravegaCluster(pravegaCluster *v1alpha1.PravegaCluster, component string) map[string]string {
	return map[string]string{
		"app":             "pravega-cluster",
		"pravega_cluster": pravegaCluster.Name,
		"component":       component,
	}
}

func PravegaControllerServiceURL(pravegaCluster v1alpha1.PravegaCluster) string {
	return fmt.Sprintf("tcp://%v.%v:%v", ServiceNameForController(pravegaCluster.Name), pravegaCluster.Namespace, "9090")
}

func PvcIsOrphan(pvcNameForK8s string, pvcMap map[string]int) bool {
	index := strings.LastIndexAny(pvcNameForK8s, "-")
	if index == -1 {
		return false
	}

	name := pvcNameForK8s[:index]
	if replica, ok := pvcMap[name]; ok {
		ordinal, err := strconv.Atoi(pvcNameForK8s[index+1:])
		if err != nil {
			return false
		}

		if ordinal >= replica {
			return true
		}
	}
	return false
}
