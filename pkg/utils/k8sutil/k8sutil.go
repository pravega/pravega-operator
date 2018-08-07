package k8sutil

import (
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
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

// GetWatchNamespaceAllowBlank returns the namespace the operator should be watching for changes
func GetWatchNamespaceAllowBlank() (string, error) {
	ns, found := os.LookupEnv(k8sutil.WatchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", k8sutil.WatchNamespaceEnvVar)
	}
	return ns, nil
}
