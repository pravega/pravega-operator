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
	"testing"

	. "github.com/onsi/gomega"

	"github.com/gruntwork-io/terratest/modules/helm"
	pravegav1alpha1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestPravegaTemplate(t *testing.T) {
	g := NewGomegaWithT(t)

	// Path to the helm chart
	helmChartPath := "../pravega"

	// Setup the args.
	options := &helm.Options{
		SetValues: map[string]string{
			"version":                      "0.4.0-beta",
			"zookeeperUri":                 "foo-client:2181",
			"externalAccess.enabled":       "true",
			"externalAccess.type":          "NodePort",
			"bookkeeper.image.repository":  "tristan1900/bookkeeper",
			"bookkeeper.replicas":          "5",
			"bookkeeper.autoRecovery":      "false",
			"pravega.image.repository":     "tristan1900/pravega",
			"pravega.controllerReplicas":   "2",
			"pravega.segmentStoreReplicas": "7",
			"pravega.debugLogging":         "true",
			"pravega.tier2":                "foo",
		},
	}

	// Run "helm template" underlying and return the result as output
	output := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/pravega.yaml"})

	// Parse output
	var p pravegav1alpha1.PravegaCluster
	helm.UnmarshalK8SYaml(t, output, &p)

	// Verify the output
	boolFalse := false
	g.Expect(p.Spec.Version).To(Equal("0.4.0-beta"))
	g.Expect(p.Spec.ZookeeperUri).To(Equal("foo-client:2181"))
	g.Expect(p.Spec.ExternalAccess.Enabled).To(BeTrue())
	g.Expect(p.Spec.ExternalAccess.Type).To(Equal(corev1.ServiceTypeNodePort))
	g.Expect(p.Spec.Bookkeeper.Image.Repository).To(Equal("tristan1900/bookkeeper"))
	g.Expect(p.Spec.Bookkeeper.Replicas).To(BeEquivalentTo(5))
	g.Expect(p.Spec.Bookkeeper.AutoRecovery).To(Equal(&boolFalse))
	g.Expect(p.Spec.Pravega.Image.Repository).To(Equal("tristan1900/pravega"))
	g.Expect(p.Spec.Pravega.ControllerReplicas).To(BeEquivalentTo(2))
	g.Expect(p.Spec.Pravega.SegmentStoreReplicas).To(BeEquivalentTo(7))
	g.Expect(p.Spec.Pravega.DebugLogging).To(BeTrue())
	g.Expect(p.Spec.Pravega.Tier2.FileSystem.PersistentVolumeClaim.ClaimName).To(Equal("foo"))
}
