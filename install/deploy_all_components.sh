#!/usr/bin/env bash

JSONPATH1="{.spec.replicas}"
JSONPATH2="{.status.readyReplicas}"

helm install ../charts/zookeeper-operator --name zoo-op
kubectl rollout status deploy/zoo-op-zookeeper-operator

helm install ../charts/zookeeper --name zoo

replicas=$(kubectl get zookeepercluster zoo-zookeeper -o jsonpath=$JSONPATH1)
ready=$(kubectl get zookeepercluster zoo-zookeeper -o jsonpath=$JSONPATH2)

until [ "$replicas" == "$ready" ]
do
	sleep 1;
	replicas=$(kubectl get zookeepercluster zoo-zookeeper -o jsonpath=$JSONPATH1)
	ready=$(kubectl get zookeepercluster zoo-zookeeper -o jsonpath=$JSONPATH2)
done

helm install ../charts/bookkeeper-operator --name pr
kubectl rollout status deploy/pr-bookkeeper-operator

helm install ../charts/bookkeeper --name br --set zookeeperUri=zoo-zookeeper-client:2181

replicas=$(kubectl get bookkeepercluster br-pravega -o jsonpath=$JSONPATH1)
ready=$(kubectl get bookkeepercluster br-pravega -o jsonpath=$JSONPATH2)

until [ "$replicas" == "$ready" ]
do
	sleep 1;
	replicas=$(kubectl get bookkeepercluster br-pravega -o jsonpath=$JSONPATH1)
	ready=$(kubectl get bookkeepercluster br-pravega -o jsonpath=$JSONPATH2)
done

helm install ../charts/pravega-operator --name foo
kubectl rollout status deploy/foo-pravega-operator

helm install ../charts/pravega --name bar --set zookeeperUri=zoo-zookeeper-client:2181 --set bookkeeperUri=br-pravega-bookie-headless:3181
