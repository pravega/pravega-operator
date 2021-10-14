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
	"flag"
	"os"
	"runtime"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/pravega/pravega-operator/pkg/apis"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
	"github.com/pravega/pravega-operator/pkg/controller"
	controllerconfig "github.com/pravega/pravega-operator/pkg/controller/config"
	"github.com/pravega/pravega-operator/pkg/util"
	"github.com/pravega/pravega-operator/pkg/version"
	log "github.com/sirupsen/logrus"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	logz "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var (
	versionFlag bool
	webhookFlag bool
)

func init() {
	flag.BoolVar(&versionFlag, "version", false, "Show version and quit")
	flag.BoolVar(&controllerconfig.TestMode, "test", false, "Enable test mode. Do not use this flag in production")
	flag.BoolVar(&controllerconfig.DisableFinalizer, "disableFinalizer", false, "Disable finalizers for pravegaclusters. Use this flag with awareness of the consequences")
	flag.BoolVar(&webhookFlag, "webhook", true, "Enable webhook, the default is enabled.")
}

func printVersion() {
	log.Printf("pravega-operator Version: %v", version.Version)
	log.Printf("Git SHA: %s", version.GitSHA)
	log.Printf("Go Version: %s", runtime.Version())
	log.Printf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Printf("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	flag.Parse()

	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface.
	logf.SetLogger(logz.New(logz.UseDevMode(false)))

	printVersion()

	if versionFlag {
		os.Exit(0)
	}

	if controllerconfig.TestMode {
		log.Warn("----- Running in test mode. Make sure you are NOT in production -----")
	}

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Fatal(err, "failed to get watch namespace")
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		log.Error(err, "failed to get operator namespace")
		os.Exit(1)
	}

	// Become the leader before proceeding
	err = util.BecomeLeader(context.TODO(), cfg, "pravega-operator-lock", operatorNs)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace})

	if err != nil {
		log.Fatal(err)
	}

	log.Print("Registering Components")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal(err)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Fatal(err)
	}

	v1beta1.Mgr = mgr
	if webhookFlag {
		if err := (&v1beta1.PravegaCluster{}).SetupWebhookWithManager(mgr); err != nil {
			log.Error(err, "unable to create webhook %s", err.Error())
			os.Exit(1)
		}
	}

	log.Print("Starting the Cmd")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Fatal(err, "manager exited non-zero")
	}
}
