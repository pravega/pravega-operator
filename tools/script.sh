#! /bin/bash
if [ "$#" -lt 1 ]; then
	echo "Error : Invalid number of arguments"
	echo "Usage: ./script.sh install namespace name or ./script delete"
	exit 1
fi

function UpgradingToPoperator(){

local namespace=$1

local name=$2

sed -i "s/namespace.*/namespace: $namespace "/ ./webhook.yaml

kubectl apply -f ./webhook.yaml

sed -i "s/name.*/name: ${name}-supported-upgrade-paths"/ ./version_map.yaml

kubectl create -f  ./version_map.yaml

kubectl apply -f "https://github.com/jetstack/cert-manager/releases/download/v0.14.1/cert-manager.yaml"

sed -i "s/namespace.*/namespace: $namespace"/ ./secret.yaml

kubectl create -f  ./secret.yaml

op_name=`kubectl get deployment | grep "pravega-operator" | awk '{print $1}'`

sed -i "/configMap.*/{n;s/name.*/name: ${name}-supported-upgrade-paths/}" ./patch.yaml

kubectl patch deployment $op_name --type merge --patch "$(cat patch.yaml)"

cabundle=`kubectl get ValidatingWebhookConfiguration pravega-webhook-config --output yaml | grep caBundle: | awk '{print $2}'`

sed -i "s/caBundle.*/caBundle: $cabundle "/ ./crd.yaml

sed -i "s/namespace.*/namespace: $namespace "/ ./crd.yaml

kubectl apply -f  ./crd.yaml

sed -i "s/name.*/name: ${name}-bk-supported-upgrade-paths"/ ./bk_version_map.yaml

kubectl create -f  ./bk_version_map.yaml

helm install charts/bookkeeper --name bkop --namespace $namespace

kubectl apply -f ./role.yaml

}

function deletepoperator(){
	
kubectl delete -f ./webhook.yaml
	
kubectl delete -f  ./version_map.yaml
	
kubectl delete -f  ./secret.yaml
	
kubectl delete -f  ./operator.yaml

kubectl delete -f  ./crd.yaml

kubectl delete -f ./bk_version_map.yaml

helm delete bkop --purge

kubectl delete -f ./roles.yaml 

}

if [ $1 == "install" ]; then
	UpgradingToPoperator $2 $3

elif [ $1 == "delete" ]; then
	deletepoperator

else
	echo "Error: Invalid argument"
	echo "Use [install] to install the cluster or [delete] to remove the existing setup"
	exit 1
fi