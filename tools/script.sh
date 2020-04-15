#! /bin/bash
set -ex

echo "Running pre-upgrade script for upgrading pravega operator from version 0.4.x to 0.5.0"

if [ "$#" -ne 2 ]; then
	echo "Error : Invalid number of arguments"
	echo "Usage: ./script.sh namespace_used_for_pravega-operator name_used_for_pravega-cluster"
	exit 1
fi

function UpgradingToPoperator(){

local namespace=$1

local name=$2

kubectl apply -f ./manifest_files/cert-manager.yaml

sed -i "s/namespace.*/namespace: $namespace"/ ./manifest_files/secret.yaml

kubectl apply -f  ./manifest_files/secret.yaml

sed -i "s|cert.*|cert-manager.io/inject-ca-from: $namespace/selfsigned-cert|" ./manifest_files/webhook.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./manifest_files/webhook.yaml

kubectl apply -f ./manifest_files/webhook.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./manifest_files/version_map.yaml

kubectl apply -f  ./manifest_files/version_map.yaml

cabundle=`kubectl get ValidatingWebhookConfiguration pravega-webhook-config --output yaml | grep caBundle: | awk '{print $2}'`

sed -i "s/caBundle.*/caBundle: $cabundle "/ ./manifest_files/crd.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./manifest_files/crd.yaml

kubectl apply -f  ./manifest_files/crd.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./manifest_files/role.yaml

kubectl apply -f ./manifest_files/role.yaml

op_name=`kubectl get deployment | grep "pravega-operator" | awk '{print $1}'`

op_image=`kubectl get deployment $op_name --output yaml | grep "image:" | awk '{print $2}'`

sed -i "/metadata:.*/{n;s/name.*/name: $op_name/}" ./manifest_files/operator.yaml

sed -i "/matchLabels:.*/{n;s/name.*/name: $op_name/}" ./manifest_files/operator.yaml

sed -i "/labels:.*/{n;s/name.*/name: $op_name/}" ./manifest_files/operator.yaml

sed -i "/containers.*/{n;s/name.*/name: $op_name/}" ./manifest_files/operator.yaml

sed -i "/name: OPERATOR_NAME.*/{n;s/value.*/value: $op_name/}" ./manifest_files/operator.yaml

sed -i "s|image:.*|image: $op_image|" ./manifest_files/operator.yaml

kubectl apply -f ./manifest_files/operator.yaml

sed -i "s/name:.*/name: ${name}-bk-supported-upgrade-paths"/ ./manifest_files/bk_version_map.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./manifest_files/bk_version_map.yaml

kubectl apply -f  ./manifest_files/bk_version_map.yaml

helm install charts/bookkeeper-operator --name bkop --namespace $namespace

}

UpgradingToPoperator $1 $2
