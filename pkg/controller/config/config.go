/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package config

// TestMode enables test mode in the operator and applies
// the following changes:
// - Disables Pravega Controller minimum number of replicas
// - Disables Segment Store minimum number of replicas
// - Enables privileged mode for Segment Store / Controller containers
var TestMode bool

// DisableFinalizer disables the finalizers for Pravega clusters and
// skips the znode cleanup phase when Pravega cluster get deleted.
// This is useful when operator deletion may happen before Pravega clusters deletion.
// NOTE: enabling this flag with caution! It causes stale znode data in zookeeper and
// leads to conflicts with subsequent Pravega clusters deployments.
var DisableFinalizer bool
