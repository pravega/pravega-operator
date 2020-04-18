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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestPravegaOperatorTemplate(t *testing.T) {
	g := NewGomegaWithT(t)

	// Path to the helm chart
	helmChartPath := "../pravega-operator"

	// Setup the args.
	options := &helm.Options{
		SetValues: map[string]string{
			"image.repository": "tristan1900/pravega-operator",
			"image.tag":        "0.3.0",
			"image.pullPolicy": "IfNotPresent",
		},
	}

	// Run "helm template" underlying and return the result as output
	output := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/operator.yaml"})

	// Parse output
	var deploy appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deploy)

	// Verify the output
	podContainers := deploy.Spec.Template.Spec.Containers

	g.Expect(podContainers[0].Image).To(Equal("tristan1900/pravega-operator:0.3.0"))
	g.Expect(podContainers[0].ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
}
