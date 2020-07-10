# Script for complete Pravega Installation

The purpose of this script is to sequentially deploy all the dependencies (i.e. Operators, Zookeeper and Bookkeeper) and Pravega Components in the right order.

## Prerequisites

  - Kubernetes 1.15+ with Beta APIs
  - Helm 3+
  - [Tier 2 Setup](https://github.com/pravega/pravega-operator#set-up-tier-2-storage)
  - Cert-Manager v0.15.0+
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

We use cert-manager for certificate management for webhook services in Kubernetes. In case you plan to use the same, you would need to [install cert-manager](https://cert-manager.io/docs/installation/kubernetes/)

## Deploying the Pravega Cluster

To deploy the Pravega Cluster along with all the required dependencies, run the following command:

```
$ ./pravegacluster.sh deploy
```

## Undeploying the Pravega Cluster

To remove the Pravega Cluster along with all its dependencies, run the following command:

```
$ ./pravegacluster.sh destroy
```
