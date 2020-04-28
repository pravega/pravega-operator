#!/usr/bin/env bash
if [ "$#" -ne 1 ]; then
	echo "Error : Invalid number of arguments"
	echo "Usage: ./pravegacluster.sh [install/delete]"
	exit 1
fi

jsonpath1="{.spec.replicas}"
jsonpath2="{.status.readyReplicas}"
zk_opr_name=zook-opr
zk_name=zook
bk_opr_name=book-opr
bk_name=book
pr_opr_name=pr-opr
pr_name=pr

install_cluster () {
	pr_chart=$(cat ../charts/pravega/Chart.yaml | grep name:)
	arr=(`echo ${pr_chart}`)
	pr_chart_name=${arr[1]}

	zk_chart=$(cat ../charts/zookeeper/Chart.yaml | grep name:)
	arr=(`echo ${zk_chart}`)
	zk_chart_name=${arr[1]}

	bk_chart=$(cat ../charts/bookkeeper/Chart.yaml | grep name:)
	arr=(`echo ${bk_chart}`)
	bk_chart_name=${arr[1]}
	
	# Installing the Zookeeper Operator	
	if ["$helmv3" == ""]
	then	
		helm install ../charts/zookeeper-operator --name $zk_opr_name
	else
		helm install $zk_opr_name ../charts/zookeeper-operator		
	fi

	kubectl rollout status deploy/$zk_opr_name-zookeeper-operator

	# Installing the Zookeeper Cluster
	if ["$helmv3" == ""]
	then
		helm install ../charts/zookeeper --name $zk_name
	else
		helm install  $zk_name ../charts/zookeeper
	fi		

	# Waiting for Zookeeper Cluster to be ready
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

	# Installing the BookKeeper Operator
	if ["$helmv3" == ""]
	then
		helm install ../charts/bookkeeper-operator --name $bk_opr_name
	else
		helm install  $bk_opr_name ../charts/bookkeeper-operator 
	fi
		
	kubectl rollout status deploy/$bk_opr_name-bookkeeper-operator

	# Altering the values for the fields PRAVEGA_CLUSTER_NAME and WAIT_FOR inside the bookkeeper configmap
	sed -i "/PRAVEGA_CLUSTER_NAME/c \ \ PRAVEGA_CLUSTER_NAME: $pr_name-$pr_chart_name" ../charts/bookkeeper/templates/config_map.yaml
	sed -i "/WAIT_FOR/c \ \ WAIT_FOR: $zk_svc:$zk_port" ../charts/bookkeeper/templates/config_map.yaml

	# Installing the BookKeeper Cluster
	if ["$helmv3" == ""]
	then
		helm install ../charts/bookkeeper --name $bk_name --set zookeeperUri=$zk_svc:$zk_port
	else
		helm install $bk_name ../charts/bookkeeper  --set zookeeperUri=$zk_svc:$zk_port
	fi		

	# Waiting for Bookkeeper Cluster to be ready
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

	# Installing the Pravega Operator
	if ["$helmv3" == ""]
	then
		helm install ../charts/pravega-operator --name $pr_opr_name
	else
		helm install  $pr_opr_name ../charts/pravega-operator
	fi	
			
	kubectl rollout status deploy/$pr_opr_name-pravega-operator

	# Installing the Pravega Cluster
	if ["$helmv3" == ""]
	then
		helm install ../charts/pravega --name $pr_name --set zookeeperUri=$zk_svc:$zk_port --set bookkeeperUri=$bk_svc:$bk_port
	else
		helm install  $pr_name ../charts/pravega --set zookeeperUri=$zk_svc:$zk_port --set bookkeeperUri=$bk_svc:$bk_port
	fi	

}

delete_cluster_v2() {
	helm del $pr_name --purge
	helm del $bk_name --purge
	helm del $zk_name --purge
	helm del $pr_opr_name --purge
	helm del $bk_opr_name --purge
	helm del $zk_opr_name --purge
}

delete_cluster_v3(){
	helm uninstall $pr_name 
	helm uninstall $bk_name 
	helm uninstall $zk_name 
	helm uninstall $pr_opr_name 
	helm uninstall $bk_opr_name 
	helm uninstall $zk_opr_name 
}

# Variable used for checking presence of helm3
helmv3=$(helm version |grep v3 | head -1)
	
if [ $1 == "install" ]; then
	install_cluster

elif [ $1 == "delete" ]; 
	then
		if ["$helmv3" == ""]
		then
			delete_cluster_v2
		else
			delete_cluster_v3
		fi	
else
	echo "Error: Invalid argument"
	echo "Use [install] to install the cluster or [delete] to remove the existing setup"
	exit 1
fi
