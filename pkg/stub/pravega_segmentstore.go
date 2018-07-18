package stub

import (
	"strings"

	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	api "github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	cacheVolumeName       = "cache"
	cacheVolumeMountPoint = "/tmp/pravega/cache"
	tier2FileMountPoint   = "/mnt/tier2"
	tier2VolumeName       = "tier2"
	segmentStoreKind      = "pravega-segmentstore"
)

func createSegmentStore(ownerRef *metav1.OwnerReference, pravegaCluster *api.PravegaCluster) {
	err := sdk.Create(makeSegmentstoreConfigMap(pravegaCluster.ObjectMeta, ownerRef, pravegaCluster.Spec.ZookeeperUri, &pravegaCluster.Spec.Pravega))
	if err != nil && !errors.IsAlreadyExists(err) {
		logrus.Error(err)
	}

	err = sdk.Create(makeSegmentStoreStatefulSet(pravegaCluster.ObjectMeta, ownerRef, &pravegaCluster.Spec.Pravega))
	if err != nil && !errors.IsAlreadyExists(err) {
		logrus.Error(err)
	}
}

func destroySegmentstoreCacheVolumes(metadata metav1.ObjectMeta) {
	logrus.WithFields(logrus.Fields{"name": metadata.Name}).Info("Destroying SegmentStore Cache volumes")

	err := deleteCollection("v1", "PersistentVolumeClaim", metadata.Namespace, fmt.Sprintf("app=%v,kind=%v", metadata.Name, segmentStoreKind))
	if err != nil {
		logrus.Error(err)
	}
}

func makeSegmentStoreStatefulSet(metadata metav1.ObjectMeta, owner *metav1.OwnerReference, pravegaSpec *api.PravegaSpec) *v1beta1.StatefulSet {
	return &v1beta1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      generateKindName(segmentStoreKind, metadata.Name),
			Namespace: metadata.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*owner,
			},
		},
		Spec: *makePravegaSegmentstoreStatefulSpec(metadata.Name, pravegaSpec),
	}
}

func makePravegaSegmentstoreStatefulSpec(name string, pravegaSpec *api.PravegaSpec) *v1beta1.StatefulSetSpec {
	return &v1beta1.StatefulSetSpec{
		ServiceName:          "segmentstore",
		Replicas:             &pravegaSpec.SegmentStoreReplicas,
		PodManagementPolicy:  v1beta1.ParallelPodManagement,
		Template:             *makeSegmentstoreStatefulTemplate(name, pravegaSpec),
		VolumeClaimTemplates: makeCacheVolumeClaimTemplate(name, pravegaSpec),
	}
}

func makeSegmentstoreStatefulTemplate(name string, pravegaSpec *api.PravegaSpec) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app":  name,
				"kind": segmentStoreKind,
			},
		},
		Spec: *makeSegmentstorePodSpec(name, pravegaSpec),
	}
}

func makeSegmentstorePodSpec(name string, pravegaSpec *api.PravegaSpec) *corev1.PodSpec {
	environment := []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: segmentstoreConfigName(name),
				},
			},
		},
	}

	environment = addTier2Secrets(environment, pravegaSpec)

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "segmentstore",
				Image:           pravegaSpec.Image.String(),
				ImagePullPolicy: pravegaSpec.Image.PullPolicy,
				Args: []string{
					"segmentstore",
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "server",
						ContainerPort: 12345,
					},
				},
				EnvFrom: environment,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      cacheVolumeName,
						MountPath: cacheVolumeMountPoint,
					},
				},
			},
		},
	}

	addTier2FilesystemVolumes(&podSpec, pravegaSpec)

	return &podSpec
}

func makeSegmentstoreConfigMap(metadata metav1.ObjectMeta, owner *metav1.OwnerReference, zkUri string, pravegaSpec *api.PravegaSpec) *corev1.ConfigMap {
	javaOpts := []string{
		"-Dconfig.controller.metricenableStatistics=false",
		"-Dpravegaservice.clusterName=" + metadata.Name,
	}

	configData := map[string]string{
		"AUTHORIZATION_ENABLED": "false",
		"CLUSTER_NAME":          metadata.Name,
		"ZK_URL":                zkUri,
		"JAVA_OPTS":             strings.Join(javaOpts, " "),
		"CONTROLLER_URL":        makeControllerUrl(metadata),
	}

	if pravegaSpec.DebugLogging {
		configData["log.level"] = "DEBUG"
	}

	for k, v := range getTier2StorageOptions(pravegaSpec) {
		configData[k] = v
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      segmentstoreConfigName(metadata.Name),
			Namespace: metadata.Namespace,
			Labels:    map[string]string{"app": metadata.Name},
			OwnerReferences: []metav1.OwnerReference{
				*owner,
			},
		},
		Data: configData,
	}
}

func makeCacheVolumeClaimTemplate(name string, pravegaSpec *api.PravegaSpec) []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   cacheVolumeName,
				Labels: map[string]string{"app": name},
			},
			Spec: pravegaSpec.CacheVolumeClaimTemplate,
		},
	}
}

func getTier2StorageOptions(pravegaSpec *api.PravegaSpec) map[string]string {
	if pravegaSpec.Tier2.FileSystem != nil {
		return map[string]string{
			"TIER2_STORAGE": "FILESYSTEM",
			"NFS_MOUNT":     tier2FileMountPoint,
		}
	}

	if pravegaSpec.Tier2.Ecs != nil {
		// EXTENDEDS3_ACCESS_KEY_ID & EXTENDEDS3_SECRET_KEY will come from secret storage
		return map[string]string{
			"TIER2_STORAGE":        "EXTENDEDS3",
			"EXTENDEDS3_BUCKET":    pravegaSpec.Tier2.Ecs.Bucket,
			"EXTENDEDS3_URI":       pravegaSpec.Tier2.Ecs.Uri,
			"EXTENDEDS3_ROOT":      pravegaSpec.Tier2.Ecs.Root,
			"EXTENDEDS3_NAMESPACE": pravegaSpec.Tier2.Ecs.Namespace,
		}
	}

	if pravegaSpec.Tier2.Hdfs != nil {
		return map[string]string{
			"TIER2_STORAGE": "HDFS",
			"HDFS_URL":      pravegaSpec.Tier2.Hdfs.Uri,
			"HDFS_ROOT":     pravegaSpec.Tier2.Hdfs.Root,
		}
	}

	return make(map[string]string)
}

func addTier2Secrets(environment []corev1.EnvFromSource, pravegaSpec *api.PravegaSpec) []corev1.EnvFromSource {
	if pravegaSpec.Tier2.Ecs != nil {
		return append(environment, corev1.EnvFromSource{
			Prefix: "EXTENDEDS3_",
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: pravegaSpec.Tier2.Ecs.Credentials,
				},
			},
		})
	}

	return environment
}

func addTier2FilesystemVolumes(podSpec *corev1.PodSpec, pravegaSpec *api.PravegaSpec) {

	if pravegaSpec.Tier2.FileSystem != nil {
		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      tier2VolumeName,
			MountPath: tier2FileMountPoint,
		})

		podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
			Name: tier2VolumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &pravegaSpec.Tier2.FileSystem.PersistentVolumeClaim,
			},
		})
	}
}

func segmentstoreConfigName(name string) string {
	return generateKindName("segmentstore-config", name)
}
