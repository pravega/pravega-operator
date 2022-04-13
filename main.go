/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/operator-framework/operator-lib/leader"
	controllerconfig "github.com/pravega/bookkeeper-operator/pkg/controller/config"
	v1beta1 "github.com/pravega/pravega-operator/api/v1beta1"
	"github.com/pravega/pravega-operator/controllers"
	"github.com/pravega/pravega-operator/pkg/version"
	"github.com/sirupsen/logrus"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

var (
	versionFlag bool
	webhookFlag bool
	log         = ctrl.Log.WithName("cmd")
	scheme      = apimachineryruntime.NewScheme()
)

func init() {
	flag.BoolVar(&versionFlag, "version", false, "Show version and quit")
	flag.BoolVar(&controllerconfig.TestMode, "test", false, "Enable test mode. Do not use this flag in production")
	flag.BoolVar(&webhookFlag, "webhook", true, "Enable webhook, the default is enabled.")
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme

}

func printVersion() {
	log.Info(fmt.Sprintf("pravega-operator Version: %v", version.Version))
	log.Info(fmt.Sprintf("Git SHA: %s", version.GitSHA))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")

	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))
	printVersion()

	if versionFlag {
		os.Exit(0)
	}

	if controllerconfig.TestMode {
		logrus.Warn("----- Running in test mode. Make sure you are NOT in production -----")
	}

	if controllerconfig.DisableFinalizer {
		logrus.Warn("----- Running with finalizer disabled. -----")
	}

	ctx := context.TODO()
	// Become the leader before proceeding
	err := leader.Become(ctx, "pravega-operator-lock")
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	log.Info("Registering Components")

	if err = (&controllers.PravegaClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "PravegaCluster")
		os.Exit(1)
	}

	v1beta1.Mgr = mgr

	if webhookFlag {
		if err = (&v1beta1.PravegaCluster{}).SetupWebhookWithManager(mgr); err != nil {
			log.Error(err, "unable to create webhook", "webhook", "PravegaCluster")
			os.Exit(1)
		}
	}
	//+kubebuilder:scaffold:builder

	log.Info("Webhook Setup completed")
	log.Info("Starting the Cmd")

	// Start the Cmd

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
