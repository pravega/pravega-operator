package pravega_test

import (
	"testing"

	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/pravega/pravega-operator/pkg/controller/pravega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBookie(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pravega")
}

var _ = Describe("Bookie", func() {

	var _ = Describe("Bookkeeper Test", func() {
		var (
			p *v1alpha1.PravegaCluster
		)

		BeforeEach(func() {
			p = &v1alpha1.PravegaCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
		})

		Context("Cluster with version 0.3.0", func() {
			var (
				customReq *corev1.ResourceRequirements
				err       error
			)

			BeforeEach(func() {
				customReq = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("4"),
						corev1.ResourceMemory: resource.MustParse("6Gi"),
					},
				}
				p.Spec = v1alpha1.ClusterSpec{
					Version: "0.3.0",
					Bookkeeper: &v1alpha1.BookkeeperSpec{
						Replicas:           5,
						Resources:          customReq,
						ServiceAccountName: "pravega-components",
						Image: &v1alpha1.BookkeeperImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Repository: "foo/bookkeeper",
							},
						},
						BookkeeperJVMOptions: &v1alpha1.BookkeeperJVMOptions{
							MemoryOpts:    []string{"-Xms2g", "-XX:MaxDirectMemorySize=2g"},
							GcOpts:        []string{"-XX:MaxGCPauseMillis=20", "-XX:-UseG1GC"},
							GcLoggingOpts: []string{"-XX:NumberOfGCLogFiles=10"},
						},
						Options: map[string]string{
							"dummy-key": "dummy-value",
						},
					},
					Pravega: &v1alpha1.PravegaSpec{
						ControllerReplicas:    2,
						SegmentStoreReplicas:  4,
						ControllerResources:   customReq,
						SegmentStoreResources: customReq,
						Image: &v1alpha1.PravegaImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Repository: "bar/pravega",
							},
						},
						ControllerJvmOptions:   []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMFraction=1"},
						SegmentStoreJVMOptions: []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMFraction=1"},
					},
					TLS: &v1alpha1.TLSPolicy{
						Static: &v1alpha1.StaticTLS{
							ControllerSecret:   "controller-secret",
							SegmentStoreSecret: "segmentstore-secret",
						},
					},
				}
				p.WithDefaults()
			})

			Context("First reconcile", func() {
				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})
			})

			Context("Bookkeeper", func() {

				It("should create a headless-service", func() {
					_ = pravega.MakeBookieHeadlessService(p)
					Ω(err).Should(BeNil())
				})

				It("should create a pod disruption budget", func() {
					_ = pravega.MakeBookiePodDisruptionBudget(p)
					Ω(err).Should(BeNil())
				})

				It("should create a config-map", func() {
					_ = pravega.MakeBookieConfigMap(p)
					Ω(err).Should(BeNil())
				})

				It("should create a statefulset", func() {
					_ = pravega.MakeBookieStatefulSet(p)
					Ω(err).Should(BeNil())
				})

			})

		})

		Context("Cluster with version 0.5.0", func() {
			var (
				customReq *corev1.ResourceRequirements
				err       error
			)

			BeforeEach(func() {
				customReq = &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("4Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("4"),
						corev1.ResourceMemory: resource.MustParse("6Gi"),
					},
				}
				p.Spec = v1alpha1.ClusterSpec{
					Version: "0.5.0",
					Bookkeeper: &v1alpha1.BookkeeperSpec{
						Replicas:           5,
						Resources:          customReq,
						ServiceAccountName: "pravega-components",
						Image: &v1alpha1.BookkeeperImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Repository: "foo/bookkeeper",
							},
						},
						BookkeeperJVMOptions: &v1alpha1.BookkeeperJVMOptions{
							MemoryOpts:    []string{"-Xms2g", "-XX:MaxDirectMemorySize=2g"},
							GcOpts:        []string{"-XX:MaxGCPauseMillis=20", "-XX:-UseG1GC"},
							GcLoggingOpts: []string{"-XX:NumberOfGCLogFiles=10"},
						},
					},
					Pravega: &v1alpha1.PravegaSpec{
						ControllerReplicas:    2,
						SegmentStoreReplicas:  4,
						ControllerResources:   customReq,
						SegmentStoreResources: customReq,
						Image: &v1alpha1.PravegaImageSpec{
							ImageSpec: v1alpha1.ImageSpec{
								Repository: "bar/pravega",
							},
						},
						ControllerJvmOptions:   []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMFraction=1"},
						SegmentStoreJVMOptions: []string{"-XX:MaxDirectMemorySize=1g", "-XX:MaxRAMFraction=1"},
					},
					TLS: &v1alpha1.TLSPolicy{
						Static: &v1alpha1.StaticTLS{
							ControllerSecret:   "controller-secret",
							SegmentStoreSecret: "segmentstore-secret",
						},
					},
				}
				p.WithDefaults()
			})

			Context("First reconcile", func() {
				It("shouldn't error", func() {
					Ω(err).Should(BeNil())
				})
			})

			Context("Bookkeeper", func() {

				It("should create a headless-service", func() {
					_ = pravega.MakeBookieHeadlessService(p)
					Ω(err).Should(BeNil())
				})

				It("should create a pod disruption budget", func() {
					_ = pravega.MakeBookiePodDisruptionBudget(p)
					Ω(err).Should(BeNil())
				})

				It("should create a config-map", func() {
					_ = pravega.MakeBookieConfigMap(p)
					Ω(err).Should(BeNil())
				})

				It("should create a statefulset", func() {
					_ = pravega.MakeBookieStatefulSet(p)
					Ω(err).Should(BeNil())
				})

			})

		})

	})
})
