/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package pravegacluster

import (
	"context"
	"encoding/json"
	pravegav1alpha1 "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

type podAnnotator struct {
	client  client.Client
	decoder types.Decoder
}

// podAnnotator Iimplements admission.Handler.
var _ admission.Handler = &podAnnotator{}

// podAnnotator adds an annotation to every incoming pods.
func (a *podAnnotator) Handle(ctx context.Context, req types.Request) types.Response {
	pravega := &pravegav1alpha1.PravegaCluster{}

	if err := json.Unmarshal(req.AdmissionRequest.Object.Raw, pravega); err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	copy := pravega.DeepCopy()


	err := a.validatePravegaManifest(ctx, copy)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}
	// admission.PatchResponse generates a Response containing patches.
	return admission.PatchResponse(pravega, copy)
}

// mutatePodsFn add an annotation to the given pod
func (a *podAnnotator) validatePravegaManifest(ctx context.Context, pod *pravegav1alpha1.PravegaCluster) error {
	pod.Annotations["verification"] = "success"
	return nil
}