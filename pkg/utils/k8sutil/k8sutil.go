/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package k8sutil

import (
	"fmt"

	"os"

	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AsOwnerRef(pravegaCluster *api.PravegaCluster) *metav1.OwnerReference {
	falseVar := false
	return &metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       api.PravegaClusterKind,
		Name:       pravegaCluster.Name,
		UID:        pravegaCluster.UID,
		Controller: &falseVar,
	}
}

func DeleteCollection(apiVersion string, kind string, namespace string, labels string) (err error) {
	resourceClient, _, err := k8sclient.GetResourceClient(apiVersion, kind, namespace)
	if err != nil {
		return fmt.Errorf("failed to get resource client: %v", err)
	}

	return resourceClient.DeleteCollection(nil, metav1.ListOptions{
		LabelSelector: labels,
	})
}

// GetWatchNamespaceAllowBlank returns the namespace the operator should be watching for changes
func GetWatchNamespaceAllowBlank() (string, error) {
	ns, found := os.LookupEnv(k8sutil.WatchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", k8sutil.WatchNamespaceEnvVar)
	}
	return ns, nil
}
