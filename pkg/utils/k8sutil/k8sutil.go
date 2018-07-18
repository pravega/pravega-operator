package k8sutil

import (
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AsOwnerRef(pravegaCluster *api.PravegaCluster) *metav1.OwnerReference {
	falseVar := false
	return &metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       api.PravegaClusterKind,
		Name:       pravegaCluster.Name,
		UID:        pravegaCluster.UID,
		Controller: &falseVar,
	}
}

func DeleteCollection(apiVersion string, kind string, namespace string, labels string) (err error) {
	resourceClient, _, err := k8sclient.GetResourceClient(apiVersion, kind, namespace)
	if err != nil {
		return fmt.Errorf("failed to get resource client: %v", err)
	}

	return resourceClient.DeleteCollection(nil, metav1.ListOptions{
		LabelSelector: labels,
	})
}
