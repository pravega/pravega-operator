#! /bin/bash
set -ex

if [ "$#" -ne 6 ]; then
	echo "Error : Invalid number of arguments"
	Usage: "./post-upgrade.sh <pravegacluster name> <pravega-release-name> <bookkeeper-release-name> <version> <namespace> <zk-svc-name>"
	exit 1
fi

name=$1
pname=$2
bkname=$3
version=$4
namespace=$5
zksvc=$6

kubectl describe PravegaCluster $name -n $namespace
currentReplicasPc=`kubectl get PravegaCluster $name -n $namespace -o jsonpath='{.status.currentReplicas}'`
readyReplicasPc=`kubectl get PravegaCluster $name -n $namespace -o jsonpath='{.status.readyReplicas}'`
if [ $currentReplicasPc != $readyReplicasPc ]; then
  exit 1
fi
## Add annotations and labels on PravegaCluster resources to make sure they're owned by the right chart.
kubectl annotate PravegaCluster $name meta.helm.sh/release-name=$pname -n $namespace --overwrite
kubectl annotate PravegaCluster $name meta.helm.sh/release-namespace=$namespace -n $namespace --overwrite
kubectl label PravegaCluster $name app.kubernetes.io/managed-by=Helm -n $namespace --overwrite

kubectl get BookkeeperCluster $name -n $namespace
currentReplicasBk=`kubectl get BookkeeperCluster $name -n $namespace -o jsonpath='{.status.replicas}'`
readyReplicasBk=`kubectl get BookkeeperCluster $name -n $namespace -o jsonpath='{.status.readyReplicas}'`
if [ $currentReplicasBk != $readyReplicasBk ]; then
  exit 2
fi
## Add annotations and labels on BookkeeperCluster resources to make sure they're owned by the right chart.
kubectl annotate BookkeeperCluster $name meta.helm.sh/release-name=$bkname -n $namespace --overwrite
kubectl annotate BookkeeperCluster $name meta.helm.sh/release-namespace=$namespace -n $namespace --overwrite
kubectl label BookkeeperCluster $name app.kubernetes.io/managed-by=Helm -n $namespace --overwrite
kubectl annotate ConfigMap pravega-config meta.helm.sh/release-name=$bkname -n $namespace --overwrite
kubectl annotate ConfigMap pravega-config meta.helm.sh/release-namespace=$namespace -n $namespace --overwrite
kubectl label ConfigMap pravega-config app.kubernetes.io/managed-by=Helm -n $namespace --overwrite

## Install the bookkeeper charts
helm repo add pravega https://charts.pravega.io
helm repo update
helm upgrade $pname pravega/pravega --version=$version --set fullnameOverride=$name --set zookeeperUri="$zksvc:2181" --set bookkeeperUri="$name-bookie-headless:3181"
helm install $bkname pravega/bookkeeper --version=$version --set fullnameOverride=$name --set zookeeperUri="$zksvc:2181"
