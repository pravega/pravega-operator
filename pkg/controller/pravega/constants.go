/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package pravega

const (
	cacheVolumeName          = "cache"
	cacheVolumeMountPoint    = "/tmp/pravega/cache"
	ltsFileMountPoint        = "/mnt/tier2"
	ltsVolumeName            = "tier2"
	segmentStoreKind         = "pravega-segmentstore"
	ssSecretVolumeName       = "ss-secret"
	tlsVolumeName            = "tls-secret"
	tlsMountDir              = "/etc/secret-volume"
	caBundleVolumeName       = "ca-bundle"
	caBundleMountDir         = "/etc/secret-volume/ca-bundle"
	authVolumeName           = "auth-passwd-secret"
	authMountDir             = "/etc/auth-passwd-volume"
	defaultTokenSigningKey   = "secret"
	controllerAuthVolumeName = "controller-auth-secret"
	ssAuthVolumeName         = "ss-auth-secret"
	controllerAuthMountDir   = "/etc/controller-auth-volume"
	ssAuthMountDir           = "/etc/ss-auth-volume"
	influxDBSecretVolumeName = "influxdb-secret"
)
