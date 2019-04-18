/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package charts

import (
	"strconv"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	pravegav1alpha1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestPravegaTemplate(t *testing.T) {
	// Setup user request
	var (
		version             = "0.4.0"
		zookeeperUri        = "foo-client:2181"
		externalAccess_type = "LoadBalancer"

		bookkeeper_image_repository = "tristan1900/bookkeeper"
		bookkeeper_replicas         = "3"
		bookkeeper_autoRecovery     = "false"

		pravega_image_repository     = "tristan1900/pravega"
		pravega_controllerReplicas   = "1"
		pravega_segmentStoreReplicas = "1"
		pravega_debugLogging         = "true"
		pravega_tier2                = "foo"
	)

	// Path to the helm chart
	helmChartPath := "../pravega"

	// Setup the args.
	options := &helm.Options{
		SetValues: map[string]string{
			"version":                      version,
			"zookeeperUri":                 zookeeperUri,
			"externalAccess.type":          externalAccess_type,
			"bookkeeper.image.repository":  bookkeeper_image_repository,
			"bookkeeper.autoRecovery":      bookkeeper_autoRecovery,
			"pravega.image.repository":     pravega_image_repository,
			"pravega.controllerReplicas":   pravega_controllerReplicas,
			"pravega.segmentStoreReplicas": pravega_segmentStoreReplicas,
			"pravega.debugLogging":         pravega_debugLogging,
			"pravega.tier2":                pravega_tier2,
		},
	}

	// Run "helm template" underlying and return the result as output
	output := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/pravega.yaml"})

	// Parse output
	var p pravegav1alpha1.PravegaCluster
	helm.UnmarshalK8SYaml(t, output, &p)

	// Verify the output
	if p.Spec.Version != version {
		t.Fatalf("Rendered pravega version (%s) is not expected (%s)", p.Spec.Version, version)
	}

	if p.Spec.ZookeeperUri != zookeeperUri {
		t.Fatalf("Rendered pravega zookeeperUri (%s) is not expected (%s)", p.Spec.ZookeeperUri, zookeeperUri)
	}

	if p.Spec.ExternalAccess.Type != corev1.ServiceTypeLoadBalancer {
		t.Fatalf("Rendered pravega external access type (%s) is not expected (%s)", p.Spec.ExternalAccess.Type, externalAccess_type)
	}

	if p.Spec.Bookkeeper.Image.Repository != bookkeeper_image_repository {
		t.Fatalf("Rendered bookkeeper repository (%s) is not expected (%s)", p.Spec.Bookkeeper.Image.Repository, bookkeeper_image_repository)
	}

	replicas, _ := strconv.Atoi(bookkeeper_replicas)
	if p.Spec.Bookkeeper.Replicas != int32(replicas) {
		t.Fatalf("Rendered bookkeeper replicas (%d) is not expected (%s)", p.Spec.Bookkeeper.Replicas, bookkeeper_replicas)
	}

	enable, _ := strconv.ParseBool(bookkeeper_autoRecovery)
	if *p.Spec.Bookkeeper.AutoRecovery != enable {
		t.Fatalf("Rendered bookkeeper autorecovery (%t) is not expected (%s)", *p.Spec.Bookkeeper.AutoRecovery, bookkeeper_autoRecovery)
	}

	if p.Spec.Pravega.Image.Repository != pravega_image_repository {
		t.Fatalf("Rendered pravega image repo (%s) is not expected (%s)", p.Spec.Pravega.Image.Repository, pravega_image_repository)
	}

	replicas, _ = strconv.Atoi(pravega_controllerReplicas)
	if p.Spec.Pravega.ControllerReplicas != int32(replicas) {
		t.Fatalf("Rendered pravega controller replicas (%d) is not expected (%s)", p.Spec.Pravega.ControllerReplicas, pravega_controllerReplicas)
	}

	enable, _ = strconv.ParseBool(pravega_debugLogging)
	if p.Spec.Pravega.DebugLogging != enable {
		t.Fatalf("Rendered pravega controller replicas (%t) is not expected (%s)", p.Spec.Pravega.DebugLogging, pravega_debugLogging)
	}

	if p.Spec.Pravega.Tier2.FileSystem.PersistentVolumeClaim.ClaimName != pravega_tier2 {
		t.Fatalf("Rendered pravega tier2 (%s) is not expected (%s)", p.Spec.Pravega.Tier2.FileSystem.PersistentVolumeClaim.ClaimName, pravega_tier2)
	}
}
