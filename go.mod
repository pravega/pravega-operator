module github.com/pravega/pravega-operator

go 1.13

require (
	4d63.com/gochecknoglobals v0.0.0-20190306162314-7c3491d2b6ec // indirect
	4d63.com/gochecknoinits v0.0.0-20200108094044-eb73b47b9fc4 // indirect
	github.com/alecthomas/gocyclo v0.0.0-20150208221726-aa8f8b160214 // indirect
	github.com/alexkohler/nakedret v1.0.0 // indirect
	github.com/gordonklaus/ineffassign v0.0.0-20200309095847-7953dde2c7bf // indirect
	github.com/hashicorp/go-version v1.1.0
	github.com/jgautheron/goconst v0.0.0-20200227150835-cda7ea3bf591 // indirect
	github.com/mdempsky/unconvert v0.0.0-20200228143138-95ecdbfc0b5f // indirect
	github.com/mibk/dupl v1.0.0 // indirect
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/opennota/check v0.0.0-20180911053232-0c771f5545ff // indirect
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/pravega/bookkeeper-operator v0.1.1-rc0
	github.com/securego/gosec v0.0.0-20200401082031-e946c8c39989 // indirect
	github.com/sirupsen/logrus v1.5.0
	github.com/tsenart/deadcode v0.0.0-20160724212837-210d2dc333e9 // indirect
	github.com/walle/lll v1.0.1 // indirect
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6
	mvdan.cc/interfacer v0.0.0-20180901003855-c20040233aed // indirect
	mvdan.cc/lint v0.0.0-20170908181259-adc824a0674b // indirect
	mvdan.cc/unparam v0.0.0-20200501210554-b37ab49443f7 // indirect
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.17.4 // Required by prometheus-operator
)
