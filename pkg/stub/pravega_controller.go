package stub

import (
	"fmt"
	"strings"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	pravegaPassword = "pravega-passwd"
)

func makeControllerStatefulSet(metadata metav1.ObjectMeta, owner *metav1.OwnerReference, pravegaSpec *v1alpha1.PravegaSpec) *v1beta1.StatefulSet {
	return &v1beta1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      prefixedName("pravega-controller", metadata.Name),
			Namespace: metadata.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*owner,
			},
		},
		Spec: *makePravegaControllerStatefulSpec(metadata.Name, pravegaSpec),
	}
}

func makePravegaControllerStatefulSpec(name string, pravegaSpec *v1alpha1.PravegaSpec) *v1beta1.StatefulSetSpec {
	return &v1beta1.StatefulSetSpec{
		ServiceName:         "pravega-controller",
		Replicas:            &pravegaSpec.ControllerReplicas,
		PodManagementPolicy: v1beta1.ParallelPodManagement,
		Template:            *makeControllerStatefulTemplate(name, pravegaSpec),
	}
}

func makeControllerStatefulTemplate(name string, pravegaSpec *v1alpha1.PravegaSpec) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app":  name,
				"kind": "pravega-controller",
			},
		},
		Spec: *makeControllerPodSpec(name, pravegaSpec),
	}
}

func makeControllerPodSpec(name string, pravegaSpec *v1alpha1.PravegaSpec) *corev1.PodSpec {
	return &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "pravega-controller",
				Image:           pravegaSpec.Image.String(),
				ImagePullPolicy: pravegaSpec.Image.PullPolicy,
				Args: []string{
					"controller",
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "rest",
						ContainerPort: 10080,
					},
					{
						Name:          "grpc",
						ContainerPort: 9090,
					},
				},
				EnvFrom: []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: controllerConfigName(name),
							},
						},
					},
				},
			},
		},
	}
}

func makeControllerConfigMap(metadata metav1.ObjectMeta, owner *metav1.OwnerReference, zkUri string, pravegaSpec *v1alpha1.PravegaSpec) *corev1.ConfigMap {
	var javaOpts = []string{
		"-Dconfig.controller.metricenableStatistics=false",
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      controllerConfigName(metadata.Name),
			Labels:    map[string]string{"app": metadata.Name},
			Namespace: metadata.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*owner,
			},
		},
		Data: map[string]string{
			"CLUSTER_NAME":           metadata.Name,
			"ZK_URL":                 zkUri,
			"JAVA_OPTS":              strings.Join(javaOpts, " "),
			"REST_SERVER_PORT":       "10080",
			"CONTROLLER_SERVER_PORT": "9090",
			"AUTHORIZATION_ENABLED":  "false",
			"TOKEN_SIGNING_KEY":      "secret",
			"USER_PASSWORD_FILE":     "/etc/pravega/conf/passwd",
			"TLS_ENABLED":            "false",
		},
	}
}

func makeControllerService(metadata metav1.ObjectMeta, owner *metav1.OwnerReference, pravegaSpec *v1alpha1.PravegaSpec) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      prefixedName("pravega-controller", metadata.Name),
			Namespace: metadata.Namespace,
			Labels:    map[string]string{"app": metadata.Name},
			OwnerReferences: []metav1.OwnerReference{
				*owner,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "rest",
					Port: 10080,
				},
				{
					Name: "grpc",
					Port: 9090,
				},
			},
			Selector: map[string]string{
				"app":  metadata.Name,
				"kind": "pravega-controller",
			},
		},
	}
}

func controllerConfigName(name string) string {
	return prefixedName("controller-config", name)
}

func makeControllerUrl(metadata metav1.ObjectMeta) string {
	return fmt.Sprintf("tcp://%v.%v:%v", prefixedName("pravega-controller", metadata.Name), metadata.Namespace, "9090")
}
