#!/usr/bin/env bash
set -ex

if [ "$#" -ne 1 ]; then
	echo "Error : Invalid number of arguments"
	echo "Usage: ./pravegacluster.sh [deploy/destroy]"
	exit 1
fi

deploy_cluster () {
	# Installing the Zookeeper Operator
	helm install zookeeper-operator ../charts/zookeeper-operator --wait

	# Installing the Zookeeper Cluster
	helm install zookeeper ../charts/zookeeper --wait

	# Installing the BookKeeper Operator
	helm install bookkeeper-operator ../charts/bookkeeper-operator --wait

	# Installing the BookKeeper Cluster
	helm install bookkeeper ../charts/bookkeeper --wait

	# Installing the issuer and certificate
	set +ex
	kubectl create -f ../deploy/certificate.yaml
	set -ex
	tls=$(kubectl get secret selfsigned-cert-tls -o yaml | grep tls.crt)
	crt=${tls/" tls.crt: "/}

	# Installing the Pravega Operator
	helm install pravega-operator ../charts/pravega-operator --set webhookCert.crt="$crt" --wait

	# Installing the Pravega Cluster
	helm install pravega ../charts/pravega --wait
}

destroy_cluster(){
	set +ex
	helm uninstall pravega
	helm uninstall bookkeeper
	helm uninstall zookeeper
	helm uninstall pravega-operator
	helm uninstall bookkeeper-operator
	helm uninstall zookeeper-operator
	set -ex
}

if [ $1 == "deploy" ]; then
	deploy_cluster

elif [ $1 == "destroy" ]; then
	destroy_cluster

else
	echo "Error: Invalid argument"
	echo "Use [deploy] to deploy the cluster or [destroy] to remove the existing setup"
	exit 1
fi
