#!/usr/bin/env bash

if [ "$#" -ne 1 ]; then
  echo "Error : Invalid number of arguments"
  echo "Usage: ./pravegacluster.sh [install/delete]"
	exit 1
fi

jsonpath1="{.spec.replicas}"
jsonpath2="{.status.readyReplicas}"
zk_opr_name=zoo-op
zk_name=zoo
bk_opr_name=pr
bk_name=br
pr_opr_name=foo
pr_name=bar

if [ $1 == "install" ]; then

  zk_chart=$(cat ../charts/zookeeper/Chart.yaml | grep name:)
  arr=(`echo ${zk_chart}`)
  zk_chart_name=${arr[1]}
  echo $zk_chart_name

  bk_chart=$(cat ../charts/bookkeeper/Chart.yaml | grep name:)
  arr=(`echo ${bk_chart}`)
  bk_chart_name=${arr[1]}
  echo $bk_chart_name

	helm install ../charts/zookeeper-operator --name $zk_opr_name
	kubectl rollout status deploy/$zk_opr_name-zookeeper-operator

	helm install ../charts/zookeeper --name $zk_name

	replicas=$(kubectl get zookeepercluster $zk_name-$zk_chart_name -o jsonpath=$jsonpath1)
	ready=$(kubectl get zookeepercluster $zk_name-$zk_chart_name -o jsonpath=$jsonpath2)

	until [ "$replicas" == "$ready" ]
	do
		sleep 1;
		replicas=$(kubectl get zookeepercluster $zk_name-$zk_chart_name -o jsonpath=$jsonpath1)
		ready=$(kubectl get zookeepercluster $zk_name-$zk_chart_name -o jsonpath=$jsonpath2)
	done

	zk_svc_details=$(kubectl get svc | grep $zk_name- | grep client)
  arr=(`echo ${zk_svc_details}`)
  zk_svc=${arr[0]}
  zk_port_name=${arr[4]}
  IFS="/" read -a arr <<< "$zk_port_name"
  zk_port=${arr[0]}

	helm install ../charts/bookkeeper-operator --name $bk_opr_name
	kubectl rollout status deploy/$bk_opr_name-bookkeeper-operator

	helm install ../charts/bookkeeper --name $bk_name --set zookeeperUri=$zk_svc:$zk_port

	replicas=$(kubectl get bookkeepercluster $bk_name-$bk_chart_name -o jsonpath=$jsonpath1)
	ready=$(kubectl get bookkeepercluster $bk_name-$bk_chart_name -o jsonpath=$jsonpath2)

	until [ "$replicas" == "$ready" ]
	do
		sleep 1;
		replicas=$(kubectl get bookkeepercluster $bk_name-$bk_chart_name -o jsonpath=$jsonpath1)
		ready=$(kubectl get bookkeepercluster $bk_name-$bk_chart_name -o jsonpath=$jsonpath2)
	done

  bk_svc_details=$(kubectl get svc | grep $bk_name- | grep headless)
  arr=(`echo ${bk_svc_details}`)
  bk_svc=${arr[0]}
  bk_port_name=${arr[4]}
  IFS="/" read -a arr <<< "$bk_port_name"
  bk_port=${arr[0]}

	helm install ../charts/pravega-operator --name $pr_opr_name
	kubectl rollout status deploy/$pr_opr_name-pravega-operator

	helm install ../charts/pravega --name $pr_name --set zookeeperUri=$zk_svc:$zk_port --set bookkeeperUri=$bk_svc:$bk_port

elif [ $1 == "delete" ]; then
	helm del $pr_name --purge
	helm del $bk_name --purge
	helm del $zk_name --purge
	helm del $pr_opr_name --purge
	helm del $bk_opr_name --purge
	helm del $zk_opr_name --purge
else
	echo "Error: Invalid argument"
  echo "Use [install] to install the cluster or [delete] to remove the existing setup"
	exit 1
fi
