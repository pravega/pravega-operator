/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package main

import (
	"context"
	"runtime"
	"time"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/pravega/pravega-operator/pkg/stub"
	"github.com/pravega/pravega-operator/pkg/utils/k8sutil"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	printVersion()

	resource := "pravega.pravega.io/v1alpha1"
	kind := "PravegaCluster"
	namespace, err := k8sutil.GetWatchNamespaceAllowBlank()
	if err != nil {
		logrus.Fatalf("Failed to get watch namespace: %v", err)
	}

	resyncPeriod := 5
	logrus.Infof("Watching %s, %s, %s, %d", resource, kind, namespace, resyncPeriod)
	sdk.Watch(resource, kind, namespace, time.Duration(resyncPeriod)*time.Second)
	sdk.Handle(stub.NewHandler())
	sdk.Run(context.TODO())
}
