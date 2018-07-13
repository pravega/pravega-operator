package stub

import (
	"context"
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.PravegaCluster:
		if event.Deleted {
			destroyPravegaCluster(o)
		} else {
			createPravegaCluster(o)
		}
	}
	return nil
}

func createPravegaCluster(pravegaCluster *v1alpha1.PravegaCluster) {
	var ownerRef = makeOwnerRef(pravegaCluster)

	createBookie(ownerRef, pravegaCluster)
	createController(ownerRef, pravegaCluster)
	createSegmentStore(ownerRef, pravegaCluster)
}

func destroyPravegaCluster(pravegaCluster *v1alpha1.PravegaCluster) {
	var ownerRef = makeOwnerRef(pravegaCluster)

	destroyBookie(ownerRef, pravegaCluster)
	destroyController(ownerRef, pravegaCluster)
	destroySegmentStore(ownerRef, pravegaCluster)
}

func makeOwnerRef(pravegaCluster *v1alpha1.PravegaCluster) *metav1.OwnerReference {
	return metav1.NewControllerRef(pravegaCluster, schema.GroupVersionKind{
		Group:   v1alpha1.SchemeGroupVersion.Group,
		Version: v1alpha1.SchemeGroupVersion.Version,
		Kind:    "PravegaCluster",
	})
}

func generateKindName(kind string, name string) string {
	return fmt.Sprintf("%s-%s", name, kind)
}

func cascadeDelete(object sdk.Object) error {
	return sdk.Delete(object, cascadeDeleteOption())
}

func cascadeDeleteOption() sdk.DeleteOption {
	propagationPolicy := metav1.DeletePropagationBackground

	return sdk.WithDeleteOptions(&metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
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
