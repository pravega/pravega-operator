package stub

import (
	"context"
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/sirupsen/logrus"
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
	logrus.Infof("Handling Event")

	switch o := event.Object.(type) {
	case *v1alpha1.PravegaCluster:
		createPravegaCluster(o)
		//err := action.Create(createPravegaCluster(o))
		//if err != nil && !errors.IsAlreadyExists(err) {
		//	logrus.Errorf("Failed to create busybox pod : %v", err)
		//	return err
		//}
	}
	return nil
}

func createPravegaCluster(pravegaCluster *v1alpha1.PravegaCluster) {
	var ownerRef = makeOwnerRef(pravegaCluster)

	logrus.WithFields(logrus.Fields{"cluster": pravegaCluster.Name}).Debug("Creating Bookie Config")
	var err = sdk.Create(makeBookieConfigMap(pravegaCluster.ObjectMeta, ownerRef, pravegaCluster.Spec.ZookeeperUri, &pravegaCluster.Spec.Bookkeeper))
	if err != nil {
		logrus.Error(err)
	}

	logrus.WithFields(logrus.Fields{"cluster": pravegaCluster.Name}).Debug("Creating Bookie StatefulSet")
	err = sdk.Create(makeBookieStatefulSet(pravegaCluster.ObjectMeta, ownerRef, &pravegaCluster.Spec.Bookkeeper))
	if err != nil {
		logrus.Error(err)
	}

	logrus.WithFields(logrus.Fields{"cluster": pravegaCluster.Name}).Debug("Creating Controller ConfigMap")
	err = sdk.Create(makeControllerConfigMap(pravegaCluster.ObjectMeta, ownerRef, pravegaCluster.Spec.ZookeeperUri, &pravegaCluster.Spec.Pravega))
	if err != nil {
		logrus.Error(err)
	}

	logrus.WithFields(logrus.Fields{"cluster": pravegaCluster.Name}).Debug("Creating Controller StatefulSet")
	err = sdk.Create(makeControllerStatefulSet(pravegaCluster.ObjectMeta, ownerRef, &pravegaCluster.Spec.Pravega))
	if err != nil {
		logrus.Error(err)
	}

	logrus.WithFields(logrus.Fields{"cluster": pravegaCluster.Name}).Debug("Creating Controller Service")
	err = sdk.Create(makeControllerService(pravegaCluster.ObjectMeta, ownerRef, &pravegaCluster.Spec.Pravega))
	if err != nil {
		logrus.Error(err)
	}

	logrus.WithFields(logrus.Fields{"cluster": pravegaCluster.Name}).Debug("Creating SegmentStore ConfigMap")
	err = sdk.Create(makeSegmentstoreConfigMap(pravegaCluster.ObjectMeta, ownerRef, pravegaCluster.Spec.ZookeeperUri, &pravegaCluster.Spec.Pravega))
	if err != nil {
		logrus.Error(err)
	}

	logrus.WithFields(logrus.Fields{"cluster": pravegaCluster.Name}).Debug("Creating SegmentStore StatefulSet")
	err = sdk.Create(makeSegmentStoreStatefulSet(pravegaCluster.ObjectMeta, ownerRef, &pravegaCluster.Spec.Pravega))
	if err != nil {
		logrus.Error(err)
	}

	logrus.WithFields(logrus.Fields{
		"cluster":   pravegaCluster.Name,
		"namespace": pravegaCluster.ObjectMeta.Namespace}).Debug("All Items Resources Created")
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
