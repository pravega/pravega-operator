/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package pravegacluster

import (
	pravegav1alpha1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	log "github.com/sirupsen/logrus"
)

func (r *ReconcilePravegaCluster) syncClusterVersion(p *pravegav1alpha1.PravegaCluster) (err error) {

	if p.Spec.Version != p.Status.CurrentVersion {
		// Set upgrading target version to p.Spec.Version in resource status
		// Set upgrading condition to True
		log.Printf("User wants to upgrade from %s to %s", p.Status.CurrentVersion, p.Spec.Version)
	}

	return nil
}
