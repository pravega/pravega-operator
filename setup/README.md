# Script for complete Pravega Installation

The purpose of this script is to sequentially install all the dependencies (i.e. Operators, Zookeeper and Bookkeeper) and Pravega Components in the right order.

## Prerequisites

  - Kubernetes 1.10+ with Beta APIs
  - Helm 2.10+
  - [Tier 2 Setup](https://github.com/pravega/pravega-operator#set-up-tier-2-storage)
  - Copy the necessary charts to the right location

First clone the [Zookeeper Operator](https://github.com/pravega/zookeeper-operator) and [Bookkeeper Operator](https://github.com/pravega/bookkeeper-operator) repositories locally using :
```
git clone https://github.com/pravega/zookeeper-operator
git clone https://github.com/pravega/bookkeeper-operator
```

Next, copy the contents of the charts directory from both these repositories inside the charts directory of this repository.
```
cp -r <path-to-zookeeper-operator-repo>/charts/ <path-to-pravega-operator-repo>/charts/.
cp -r <path-to-bookkeeper-operator-repo>/charts/ <path-to-pravega-operator-repo>/charts/.
```

This will result in separate sub-directories for [zookeeper-operator](https://github.com/pravega/zookeeper-operator/tree/master/charts/zookeeper-operator), [zookeeper](https://github.com/pravega/zookeeper-operator/tree/master/charts/zookeeper), [bookkeeper-operator](https://github.com/pravega/bookkeeper-operator/tree/master/charts/pravega-operator) and [bookkeeper](https://github.com/pravega/bookkeeper-operator/tree/master/charts/pravega) charts alongside the directories for [pravega-operator](https://github.com/pravega/pravega-operator/tree/master/charts/pravega-operator) and [pravega](https://github.com/pravega/pravega-operator/tree/master/charts/pravega) charts inside the [charts](https://github.com/pravega/pravega-operator/tree/master/charts) directory.

## Installing the Pravega Cluster

To install the Pravega Cluster along with all the required dependencies, run the following command:

in case of helm3+ run the following command
```
$ ./pravegacluster_helm3.sh install
```

otherwise run the fllowing command
```
$ ./pravegacluster.sh install
```


## Uninstalling the Pravega Cluster

To uninstall the Pravega Cluster along with all its dependencies, run the following command:

in case of helm3+ run the following command
```
$ ./pravegacluster_helm3.sh  delete
```
otherwise run the fllowing command
```
$ ./pravegacluster.sh delete
```
