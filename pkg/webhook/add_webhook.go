/**
 * Copyright (c) 2019 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package webhook

import (
	"log"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	pravegav1alpha1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
)

const (
	CertDir           = "/tmp"
	WebhookConfigName = "pravega-webhook-config"
	WebhookName       = "pravegawebhook.pravega.io"
	WebhookSvcName    = "pravega-webhook-svc"
)

// AddToManagerFuncs is a list of functions to add all Webhooks to the Manager
var AddToManagerFuncs []func(manager.Manager) error

func init() {
	// AddToManagerFuncs is a list of functions to create Webhooks and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, Add)
}

// AddToManager adds all Webhooks to the Manager
func AddToManager(m manager.Manager) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}

// Create webhook server and register webhook to it
func Add(mgr manager.Manager) error {
	log.Printf("Initializing webhook")

	svr, err := newWebhookServer(mgr)
	if err != nil {
		log.Printf("Failed to create webhook server: %v", err)
		return err
	}

	wh, err := newValidatingWebhook(mgr)
	if err != nil {
		log.Printf("Failed to create validating webhook: %v", err)
		return err
	}

	svr.Register(wh)
	return nil
}

func newValidatingWebhook(mgr manager.Manager) (*admission.Webhook, error) {
	return builder.NewWebhookBuilder().
		Name(WebhookName).
		Mutating().
		Operations(admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update).
		ForType(&pravegav1alpha1.PravegaCluster{}).
		Handlers(&pravegaWebhookHandler{}).
		WithManager(mgr).
		Build()
}

func newWebhookServer(mgr manager.Manager) (*webhook.Server, error) {
	namespace, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return nil, err
	}
	return webhook.NewServer(WebhookSvcName, mgr, webhook.ServerOptions{
		CertDir: CertDir,
		BootstrapOptions: &webhook.BootstrapOptions{
			MutatingWebhookConfigName: WebhookConfigName,
			// TODO: garbage collect webhook k8s service
			Service: &webhook.Service{
				Namespace: namespace,
				Name:      WebhookSvcName,
				Selectors: map[string]string{
					"component": "pravega-operator",
				},
			},
		},
	})
}
