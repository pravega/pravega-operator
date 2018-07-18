package stub

import (
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LedgerDiskName  = "ledger"
	JournalDiskName = "journal"
)

func createBookie(ownerRef *metav1.OwnerReference, pravegaCluster *v1alpha1.PravegaCluster) {
	var err = sdk.Create(makeBookieConfigMap(pravegaCluster.ObjectMeta, ownerRef, pravegaCluster.Spec.ZookeeperUri, &pravegaCluster.Spec.Bookkeeper))
	if err != nil && !errors.IsAlreadyExists(err) {
		logrus.Error(err)
	}

	err = sdk.Create(makeBookieStatefulSet(pravegaCluster.ObjectMeta, ownerRef, &pravegaCluster.Spec.Bookkeeper))
	if err != nil && !errors.IsAlreadyExists(err) {
		logrus.Error(err)
	}
}

func destroyBookieVolumes(metadata metav1.ObjectMeta) {
	logrus.WithFields(logrus.Fields{"name": metadata.Name}).Info("Destroying Bookie volumes")

	err := deleteCollection("v1", "PersistentVolumeClaim", metadata.Namespace, fmt.Sprintf("app=%v,kind=bookie", metadata.Name))
	if err != nil {
		logrus.Error(err)
	}
}

func makeBookieStatefulSet(metadata metav1.ObjectMeta, owner *metav1.OwnerReference, bookkeeperSpec *v1alpha1.BookkeeperSpec) *v1beta1.StatefulSet {
	return &v1beta1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      generateKindName("bookie", metadata.Name),
			Namespace: metadata.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*owner,
			},
		},
		Spec: *makeBookieStatefulSetSpec(metadata.Name, bookkeeperSpec),
	}
}

func makeBookieStatefulSetSpec(name string, spec *v1alpha1.BookkeeperSpec) *v1beta1.StatefulSetSpec {
	return &v1beta1.StatefulSetSpec{
		ServiceName:          "bookkeeper",
		Replicas:             &spec.Replicas,
		PodManagementPolicy:  v1beta1.ParallelPodManagement,
		Template:             *makeBookieStatefulTemplate(name, spec),
		VolumeClaimTemplates: makeBookieVolumeClaimTemplates(name, spec),
	}
}

func makeBookieStatefulTemplate(name string, spec *v1alpha1.BookkeeperSpec) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app":  name,
				"kind": "bookie",
			},
		},
		Spec: *makeBookiePodSpec(name, spec),
	}
}

func makeBookiePodSpec(name string, bookkeeperSpec *v1alpha1.BookkeeperSpec) *corev1.PodSpec {
	return &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "bookie",
				Image:           bookkeeperSpec.Image.String(),
				ImagePullPolicy: bookkeeperSpec.Image.PullPolicy,
				Command: []string{
					"/bin/bash", "/opt/bookkeeper/entrypoint.sh",
				},
				Args: []string{
					"/opt/bookkeeper/bin/bookkeeper", "bookie",
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "bookie",
						ContainerPort: 3181,
					},
				},
				EnvFrom: []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: bookieConfigName(name),
							},
						},
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      LedgerDiskName,
						MountPath: "/bk/journal",
					},
					{
						Name:      JournalDiskName,
						MountPath: "/bk/ledgers",
					},
				},
			},
		},
	}
}

func makeBookieVolumeClaimTemplates(name string, spec *v1alpha1.BookkeeperSpec) []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   JournalDiskName,
				Labels: map[string]string{"app": name},
			},
			Spec: spec.Storage.JournalVolumeClaimTemplate,
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   LedgerDiskName,
				Labels: map[string]string{"app": name},
			},
			Spec: spec.Storage.LedgerVolumeClaimTemplate,
		},
	}
}

func makeBookieConfigMap(metadata metav1.ObjectMeta, owner *metav1.OwnerReference, zkUri string, bookkeeperSpec *v1alpha1.BookkeeperSpec) *corev1.ConfigMap {
	// TODO: Add Spec Options

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bookieConfigName(metadata.Name),
			Namespace: metadata.Namespace,
			Labels:    map[string]string{"app": metadata.Name},
			OwnerReferences: []metav1.OwnerReference{
				*owner,
			},
		},
		Data: map[string]string{
			"BK_BOOKIE_EXTRA_OPTS":     "-Xms1g -Xmx1g -XX:MaxDirectMemorySize=1g -XX:+UseG1GC  -XX:MaxGCPauseMillis=10 -XX:+ParallelRefProcEnabled -XX:+UnlockExperimentalVMOptions -XX:+AggressiveOpts -XX:+DoEscapeAnalysis -XX:ParallelGCThreads=32 -XX:ConcGCThreads=32 -XX:G1NewSizePercent=50 -XX:+DisableExplicitGC -XX:-ResizePLAB",
			"ZK_URL":                   zkUri,
			"BK_useHostNameAsBookieID": "true",
			"PRAVEGA_CLUSTER_NAME":     metadata.Name,
		},
	}
}

func bookieConfigName(name string) string {
	return generateKindName("bookie-config", name)
}
