#!/bin/bash
#
# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
namespace="default"

#############################################################################
## Preparations
#############################################################################

# Grant ourselves permissions to create roles/cluster roles in k8s
echo "Creating clusterrolebinding for cluster-admin"
kubectl create clusterrolebinding default-cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)

# Create namespace if it does not exist
kubectl get namespace $namespace > /dev/null 2>&1
if [ $? != 0 ]
then
  echo "Creating Namespace: $namespace"
  kubectl create namespace $namespace
else
  echo "Using namespace: $namespace"
fi

#############################################################################
## Install Tier 2 Storage
#############################################################################
# Download helm if you haven't done that
# Choose a release from https://github.com/helm/helm/releases
echo "Initializing helm"
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller
helm repo update
sleep 10

# Install nfs-server-provisioner using helm
echo "Installing nfs-server-provisioner"
helm install stable/nfs-server-provisioner

echo "Creating tier 2 storage"
kubectl create -n $namespace -f https://raw.githubusercontent.com/pravega/pravega-operator/master/example/gke/pvc.yaml

#############################################################################
## Install Zookeeper cluster
#############################################################################
echo "Creating Zookeeper Operator"
kubectl create -n $namespace -f https://raw.githubusercontent.com/pravega/zookeeper-operator/master/deploy/crds/zookeeper_v1beta1_zookeepercluster_crd.yaml
kubectl create -n $namespace -f https://raw.githubusercontent.com/pravega/zookeeper-operator/master/deploy/all_ns/rbac.yaml
kubectl create -n $namespace -f https://raw.githubusercontent.com/pravega/zookeeper-operator/master/deploy/all_ns/operator.yaml

echo "Creating Zookeeper Cluster"
kubectl create -n $namespace -f https://raw.githubusercontent.com/pravega/pravega-operator/master/example/gke/zk.yaml

#############################################################################
## Install Pravega cluster
#############################################################################
echo "Creating Pravega Operator"
kubectl create -n $namespace -f https://raw.githubusercontent.com/pravega/pravega-operator/master/deploy/crd.yaml
kubectl create -n $namespace -f https://raw.githubusercontent.com/pravega/pravega-operator/master/deploy/role.yaml
kubectl create -n $namespace -f https://raw.githubusercontent.com/pravega/pravega-operator/master/deploy/role_binding.yaml
kubectl create -n $namespace -f https://raw.githubusercontent.com/pravega/pravega-operator/master/deploy/service_account.yaml
kubectl create -n $namespace -f https://raw.githubusercontent.com/pravega/pravega-operator/master/deploy/operator.yaml

echo "Creating Pravega Cluster"
kubectl create -n $namespace -f https://raw.githubusercontent.com/pravega/pravega-operator/master/example/gke/pravega.yaml