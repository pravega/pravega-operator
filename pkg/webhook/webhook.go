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
	pravegav1alpha1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/util"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
	admissiontypes "sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"

	log "github.com/sirupsen/logrus"
)

const (
	CertDir = "/var/run/secrets/kubernetes.io/"
)

var (
	compatibilityMatrix = []string{
		"0.1.0",
		"0.2.0",
		"0.2.1",
		"0.3.0",
		"0.3.1",
		"0.3.2",
		"0.4.0",
	}
	upgradePath = map[string][]string{
		"0.1.0": []string{"0.2.0"},
		"0.2.0": []string{"0.3.0"},
		"0.3.0": []string{"0.3.1"},
		"0.3.1": []string{"0.3.2"},
		"0.3.2": []string{"0.4.0"},
		"0.4.0": []string{"0.5.0"},
	}
)

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
		Validating().
		Operations(admissionregistrationv1beta1.Create).
		ForType(&pravegav1alpha1.PravegaCluster{}).
		Handlers(&pravegaWebhookHandler{}).
		WithManager(mgr).
		Build()
}

func newWebhookServer(mgr manager.Manager) (*webhook.Server, error) {
	return webhook.NewServer("admission-webhook-server", mgr, webhook.ServerOptions{
		CertDir: CertDir,
		BootstrapOptions: &webhook.BootstrapOptions{
			Secret: &types.NamespacedName{
				Namespace: os.Getenv("WATCH_NAMESPACE"),
				Name:      os.Getenv("SECRET_NAME"),
			},

			// TODO: garbage collect webhook service
			Service: &webhook.Service{
				Namespace: os.Getenv("WATCH_NAMESPACE"),
				Name:      "admission-webhook-server-service",
				Selectors: map[string]string{
					"component": "operator",
				},
			},
		},
	})
}

type pravegaWebhookHandler struct {
	client  client.Client
	scheme  *runtime.Scheme
	decoder admissiontypes.Decoder
}

var _ admission.Handler = &pravegaWebhookHandler{}

// Webhook server will call this func when request comes in
func (pwh *pravegaWebhookHandler) Handle(ctx context.Context, req admissiontypes.Request) admissiontypes.Response {
	log.Printf("Webhook is handling incoming requests")
	pravega := &pravegav1alpha1.PravegaCluster{}

	if err := pwh.decoder.Decode(req, pravega); err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	copy := pravega.DeepCopy()

	err := pwh.validatePravegaManifest(ctx, copy)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}

	return admission.ValidationResponse(true, "")
}

func (pwh *pravegaWebhookHandler) validatePravegaManifest(ctx context.Context, p *pravegav1alpha1.PravegaCluster) error {
	if error := pwh.validatePravegaVersion(ctx, p); error != nil {
		return fmt.Errorf("Failed to pass version validation: %v", error)
	}

	//TODO: Other validations
	return nil
}

func (pwh *pravegaWebhookHandler) validatePravegaVersion(ctx context.Context, p *pravegav1alpha1.PravegaCluster) error {
	// Check if the request has a legal Pravega version
	if p.Spec.Version == "" {
		if p.Spec.Pravega != nil && p.Spec.Pravega.Image != nil && p.Spec.Pravega.Image.Tag != "" {
			p.Spec.Version = p.Spec.Pravega.Image.Tag
		} else {
			p.Spec.Version = pravegav1alpha1.DefaultPravegaVersion
		}
	}
	if !util.ContainsVersion(compatibilityMatrix, p.Spec.Version) {
		return fmt.Errorf("Illegal Pravega cluster version: %s", p.Spec.Version)
	}

	// Check if the request requires an upgrade
	found := &pravegav1alpha1.PravegaCluster{}
	nn := types.NamespacedName{
		Namespace: p.Namespace,
		Name:      p.Name,
	}
	err := pwh.client.Get(context.TODO(), nn, found)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failed to get current Pravega cluster: %v", err)
	}

	// This is not an upgrade if "found" is empty or the requested version is equal to the running version
	fmt.Printf("the error is %v", err)
	if errors.IsNotFound(err) || found.Spec.Version == p.Spec.Version {
		return nil
	}

	// This is an upgrade, check if this requested version is in the upgrade path
	path, ok := upgradePath[found.Spec.Version]
	if !ok {
		// This should never happen
		return fmt.Errorf("Failed to find current cluster version in the upgrade path")
	}
	if !util.ContainsVersion(path, p.Spec.Version) {
		return fmt.Errorf("Upgrading from version %s to version %s is not supported", found.Spec.Version, p.Spec.Version)
	}

	return nil
}

// pravegaWebhookHandler implements inject.Client.
var _ inject.Client = &pravegaWebhookHandler{}

// InjectClient injects the client into the pravegaWebhookHandler
func (pwh *pravegaWebhookHandler) InjectClient(c client.Client) error {
	pwh.client = c
	return nil
}

// pravegaWebhookHandler implements inject.Decoder.
var _ inject.Decoder = &pravegaWebhookHandler{}

// InjectDecoder injects the decoder into the pravegaWebhookHandler
func (pwh *pravegaWebhookHandler) InjectDecoder(d admissiontypes.Decoder) error {
	pwh.decoder = d
	return nil
}

// pravegaWebhookHandler implements inject.Scheme.
var _ inject.Scheme = &pravegaWebhookHandler{}

// InjectClient injects the client into the pravegaWebhookHandler
func (pwh *pravegaWebhookHandler) InjectScheme(s *runtime.Scheme) error {
	pwh.scheme = s
	return nil
}
