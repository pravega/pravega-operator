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
	"context"
	"fmt"
	"log"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	pravegav1alpha1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	wh, err := newMutatingWebhook(mgr)
	if err != nil {
		log.Printf("Failed to create validating webhook: %v", err)
		return err
	}

	svr.Register(wh)

	err = addOwnerReferenceToWebhookK8sService(mgr)
	if err != nil {
		log.Printf("Failed to update webhook svc: %v", err)
	}
	return nil
}

func newMutatingWebhook(mgr manager.Manager) (*admission.Webhook, error) {
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

func addOwnerReferenceToWebhookK8sService(mgr manager.Manager) error {
	// Use non-default kube client to talk to apiserver directly since the default kube client uses cache and
	// that cache is not updated quickly enough for this method to use to get the operator deployment.
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	c, _ := client.New(cfg, client.Options{Scheme: mgr.GetScheme()})

	// get webhook k8s service object
	svc := &corev1.Service{}
	ns, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return err
	}
	nn := types.NamespacedName{Namespace: ns, Name: WebhookSvcName}

	err = c.Get(context.TODO(), nn, svc)
	if err != nil {
		return fmt.Errorf("failed to get webhook service: %v", err)
	}

	// get operator deployment
	name, err := k8sutil.GetOperatorName()
	if err != nil {
		return err
	}
	nn = types.NamespacedName{Namespace: ns, Name: name}
	deployment := &appsv1.Deployment{}

	err = c.Get(context.TODO(), nn, deployment)
	if err != nil {
		return fmt.Errorf("failed to get operator deployment: %v", err)
	}

	// add owner reference so the k8s service could be garbage collected
	controllerutil.SetControllerReference(deployment, svc, mgr.GetScheme())

	// update webhook k8s service
	err = c.Update(context.TODO(), svc)
	if err != nil {
		return fmt.Errorf("failed to update webhook k8s service: %v", err)
	}
	return nil
}
