package k8s

import (
	"errors"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// The following line loads the gcp plugin which is required to authenticate against GKE clusters.
	// See: https://github.com/kubernetes/client-go/issues/242
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// GetNodes queries Kubernetes for information about the worker nodes registered to the cluster. If anything goes wrong,
// the function will automatically fail the test.
func GetNodes(t *testing.T) []corev1.Node {
	nodes, err := GetNodesE(t)
	if err != nil {
		t.Fatal(err)
	}
	return nodes
}

// GetNodesE queries Kubernetes for information about the worker nodes registered to the cluster.
func GetNodesE(t *testing.T) ([]corev1.Node, error) {
	logger.Logf(t, "Getting list of nodes from Kubernetes")
	clientset, err := GetKubernetesClientE(t)
	if err != nil {
		return nil, err
	}

	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodes.Items, err
}

// GetReadyNodes queries Kubernetes for information about the worker nodes registered to the cluster and only returns
// those that are in the ready state. If anything goes wrong, the function will automatically fail the test.
func GetReadyNodes(t *testing.T) []corev1.Node {
	nodes, err := GetReadyNodesE(t)
	if err != nil {
		t.Fatal(err)
	}
	return nodes
}

// GetReadyNodesE queries Kubernetes for information about the worker nodes registered to the cluster and only returns
// those that are in the ready state.
func GetReadyNodesE(t *testing.T) ([]corev1.Node, error) {
	nodes, err := GetNodesE(t)
	if err != nil {
		return nil, err
	}
	logger.Logf(t, "Filtering list of nodes from Kubernetes for Ready nodes")
	nodesFiltered := []corev1.Node{}
	for _, node := range nodes {
		if IsNodeReady(node) {
			nodesFiltered = append(nodesFiltered, node)
		}
	}
	return nodesFiltered, nil
}

// IsNodeReady takes a Kubernetes Node information object and checks if the Node is in the ready state.
func IsNodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// WaitUntilAllNodesReady continuously polls the Kubernetes cluster until all nodes in the cluster reach the ready
// state, or runs out of retries. Will fail the test immediately if it times out.
func WaitUntilAllNodesReady(t *testing.T, retries int, sleepBetweenRetries time.Duration) {
	err := WaitUntilAllNodesReadyE(t, retries, sleepBetweenRetries)
	if err != nil {
		t.Fatal(err)
	}
}

// WaitUntilAllNodesReadyE continuously polls the Kubernetes cluster until all nodes in the cluster reach the ready
// state, or runs out of retries.
func WaitUntilAllNodesReadyE(t *testing.T, retries int, sleepBetweenRetries time.Duration) error {
	message, err := retry.DoWithRetryE(
		t,
		"Wait for all Kube Nodes to be ready",
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			_, err := AreAllNodesReadyE(t)
			if err != nil {
				return "", err
			}
			return "All nodes ready", nil
		},
	)
	logger.Logf(t, message)
	return err
}

// AreAllNodesReady checks if all nodes are ready in the Kubernetes cluster targeted by the current config context
func AreAllNodesReady(t *testing.T) bool {
	nodesReady, _ := AreAllNodesReadyE(t)
	return nodesReady
}

// AreAllNodesReadyE checks if all nodes are ready in the Kubernetes cluster targeted by the current config context. If
// false, returns an error indicating the reason.
func AreAllNodesReadyE(t *testing.T) (bool, error) {
	nodes, err := GetNodesE(t)
	if err != nil {
		return false, err
	}
	if len(nodes) == 0 {
		return false, errors.New("No nodes available")
	}
	for _, node := range nodes {
		if !IsNodeReady(node) {
			return false, errors.New("Not all nodes ready")
		}
	}
	return true, nil
}
