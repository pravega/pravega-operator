module github.com/pravega/pravega-operator

go 1.13

require (
	4d63.com/gochecknoinits v0.0.0-20200108094044-eb73b47b9fc4 // indirect
	github.com/hashicorp/go-version v1.1.0
	github.com/mdempsky/maligned v0.0.0-20180708014732-6e39bd26a8c8 // indirect
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/pravega/bookkeeper-operator v0.1.1-rc0
	github.com/pravega/zookeeper-operator v0.2.8
	github.com/rogpeppe/go-internal v1.5.2 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
	github.com/securego/gosec v0.0.0-20200401082031-e946c8c39989 // indirect
	github.com/sirupsen/logrus v1.5.0
	github.com/stripe/safesql v0.2.0 // indirect
	github.com/tsenart/deadcode v0.0.0-20160724212837-210d2dc333e9 // indirect
	github.com/walle/lll v1.0.1 // indirect
	golang.org/x/tools v0.0.0-20200426102838-f3a5411a4c3b // indirect
	k8s.io/api v0.17.5
	k8s.io/apimachinery v0.17.5
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM

	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm

	k8s.io/api => k8s.io/api v0.17.5

	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.5

	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6-beta.0

	k8s.io/apiserver => k8s.io/apiserver v0.17.5

	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.5

	k8s.io/client-go => k8s.io/client-go v0.17.5

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
)
