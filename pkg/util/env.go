/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */
package util

import (
	"log"
	"os"
)

func GetEnvOrDie(name string) string {
	value, found := os.LookupEnv(name)
	if found && value != "" {
		return value
	}

	log.Panicf("envvar %s not set or empty", name)
	return ""
}

func EnvPodNamespaceName() string {
	return GetEnvOrDie("POD_NAMESPACE")
}
