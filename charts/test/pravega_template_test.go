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
			"version":                                  "0.4.0-beta",
			"zookeeperUri":                             "foo-client:2181",
			"externalAccessEnabled":                    "true",
			"pravega.controller.replicas":              "2",
			"pravega.controller.externalAccess.type":   "NodePort",
			"pravega.controller.debugLogging":          "true",
			"pravega.segmentstore.externalAccess.type": "NodePort",
			"pravega.segmentstore.replicas":            "7",
			"pravega.segmentstore.debugLogging":        "true",
			"pravega.image.repository":                 "tristan1900/pravega",
			"pravega.tier2":                            "foo",
			"bookkeeper.image.repository":              "tristan1900/bookkeeper",
			"bookkeeper.replicas":                      "5",
			"bookkeeper.autoRecovery":                  "false",
		},
	}

	// Run "helm template" underlying and return the result as output
	output := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/pravega.yaml"})
	//fmt.Println(output)
	// Parse output
	var p pravegav1alpha1.PravegaCluster
	helm.UnmarshalK8SYaml(t, output, &p)

	// Verify the output
	boolFalse := false
	g.Expect(p.Spec.Version).To(Equal("0.4.0-beta"))
	g.Expect(p.Spec.ZookeeperUri).To(Equal("foo-client:2181"))
	g.Expect(p.Spec.ExternalAccessEnabled).To(BeTrue())
	g.Expect(p.Spec.Bookkeeper.Image.Repository).To(Equal("tristan1900/bookkeeper"))
	g.Expect(p.Spec.Bookkeeper.Replicas).To(BeEquivalentTo(5))
	g.Expect(p.Spec.Bookkeeper.AutoRecovery).To(Equal(&boolFalse))
	g.Expect(p.Spec.Pravega.Tier2.FileSystem.PersistentVolumeClaim.ClaimName).To(Equal("foo"))
	g.Expect(p.Spec.Pravega.Controller.ExternalAccess).NotTo(BeNil())
	g.Expect(p.Spec.Pravega.Controller.ExternalAccess.Type).To(Equal(corev1.ServiceTypeNodePort))
	g.Expect(p.Spec.Pravega.SegmentStore.ExternalAccess.Type).To(Equal(corev1.ServiceTypeNodePort))
	g.Expect(p.Spec.Pravega.Image.Repository).To(Equal("tristan1900/pravega"))
	g.Expect(p.Spec.Pravega.Controller.Replicas).To(BeEquivalentTo(2))
	g.Expect(p.Spec.Pravega.SegmentStore.Replicas).To(BeEquivalentTo(7))
	g.Expect(p.Spec.Pravega.Controller.DebugLogging).To(BeTrue())
	g.Expect(p.Spec.Pravega.SegmentStore.DebugLogging).To(BeTrue())

}
