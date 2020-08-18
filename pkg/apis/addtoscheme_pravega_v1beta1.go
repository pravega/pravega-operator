/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package apis

import (
	bkapi "github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	zkapi "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1beta1.SchemeBuilder.AddToScheme)

	AddToSchemes = append(AddToSchemes, bkapi.SchemeBuilder.AddToScheme)

	AddToSchemes = append(AddToSchemes, zkapi.SchemeBuilder.AddToScheme)
}
