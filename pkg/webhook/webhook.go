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
	"net/http"

	pravegav1alpha1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	admissiontypes "sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"

	log "github.com/sirupsen/logrus"
)

var (
	// The key is the supported versions, the value is a list of versions that can upgrade to.
	supportedVersions = map[string][]string{
		"0.1.0": []string{"0.1.0"},
		"0.2.0": []string{"0.2.0"},
		"0.3.0": []string{"0.3.0", "0.3.1", "0.3.2"},
		"0.3.1": []string{"0.3.1", "0.3.2"},
		"0.3.2": []string{"0.3.2"},
		"0.4.0": []string{"0.4.0"},
		"0.5.0": []string{"0.5.0", "0.5.1", "0.6.0", "0.6.1", "0.7.0"},
		"0.5.1": []string{"0.5.1", "0.6.0", "0.6.1", "0.7.0"},
		"0.6.0": []string{"0.6.0", "0.6.1", "0.7.0"},
		"0.6.1": []string{"0.6.1", "0.7.0"},
		"0.7.0": []string{"0.7.0"},
	}
)

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

	if err := pwh.clusterIsAvailable(ctx, copy); err != nil {
		return admission.ErrorResponse(http.StatusServiceUnavailable, err)
	}

	if err := pwh.mutatePravegaManifest(ctx, copy); err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	return admission.PatchResponse(pravega, copy)
}

func (pwh *pravegaWebhookHandler) mutatePravegaManifest(ctx context.Context, p *pravegav1alpha1.PravegaCluster) error {
	if err := pwh.mutatePravegaVersion(ctx, p); err != nil {
		return err
	}

	//Add other validators here
	return nil
}

func (pwh *pravegaWebhookHandler) mutatePravegaVersion(ctx context.Context, p *pravegav1alpha1.PravegaCluster) error {
	// Identify the request Pravega version
	// Mutate the version if it is empty
	if p.Spec.Version == "" {
		if p.Spec.Pravega != nil && p.Spec.Pravega.Image != nil && p.Spec.Pravega.Image.Tag != "" {
			p.Spec.Version = p.Spec.Pravega.Image.Tag
		} else {
			p.Spec.Version = pravegav1alpha1.DefaultPravegaVersion
		}
	}
	// Set Pravega and Bookkeeper tag to empty
	if p.Spec.Pravega != nil && p.Spec.Pravega.Image != nil && p.Spec.Pravega.Image.Tag != "" {
		p.Spec.Pravega.Image.Tag = ""
	}
	if p.Spec.Bookkeeper != nil && p.Spec.Bookkeeper.Image != nil && p.Spec.Bookkeeper.Image.Tag != "" {
		p.Spec.Bookkeeper.Image.Tag = ""
	}

	requestVersion := p.Spec.Version

	if p.Status.IsClusterInUpgradeFailedState() {
		if requestVersion != p.Status.GetLastVersion() {
			return fmt.Errorf("Rollback to version %s not supported. Only rollback to version %s is supported.", requestVersion, p.Status.GetLastVersion())
		}
		return nil
	}

	// Allow upgrade only if Cluster is in Ready State
	// Check if the request has a valid Pravega version
	normRequestVersion, err := util.NormalizeVersion(requestVersion)
	if err != nil {
		return fmt.Errorf("request version is not in valid format: %v", err)
	}
	if _, ok := supportedVersions[normRequestVersion]; !ok {
		return fmt.Errorf("unsupported Pravega cluster version %s", requestVersion)
	}

	// Check if the request is an upgrade
	found := &pravegav1alpha1.PravegaCluster{}
	nn := types.NamespacedName{
		Namespace: p.Namespace,
		Name:      p.Name,
	}
	err = pwh.client.Get(context.TODO(), nn, found)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to obtain PravegarequestVersionCluster resource: %v", err)
	}

	foundVersion := found.Spec.Version
	// This is not an upgrade if "found" is empty or the requested version is equal to the running version
	if errors.IsNotFound(err) || foundVersion == requestVersion {
		return nil
	}

	// This is an upgrade, check if this requested version is in the upgrade path
	normFoundVersion, err := util.NormalizeVersion(foundVersion)
	if err != nil {
		// It should never happen
		return fmt.Errorf("found version is not in valid format, something bad happens: %v", err)
	}
	upgradeList, ok := supportedVersions[normFoundVersion]
	if !ok {
		// It should never happen
		return fmt.Errorf("failed to find current cluster version in the supported versions")
	}
	if !util.ContainsVersion(upgradeList, normRequestVersion) {
		return fmt.Errorf("unsupported upgrade from version %s to %s", foundVersion, requestVersion)
	}
	return nil
}

func (pwh *pravegaWebhookHandler) clusterIsAvailable(ctx context.Context, p *pravegav1alpha1.PravegaCluster) error {
	found := &pravegav1alpha1.PravegaCluster{}
	nn := types.NamespacedName{
		Namespace: p.Namespace,
		Name:      p.Name,
	}
	err := pwh.client.Get(context.TODO(), nn, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to obtain PravegaCluster resource: %v", err)
	}

	if found.Status.IsClusterInUpgradingState() {
		// Reject the request if the requested version is new.
		if p.Spec.Version != found.Spec.Version && p.Spec.Version != found.Status.CurrentVersion {
			return fmt.Errorf("failed to process the request, cluster is upgrading")
		}
	}

	if found.Status.IsClusterInRollbackState() {
		// Reject the request if the requested version is new.
		if p.Spec.Version != found.Spec.Version {
			return fmt.Errorf("failed to process the request, cluster is in rollback")
		}
	}

	if p.Status.IsClusterInErrorState() && !p.Status.IsClusterInUpgradeFailedState() {
		return fmt.Errorf("failed to process the request, cluster is in error state.")
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
