package util

import (
	"container/list"
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/samuel/go-zookeeper/zk"
)

const (
	// Set in https://github.com/pravega/pravega/blob/master/docker/bookkeeper/entrypoint.sh#L21
	PravegaPath = "pravega"
	ZkFinalizer = "cleanUpZookeeper"
)

// Wait for pods in cluster to be terminated
func WaitForClusterToTerminate(kubeClient client.Client, p *v1alpha1.PravegaCluster) (err error) {
	listOptions := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(LabelsForPravegaCluster(p)),
	}

	err = wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		podList := &corev1.PodList{}
		err = kubeClient.List(context.TODO(), listOptions, podList)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range podList.Items {
			pod := &podList.Items[i]
			names = append(names, pod.Name)
		}

		if len(names) != 0 {
			return false, nil
		}
		return true, nil
	})

	return err
}

// Delete all znodes related to a specific Pravega cluster
func DeleteAllZnodes(p *v1alpha1.PravegaCluster) (err error) {
	host := []string{p.Spec.ZookeeperUri}
	conn, _, err := zk.Connect(host, time.Second*5)
	if err != nil {
		return fmt.Errorf("failed to connect to zookeeper: %v", err)
	}
	defer conn.Close()

	root := fmt.Sprintf("/%s/%s", PravegaPath, p.Name)
	exist, _, err := conn.Exists(root)
	if err != nil {
		return fmt.Errorf("failed to check if zookeeper path exists: %v", err)
	}

	if exist {
		// Construct BFS tree to delete all znodes recursively
		tree, err := ListSubTreeBFS(conn, root)
		if err != nil {
			return fmt.Errorf("failed to construct BFS tree: %v", err)
		}

		for tree.Len() != 0 {
			err := conn.Delete(tree.Back().Value.(string), -1)
			if err != nil {
				return fmt.Errorf("failed to delete znode (%s): %v", tree.Back().Value.(string), err)
			}
			tree.Remove(tree.Back())
		}
	}
	return nil
}

// Construct a BFS tree
func ListSubTreeBFS(conn *zk.Conn, root string) (*list.List, error) {
	queue := list.New()
	tree := list.New()
	queue.PushBack(root)
	tree.PushBack(root)

	for {
		if queue.Len() == 0 {
			break
		}
		node := queue.Front()
		children, _, err := conn.Children(node.Value.(string))
		if err != nil {
			return tree, err
		}

		for _, child := range children {
			childPath := fmt.Sprintf("%s/%s", node.Value.(string), child)
			queue.PushBack(childPath)
			tree.PushBack(childPath)
		}
		queue.Remove(node)
	}
	return tree, nil
}
