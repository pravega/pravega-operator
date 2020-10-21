#! /bin/bash
set -ex

if [[ "$#" -lt 1 || "$#" -gt 2 ]]; then
	echo "Error : Invalid number of arguments"
	Usage: "./pre-upgrade.sh <pravega-operator-release-name> <pravega-operator-namespace>"
	exit 1
fi

name=$1
namespace=${2:-default}

kubectl annotate Service pravega-webhook-svc meta.helm.sh/release-name=$name -n $namespace --overwrite
kubectl annotate Service pravega-webhook-svc meta.helm.sh/release-namespace=$namespace -n $namespace --overwrite
kubectl label Service pravega-webhook-svc app.kubernetes.io/managed-by=Helm -n $namespace --overwrite

# Deleting required resources
kubectl delete mutatingwebhookconfiguration pravega-webhook-config
kubectl delete cm pravega-operator-lock
