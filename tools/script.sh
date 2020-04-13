#! /bin/bash

if [ "$#" -ne 2 ]; then
	echo "Error : Invalid number of arguments"
	echo "Usage: ./script.sh namespace name"
	exit 1
fi


function UpgradingToPoperator(){

local namespace=$1

local name=$2

sed -i "/namespace/c \ \ namespace: $namespace " chart_new/pravega-operator/templates/webhook.yaml 

kubectl create -f ./templates/webhook.yaml

sed -i "/namespace/c \ \ name: $name" chart_new/pravega-operator/templates/version_map.yaml

kubectl create -f ./templates/version_map.yaml

kubectl create -f ./templates/secret.yaml

kubectl apply -f new_operator.yaml

kubectl apply -f new_crd.yaml

echo "helm install charts/bookkeeper --name bkop --namespace $namespace"

helm install charts/bookkeeper --name bkop --namespace $namespace

kubectl apply -f new_roles.yaml

sed -i "/namespace/c \ \ name: $name" chart_new/pravega-operator/templates/bk_version_map.yaml

kubectl create -f ./templates/bk_version_map.yaml

}

UpgradingToPoperator
