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
	"strings"

	bookkeeperv1alpha1 "github.com/pravega/bookkeeper-operator/pkg/apis/bookkeeper/v1alpha1"
	"github.com/pravega/bookkeeper-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	admissiontypes "sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"

	log "github.com/sirupsen/logrus"
)

type bookkeeperWebhookHandler struct {
	client  client.Client
	scheme  *runtime.Scheme
	decoder admissiontypes.Decoder
}

var _ admission.Handler = &bookkeeperWebhookHandler{}

// Webhook server will call this func when request comes in
func (pwh *bookkeeperWebhookHandler) Handle(ctx context.Context, req admissiontypes.Request) admissiontypes.Response {
	log.Printf("Webhook is handling incoming requests")
	bookkeeper := &bookkeeperv1alpha1.BookkeeperCluster{}

	if err := pwh.decoder.Decode(req, bookkeeper); err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	copy := bookkeeper.DeepCopy()

	if err := pwh.clusterIsAvailable(ctx, copy); err != nil {
		return admission.ErrorResponse(http.StatusServiceUnavailable, err)
	}

	if err := pwh.mutateBookkeeperManifest(ctx, copy); err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	return admission.PatchResponse(bookkeeper, copy)
}

func (pwh *bookkeeperWebhookHandler) mutateBookkeeperManifest(ctx context.Context, p *bookkeeperv1alpha1.BookkeeperCluster) error {
	if err := pwh.mutateBookkeeperVersion(ctx, p); err != nil {
		return err
	}
	//Add other validators here
	return nil
}

func (pwh *bookkeeperWebhookHandler) mutateBookkeeperVersion(ctx context.Context, bk *bookkeeperv1alpha1.BookkeeperCluster) error {
	configMap := &corev1.ConfigMap{}
	err := pwh.client.Get(ctx, types.NamespacedName{Name: util.ConfigMapNameForBookieVersions(bk.Name), Namespace: bk.Namespace}, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("config map %s not found. Please create this config map first and then retry", util.ConfigMapNameForBookieVersions(bk.Name))
		}
		return err
	}
	// The key is the supported versions, the value is a list of versions that can be upgraded to.
	supportedVersions := configMap.Data

	// Identify the request Bookkeeper version
	// Mutate the version if it is empty
	if bk.Spec.Version == "" {
		if bk.Spec.Image != nil && bk.Spec.Image.Tag != "" {
			bk.Spec.Version = bk.Spec.Image.Tag
		} else {
			bk.Spec.Version = bookkeeperv1alpha1.DefaultBookkeeperVersion
		}
	}
	// Set Bookkeeper and Bookkeeper tag to empty
	if bk.Spec.Image != nil && bk.Spec.Image.Tag != "" {
		bk.Spec.Image.Tag = ""
	}
	if bk.Spec.Image != nil && bk.Spec.Image.Tag != "" {
		bk.Spec.Image.Tag = ""
	}
	requestVersion := bk.Spec.Version

	if bk.Status.IsClusterInUpgradeFailedState() {
		if requestVersion != bk.Status.GetLastVersion() {
			return fmt.Errorf("Rollback to version %s not supported. Only rollback to version %s is supported.", requestVersion, bk.Status.GetLastVersion())
		}
		return nil
	}

	// Allow upgrade only if Cluster is in Ready State
	// Check if the request has a valid Bookkeeper version
	normRequestVersion, err := util.NormalizeVersion(requestVersion)
	if err != nil {
		return fmt.Errorf("request version is not in valid format: %v", err)
	}
	if _, ok := supportedVersions[normRequestVersion]; !ok {
		return fmt.Errorf("unsupported Bookkeeper cluster version %s", requestVersion)
	}

	// Check if the request is an upgrade
	found := &bookkeeperv1alpha1.BookkeeperCluster{}
	nn := types.NamespacedName{
		Namespace: bk.Namespace,
		Name:      bk.Name,
	}
	err = pwh.client.Get(context.TODO(), nn, found)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to obtain BookkeeperrequestVersionCluster resource: %v", err)
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
	upgradeString, ok := supportedVersions[normFoundVersion]
	if !ok {
		// It should never happen
		return fmt.Errorf("failed to find current cluster version in the supported versions")
	}
	upgradeList := strings.Split(upgradeString, ",")
	if !util.ContainsVersion(upgradeList, normRequestVersion) {
		return fmt.Errorf("unsupported upgrade from version %s to %s", foundVersion, requestVersion)
	}
	return nil
}

func (pwh *bookkeeperWebhookHandler) clusterIsAvailable(ctx context.Context, p *bookkeeperv1alpha1.BookkeeperCluster) error {
	found := &bookkeeperv1alpha1.BookkeeperCluster{}
	nn := types.NamespacedName{
		Namespace: p.Namespace,
		Name:      p.Name,
	}
	err := pwh.client.Get(context.TODO(), nn, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to obtain BookkeeperCluster resource: %v", err)
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

// BookkeeperWebhookHandler implements inject.Client.
var _ inject.Client = &bookkeeperWebhookHandler{}

// InjectClient injects the client into the bookkeeperWebhookHandler
func (pwh *bookkeeperWebhookHandler) InjectClient(c client.Client) error {
	pwh.client = c
	return nil
}

// BookkeeperWebhookHandler implements inject.Decoder.
var _ inject.Decoder = &bookkeeperWebhookHandler{}

// InjectDecoder injects the decoder into the bookkeeperWebhookHandler
func (pwh *bookkeeperWebhookHandler) InjectDecoder(d admissiontypes.Decoder) error {
	pwh.decoder = d
	return nil
}

// bookkeeperWebhookHandler implements inject.Scheme.
var _ inject.Scheme = &bookkeeperWebhookHandler{}

// InjectClient injects the client into the bookkeeperWebhookHandler
func (pwh *bookkeeperWebhookHandler) InjectScheme(s *runtime.Scheme) error {
	pwh.scheme = s
	return nil
}
