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
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestPravegaOperatorTemplate(t *testing.T) {
	// Setup user request
	var (
		repo   = "tristan1900/pravega-operator"
		tag    = "0.3.0"
		policy = "IfNotPresent"
	)

	// Path to the helm chart
	helmChartPath := "../pravega-operator"

	// Setup the args.
	options := &helm.Options{
		SetValues: map[string]string{
			"pravegaOperator.image.repository": repo,
			"pravegaOperator.image.tag":        tag,
			"pravegaOperator.image.pullPolicy": policy,
		},
	}

	// Run "helm template" underlying and return the result as output
	output := helm.RenderTemplate(t, options, helmChartPath, []string{"templates/operator.yaml"})

	// Parse output
	var deploy appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deploy)

	// Verify the output
	image := strings.Join([]string{repo, tag}, ":")
	podContainers := deploy.Spec.Template.Spec.Containers

	if podContainers[0].Image != image {
		t.Fatalf("Rendered container image (%s) is not expected (%s)", podContainers[0].Image, image)
	}

	if podContainers[0].ImagePullPolicy != corev1.PullIfNotPresent {
		t.Fatalf("Rendered container image policy (%s) is not expected (%s)", podContainers[0].ImagePullPolicy, policy)
	}
}
