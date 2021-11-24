module github.com/pravega/pravega-operator

go 1.13

require (
	github.com/hashicorp/go-version v1.1.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/operator-framework/operator-lib v0.7.0
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/pravega/bookkeeper-operator v0.1.3
	github.com/pravega/zookeeper-operator v0.2.8
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6 // indirect
	golang.org/x/tools v0.1.5 // indirect
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73
	sigs.k8s.io/controller-runtime v0.9.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM

	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm

	github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0

	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.12.0

	github.com/onsi/gomega => github.com/onsi/gomega v1.9.0

	k8s.io/api => k8s.io/api v0.19.3

	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.5

	k8s.io/apimachinery => k8s.io/apimachinery v0.19.14-rc.0

	k8s.io/apiserver => k8s.io/apiserver v0.17.5

	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.5

	k8s.io/client-go => k8s.io/client-go v0.19.13

	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.5

	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.5

	k8s.io/code-generator => k8s.io/code-generator v0.17.6-beta.0

	k8s.io/component-base => k8s.io/component-base v0.17.5

	k8s.io/cri-api => k8s.io/cri-api v0.17.7-rc.0

	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.5

	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.5

	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.5

	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.5

	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.5

	k8s.io/kubectl => k8s.io/kubectl v0.17.5

	k8s.io/kubelet => k8s.io/kubelet v0.17.5

	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.5

	k8s.io/metrics => k8s.io/metrics v0.17.5

	k8s.io/node-api => k8s.io/node-api v0.17.5

	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.5

	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.17.5

	k8s.io/sample-controller => k8s.io/sample-controller v0.17.5

	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.5
)
