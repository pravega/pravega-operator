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
