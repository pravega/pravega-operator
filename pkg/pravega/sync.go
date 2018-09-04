package pravega

import (
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/pravega/pravega-operator/pkg/utils/k8sutil"
)

func ReconcilePravegaCluster(pravegaCluster *api.PravegaCluster) (err error) {

	deployBookie(pravegaCluster)
	if err != nil {
		return err
	}

	deployController(pravegaCluster)
	if err != nil {
		return err
	}

	deploySegmentStore(pravegaCluster)
	if err != nil {
		return err
	}

	syncClusterSize(pravegaCluster)
	if err != nil {
		return err
	}

	return nil
}

func syncClusterSize(pravegaCluster *api.PravegaCluster) (err error) {
	syncBookieSize(pravegaCluster)
	syncSegmentStoreSize(pravegaCluster)
	return nil
}

func syncBookieSize(pravegaCluster *api.PravegaCluster) (err error) {
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.StatefulSetNameForBookie(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
		},
	}

	err = sdk.Get(sts)
	if err != nil {
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}

	if *sts.Spec.Replicas != pravegaCluster.Spec.Bookkeeper.Replicas {
		sts.Spec.Replicas = &(pravegaCluster.Spec.Bookkeeper.Replicas )
		err = sdk.Update(sts)
		if err != nil {
			return fmt.Errorf("failed to update size of stateful-set (%s): %v", sts.Name, err)
		}
	}

	return nil
}

func syncSegmentStoreSize(pravegaCluster *api.PravegaCluster) (err error) {
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k8sutil.StatefulSetNameForSegmentstore(pravegaCluster.Name),
			Namespace: pravegaCluster.Namespace,
		},
	}

	err = sdk.Get(sts)
	if err != nil {
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}

	if *sts.Spec.Replicas != pravegaCluster.Spec.Pravega.SegmentStoreReplicas {
		sts.Spec.Replicas = &(pravegaCluster.Spec.Pravega.SegmentStoreReplicas )
		err = sdk.Update(sts)
		if err != nil {
			return fmt.Errorf("failed to update size of stateful-set (%s): %v", sts.Name, err)
		}
	}
	return nil
}
