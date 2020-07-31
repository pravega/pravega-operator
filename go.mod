module github.com/pravega/pravega-operator

go 1.13

require (
	github.com/hashicorp/go-version v1.1.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/pravega/bookkeeper-operator v0.1.1-rc0
	github.com/rogpeppe/go-internal v1.5.2 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
	github.com/sirupsen/logrus v1.5.0
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/tools v0.0.0-20200426102838-f3a5411a4c3b // indirect
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.17.4 // Required by prometheus-operator
)
