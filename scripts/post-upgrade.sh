#! /bin/bash
set -ex

if [[ "$#" -lt 4 || "$#" -gt 6 ]]; then
	echo "Error : Invalid number of arguments"
	Usage: "./post-upgrade.sh <pravegacluster name> <pravega-release-name> <bookkeeper-release-name> <version> <namespace> <zk-svc-name>"
	exit 1
fi

name=$1
pname=$2
bkname=$3
version=$4
namespace=${5:-default}
zksvc=${6:-zookeeper-client}

echo "Checking that the PravegaCluster resource is in ready state"
kubectl describe PravegaCluster $name -n $namespace
currentReplicasPc=`kubectl get PravegaCluster $name -n $namespace -o jsonpath='{.status.currentReplicas}'`
readyReplicasPc=`kubectl get PravegaCluster $name -n $namespace -o jsonpath='{.status.readyReplicas}'`
if [ $currentReplicasPc != $readyReplicasPc ]; then
	echo "Error : Pravega Cluster is not in ready state"
  exit 1
fi
echo "Adding required annotations and labels to the PravegaCluster resource to make sure they're owned by the right chart"
kubectl annotate PravegaCluster $name meta.helm.sh/release-name=$pname -n $namespace --overwrite
kubectl annotate PravegaCluster $name meta.helm.sh/release-namespace=$namespace -n $namespace --overwrite
kubectl label PravegaCluster $name app.kubernetes.io/managed-by=Helm -n $namespace --overwrite

echo "Checking that the BookkeeperCluster resource is in ready state"
kubectl get BookkeeperCluster $name -n $namespace
currentReplicasBk=`kubectl get BookkeeperCluster $name -n $namespace -o jsonpath='{.status.replicas}'`
readyReplicasBk=`kubectl get BookkeeperCluster $name -n $namespace -o jsonpath='{.status.readyReplicas}'`
if [ $currentReplicasBk != $readyReplicasBk ]; then
	echo "Error : Bookkeeper Cluster is not in ready state"
  exit 2
fi
echo "Adding required annotations and labels to the BookkeeperCluster resource to make sure they're owned by the right chart"
kubectl annotate BookkeeperCluster $name meta.helm.sh/release-name=$bkname -n $namespace --overwrite
kubectl annotate BookkeeperCluster $name meta.helm.sh/release-namespace=$namespace -n $namespace --overwrite
kubectl label BookkeeperCluster $name app.kubernetes.io/managed-by=Helm -n $namespace --overwrite
kubectl annotate ConfigMap pravega-config meta.helm.sh/release-name=$bkname -n $namespace --overwrite
kubectl annotate ConfigMap pravega-config meta.helm.sh/release-namespace=$namespace -n $namespace --overwrite
kubectl label ConfigMap pravega-config app.kubernetes.io/managed-by=Helm -n $namespace --overwrite

echo "Installing / Upgrading the pravega and bookkeeper charts"
helm repo add pravega https://charts.pravega.io
helm repo update
helm upgrade $pname pravega/pravega --version=$version --set fullnameOverride=$name --set zookeeperUri="$zksvc:2181" --set bookkeeperUri="$name-bookie-headless:3181"
helm install $bkname pravega/bookkeeper --version=$version --set fullnameOverride=$name --set zookeeperUri="$zksvc:2181"
