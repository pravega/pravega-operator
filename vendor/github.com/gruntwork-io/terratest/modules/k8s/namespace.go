package k8s

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateNamespace will create a new Kubernetes namespace on the cluster targeted by the provided options. This will
// fail the test if there is an error in creating the namespace.
func CreateNamespace(t *testing.T, options *KubectlOptions, namespaceName string) {
	require.NoError(t, CreateNamespaceE(t, options, namespaceName))
}

// CreateNamespaceE will create a new Kubernetes namespace on the cluster targeted by the provided options.
func CreateNamespaceE(t *testing.T, options *KubectlOptions, namespaceName string) error {
	clientset, err := GetKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return err
	}

	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	_, err = clientset.CoreV1().Namespaces().Create(&namespace)
	return err
}

// GetNamespace will query the Kubernetes cluster targeted by the provided options for the requested namespace. This will
// fail the test if there is an error in creating the namespace.
func GetNamespace(t *testing.T, options *KubectlOptions, namespaceName string) corev1.Namespace {
	namespace, err := GetNamespaceE(t, options, namespaceName)
	require.NoError(t, err)
	return namespace
}

// GetNamespaceE will query the Kubernetes cluster targeted by the provided options for the requested namespace.
func GetNamespaceE(t *testing.T, options *KubectlOptions, namespaceName string) (corev1.Namespace, error) {
	clientset, err := GetKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return corev1.Namespace{}, err
	}

	namespace, err := clientset.CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})
	return *namespace, err
}

// DeleteNamespace will delete the requested namespace from the Kubernetes cluster targeted by the provided options. This will
// fail the test if there is an error in creating the namespace.
func DeleteNamespace(t *testing.T, options *KubectlOptions, namespaceName string) {
	require.NoError(t, DeleteNamespaceE(t, options, namespaceName))
}

// DeleteNamespaceE will delete the requested namespace from the Kubernetes cluster targeted by the provided options.
func DeleteNamespaceE(t *testing.T, options *KubectlOptions, namespaceName string) error {
	clientset, err := GetKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return err
	}

	return clientset.CoreV1().Namespaces().Delete(namespaceName, &metav1.DeleteOptions{})
}
