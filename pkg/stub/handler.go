/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package stub

import (
	"context"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/pravega"
	"github.com/sirupsen/logrus"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) (err error) {
	if event.Deleted {
		// K8s will garbage collect and resources until pravega cluster delete
		return nil
	}

	switch o := event.Object.(type) {
	case *api.PravegaCluster:
		err = pravega.ReconcilePravegaCluster(o)
		if err != nil {
			logrus.Error(err)
			return err
		}
	}

	return nil
}
