#! /bin/bash
set -ex

echo "Running pre-upgrade script for upgrading pravega operator from version 0.4.x to 0.5.0"

if [ "$#" -ne 3 ]; then
	echo "Error : Invalid number of arguments"
	Usage: "./operatorUpgrade.sh <pravega-operator deployment name> <pravega-operator deployment namespace> <pravega-operator new image-repo/image-tag>"
	exit 1
fi

function UpgradingToPoperator(){

local op_deployment_name=$1

local namespace=$2

local op_name=`kubectl describe deploy ${op_deployment_name} -n ${namespace}| grep "Name:" | awk '{print $2}' | head -1`

local op_image=$3

sed -i "s/namespace.*/namespace: $namespace"/ ./manifest_files/secret.yaml

local temp_string_for_dns=pravega-webhook-svc.${namespace}

sed -i "s/pravega-webhook-svc.default/${temp_string_for_dns}"/ ./manifest_files/secret.yaml

#Installing the secrets 
kubectl apply -f  ./manifest_files/secret.yaml

#reverting the changes back in the secret.yaml file
sed -i "s/${temp_string_for_dns}/pravega-webhook-svc.default"/ ./manifest_files/secret.yaml

sed -i "s|cert.*|cert-manager.io/inject-ca-from: $namespace/selfsigned-cert|" ./manifest_files/webhook.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./manifest_files/webhook.yaml

#Insalling the webhook
kubectl apply -f ./manifest_files/webhook.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./manifest_files/version_map.yaml

#Insalling the version map for pravega-operator
kubectl apply -f  ./manifest_files/version_map.yaml

cabundle=`kubectl get ValidatingWebhookConfiguration pravega-webhook-config --namespace ${namespace} --output yaml | grep caBundle: | awk '{print $2}'`

sed -i "s/caBundle.*/caBundle: $cabundle "/ ./manifest_files/crd.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./manifest_files/crd.yaml

#updating the crd for pravega-operator
kubectl apply -f  ./manifest_files/crd.yaml

#Installing the bookkeeper-operator
helm install charts/bookkeeper-operator --name bkop --namespace $namespace

sed -i "s/name:.*/name: $op_name"/ ./manifest_files/role.yaml

sed -i "s/namespace:.*/namespace: $namespace "/ ./manifest_files/role.yaml

#updating the roles for pravega-operator
kubectl apply -f ./manifest_files/role.yaml

sed -i "s|image:.*|image: $op_image|" ./manifest_files/patch.yaml

sed -i "s/value:.*/value: $op_name "/ ./manifest_files/patch.yaml

sed -i "/imagePullPolicy:.*/{n;s/name.*/name: $op_name/}" ./manifest_files/patch.yaml

#updating the operator using patch file
kubectl patch deployment $op_name --namespace ${namespace} --type merge --patch "$(cat ./manifest_files/patch.yaml)"

}

UpgradingToPoperator $1 $2 $3 
