package stub

import (
	"context"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/pravega"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	if event.Deleted {
		// K8s will garbage collect and resources until pravega cluster delete
		return nil
	}

	switch o := event.Object.(type) {
	case *api.PravegaCluster:
		pravega.ReconcilePravegaCluster(o)
	}
	return nil
}
