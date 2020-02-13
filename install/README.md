# Script for complete Pravega Installation

The purpose of this script is to sequentially install all the dependencies (i.e. Operators, Zookeeper and Bookkeeper) and Pravega Components in the right order.

## Prerequisites

  - Kubernetes 1.10+ with Beta APIs
  - Helm 2.10+
  - [Tier 2 Setup](https://github.com/pravega/pravega-operator#set-up-tier-2-storage)
  - Copy the necessary charts to the right location

To copy the charts from [Zookeeper Operator](https://github.com/pravega/zookeeper-operator/tree/master/charts) and [Bookkeeper Operator](https://github.com/pravega/bookkeeper-operator/tree/master/charts) repositories inside the charts directory of this repository.

Create separate sub-directories for [zookeeper-operator](https://github.com/pravega/zookeeper-operator/tree/master/charts/zookeeper-operator), [zookeeper](https://github.com/pravega/zookeeper-operator/tree/master/charts/zookeeper), [bookkeeper-operator](https://github.com/pravega/bookkeeper-operator/tree/master/charts/pravega-operator) and [bookkeeper](https://github.com/pravega/bookkeeper-operator/tree/master/charts/pravega) charts alongside the directories for [pravega-operator](https://github.com/pravega/pravega-operator/tree/master/charts/pravega-operator) and [pravega](https://github.com/pravega/pravega-operator/tree/master/charts/pravega) charts inside the [charts](https://github.com/pravega/pravega-operator/tree/master/charts) directory.

## Installing the Pravega Cluster

To install the Pravega Cluster along with all the required dependencies, run the following command:

```
$ ./install.sh 1
```

## Uninstalling the Pravega Cluster

To uninstall the Pravega Cluster along with all its dependencies, run the following command:

```
$ ./install.sh 2
```
