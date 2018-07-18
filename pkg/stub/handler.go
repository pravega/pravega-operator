package stub

import (
	"context"
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	if event.Deleted {
		// K8s will garbage collect and resources until zookeeper cluster delete
		return nil
	}

	switch o := event.Object.(type) {
	case *api.PravegaCluster:
		createPravegaCluster(o)
	}
	return nil
}

func createPravegaCluster(pravegaCluster *api.PravegaCluster) {
	var ownerRef = asOwnerRef(pravegaCluster)

	createBookie(ownerRef, pravegaCluster)
	createController(ownerRef, pravegaCluster)
	createSegmentStore(ownerRef, pravegaCluster)
}

func asOwnerRef(pravegaCluster *api.PravegaCluster) *metav1.OwnerReference {
	falseVar := false
	return &metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       api.PravegaClusterKind,
		Name:       pravegaCluster.Name,
		UID:        pravegaCluster.UID,
		Controller: &falseVar,
	}
}

func generateKindName(kind string, name string) string {
	return fmt.Sprintf("%s-%s", name, kind)
}

func deleteCollection(apiVersion string, kind string, namespace string, labels string) (err error) {
	resourceClient, _, err := k8sclient.GetResourceClient(apiVersion, kind, namespace)
	if err != nil {
		return fmt.Errorf("failed to get resource client: %v", err)
	}

	return resourceClient.DeleteCollection(nil, metav1.ListOptions{
		LabelSelector: labels,
	})
}
