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
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/ready"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/pravega/bookkeeper-operator/pkg/apis"
	"github.com/pravega/bookkeeper-operator/pkg/controller"
	controllerconfig "github.com/pravega/bookkeeper-operator/pkg/controller/config"
	"github.com/pravega/bookkeeper-operator/pkg/version"
	"github.com/pravega/bookkeeper-operator/pkg/webhook"
	log "github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
	flag.BoolVar(&webhookFlag, "webhook", true, "Enable webhook, the default is enabled.")
}

func printVersion() {
	log.Printf("bookkeeper-operator Version: %v", version.Version)
	log.Printf("Git SHA: %s", version.GitSHA)
	log.Printf("Go Version: %s", runtime.Version())
	log.Printf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Printf("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	flag.Parse()

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

	// Become the leader before proceeding
	leader.Become(context.TODO(), "bookkeeper-operator-lock")

	r := ready.NewFileReady()
	err = r.Set()
	if err != nil {
		log.Fatal(err, "")
	}
	defer r.Unset()

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
	log.Print("Scheme Setup completed")

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Fatal(err)
	}
	log.Print("Controller Setup completed")

	if webhookFlag {
		// Setup webhook
		if err := webhook.AddToManager(mgr); err != nil {
			log.Fatal(err)
		}
	}

	log.Print("Webhook Setup completed")
	log.Print("Starting the Cmd")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Fatal(err, "manager exited non-zero")
	}
}
