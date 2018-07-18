package pravega

import (
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
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

	return nil
}
