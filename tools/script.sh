#! /bin/bash
#set +ex

echo "This is a pre upgrade script required for updating pravega operator <= 0.4.x to >=0.5.x "

if [ "$#" -eq 2 ]; then
	echo "Error : Invalid number of arguments"
	echo "Usage: ./script.sh namespace(pravega operator) name(pravega cluster)"
	exit 1
fi

function UpgradingToPoperator(){

local namespace=$1

local name=$2

sed "s|cert.*|cert-manager.io/inject-ca-from: $namespace/selfsigned-cert|" ./manifest_files/webhook.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./manifest_files/webhook.yaml

kubectl apply -f ./manifest_files/webhook.yaml

kubectl apply -f  ./manifest_files/version_map.yaml

kubectl apply -f "https://github.com/jetstack/cert-manager/releases/download/v0.14.1/cert-manager.yaml"

sed -i "s/namespace.*/namespace: $namespace"/ ./manifest_files/secret.yaml

kubectl apply -f  ./manifest_files/secret.yaml

kubectl patch deployment $op_name --type merge --patch "$(cat ./manifest_files/patch.yaml)"

cabundle=`kubectl get ValidatingWebhookConfiguration pravega-webhook-config --output yaml | grep caBundle: | awk '{print $2}'`

sed -i "s/caBundle.*/caBundle: $cabundle "/ ./manifest_files/crd.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./manifest_files/crd.yaml

kubectl apply -f  ./manifest_files/crd.yaml

sed -i "s/name.*/name: ${name}-bk-supported-upgrade-paths"/ ./manifest_files/bk_version_map.yaml

kubectl apply -f  ./manifest_files/bk_version_map.yaml

helm install charts/bookkeeper-operator --name bkop --namespace $namespace

kubectl apply -f ./manifest_files/role.yaml

}


UpgradingToPoperator $1 $2
