#!/usr/bin/env bash

if [ "$#" -ne 1 ]; then
    echo "Invalid number of parameters"
		exit 1
fi

JSONPATH1="{.spec.replicas}"
JSONPATH2="{.status.readyReplicas}"
ZkOprName=zoo-op
ZkName=zoo
BkOprName=pr
BkName=br
PrOprName=foo
PrName=bar

if [ "$1" == "1" ]; then
	helm install ../charts/zookeeper-operator --name $ZkOprName
	kubectl rollout status deploy/$ZkOprName-zookeeper-operator

	helm install ../charts/zookeeper --name $ZkName

	replicas=$(kubectl get zookeepercluster $ZkName-zookeeper -o jsonpath=$JSONPATH1)
	ready=$(kubectl get zookeepercluster $ZkName-zookeeper -o jsonpath=$JSONPATH2)

	until [ "$replicas" == "$ready" ]
	do
		sleep 1;
		replicas=$(kubectl get zookeepercluster $ZkName-zookeeper -o jsonpath=$JSONPATH1)
		ready=$(kubectl get zookeepercluster $ZkName-zookeeper -o jsonpath=$JSONPATH2)
	done

	ZkSvc=$(kubectl get svc | grep $ZkName- | grep client | head -n1 | cut -d " " -f1)

	helm install ../charts/bookkeeper-operator --name $BkOprName
	kubectl rollout status deploy/$BkOprName-bookkeeper-operator

	helm install ../charts/bookkeeper --name $BkName --set zookeeperUri=$ZkSvc:2181

	replicas=$(kubectl get bookkeepercluster $BkName-pravega-bk -o jsonpath=$JSONPATH1)
	ready=$(kubectl get bookkeepercluster $BkName-pravega-bk -o jsonpath=$JSONPATH2)

	until [ "$replicas" == "$ready" ]
	do
		sleep 1;
		replicas=$(kubectl get bookkeepercluster $BkName-pravega-bk -o jsonpath=$JSONPATH1)
		ready=$(kubectl get bookkeepercluster $BkName-pravega-bk -o jsonpath=$JSONPATH2)
	done

	BkSvc=$(kubectl get svc | grep $BkName- | grep headless | head -n1 | cut -d " " -f1)

	helm install ../charts/pravega-operator --name $PrOprName
	kubectl rollout status deploy/$PrOprName-pravega-operator

	helm install ../charts/pravega --name $PrName --set zookeeperUri=$ZkSvc:2181 --set bookkeeperUri=$BkSvc:3181

elif [  "$1" == "2"  ]; then
	helm del $PrName --purge
	helm del $BkName --purge
	helm del $ZkName --purge
	helm del $PrOprName --purge
	helm del $BkOprName --purge
	helm del $ZkOprName --purge
else
	echo "Invalid argument"
	exit 1
fi
