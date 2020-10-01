#!/usr/bin/env bash
set -ex

if [ "$#" -lt 1 ]; then
	echo "Error : Invalid number of arguments"
	echo "Usage: ./pravegacluster.sh [deploy/destroy]"
	exit 1
fi

args=$#
cert=${2:-selfsigned-cert-bk}
secret=${3:-selfsigned-cert-tls-bk}

deploy_cluster () {
	# Adding and updating the pravega charts repo
	helm repo add pravega https://charts.pravega.io
	helm repo update

	# Installing the Zookeeper Operator
	helm install zookeeper-operator pravega/zookeeper-operator --wait

	# Installing the Zookeeper Cluster
	helm install zookeeper pravega/zookeeper --wait

	# Installing the BookKeeper Operator
	if [[ "$args" -gt 1 ]]; then
		helm install bookkeeper-operator pravega/bookkeeper-operator --set webhookCert.certName=$cert --set webhookCert.secretName=$secret --wait
	else
		helm install bookkeeper-operator pravega/bookkeeper-operator --wait
	fi

	# Installing the BookKeeper Cluster
	helm install bookkeeper pravega/bookkeeper --wait

	# Installing the issuer and certificate
	set +ex
	kubectl create -f ../deploy/certificate.yaml
	set -ex

	# Installing the Pravega Operator
	helm install pravega-operator pravega/pravega-operator --wait

	# Installing the Pravega Cluster
	helm install pravega pravega/pravega --wait
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
