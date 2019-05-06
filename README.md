# Pravega Operator

 [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/pravega/pravega-operator?status.svg)](https://godoc.org/github.com/pravega/pravega-operator) [![Build Status](https://travis-ci.org/pravega/pravega-operator.svg?branch=master)](https://travis-ci.org/pravega/pravega-operator) [![Go Report](https://goreportcard.com/badge/github.com/pravega/pravega-operator)](https://goreportcard.com/report/github.com/pravega/pravega-operator)

### Project status: alpha

The project is currently alpha. While no breaking API changes are currently planned, we reserve the right to address bugs and change the API before the project is declared stable.

## Table of Contents

 * [Overview](#overview)
 * [Requirements](#requirements)
 * [Quickstart](#quickstart)    
    * [Installation of the Operator](#install-the-operator)
    * [Install a sample Pravega Cluster](#install-a-sample-pravega-cluster)
    * [Scale a Pravega Cluster](#scale-a-pravega-cluster)
    * [Upgrade a Pravega Cluster](#upgrade-a-pravega-cluster)
    * [Uninstall the Pravega Cluster](#uninstall-the-pravega-cluster)
    * [Uninstall the Operator](#uninstall-the-operator)
 * [Configuration](#configuration)
 * [Development](#development)
* [Releases](#releases)
* [Troubleshooting](#troubleshooting)

## Overview

[Pravega](http://pravega.io) is an open source distributed storage service implementing Streams. It offers Stream as the main primitive for the foundation of reliable storage systems: *a high-performance, durable, elastic, and unlimited append-only byte stream with strict ordering and consistency*.

The Pravega Operator manages Pravega clusters deployed to Kubernetes and automates tasks related to operating a Pravega cluster.

- [x] Create and destroy a Pravega cluster
- [x] Resize cluster
- [x] Rolling upgrades (experimental)

> Note that unchecked features are in the roadmap but not available yet.

## Requirements

- Kubernetes 1.8+
- An existing Apache Zookeeper 3.5 cluster. This can be easily deployed using our [Zookeeper operator](https://github.com/pravega/zookeeper-operator)

## Quickstart

### Install the Operator

> Note: If you are running on Google Kubernetes Engine (GKE), please [check this first](doc/development.md#installation-on-google-kubernetes-engine).

Use Helm to quickly deploy a Pravega operator with the release name `foo`

```
$ helm install charts/pravega-operator --name foo
```

Verify that the Pravega Operator is running.

```
$ kubectl get deploy
NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
foo-pravega-operator     1         1         1            1           17s
```
You can also manually install the operator, check [here](doc/development.md#install-the-operator-manually) for instructions.
### Install a sample Pravega cluster
#### Install Tier2 Storage 
Pravega requires a long term storage provider known as Tier 2 storage.

See [instructions](doc/tier2.md#use-NFS-as-Tier2) on how to install a NFS volume as Tier 2

#### Install Pravega cluster
Use Helm to install a sample Pravega cluster with release name `bar`
```
$ helm install charts/pravega --name bar --set zookeeperUri=[ZOOKEEPER_HOST] --set pravega.tier2=[TIER2_NAME]
```

where:

- `[ZOOKEEPER_HOST]` is the host or IP address of your Zookeeper deployment. E.g. `zk-client:2181`
> Note: If multiple zk URIs are specified, use a comma separated list and DO NOT leave any spaces in between.
- `[Tier2_NAME]` is the Tier 2 PersistentVolumeClaim name. E.g. `pravega-tier2`


Verify that the cluster instances and its components are being created.

```
$ kubectl get PravegaCluster
NAME          VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
bar-pravega   0.4.0     7                 0               25s
```

After a couple of minutes, all cluster members should become ready.

```
$ kubectl get PravegaCluster
NAME         VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
bar-pravega  0.4.0     7                 7               2m
```

```
$ kubectl get all -l pravega_cluster=bar-pravega
NAME                                              READY   STATUS    RESTARTS   AGE
pod/bar-bookie-0                              1/1     Running   0          2m
pod/bar-bookie-1                              1/1     Running   0          2m
pod/bar-bookie-2                              1/1     Running   0          2m
pod/bar-pravega-controller-64ff87fc49-kqp9k   1/1     Running   0          2m
pod/bar-pravega-segmentstore-0                1/1     Running   0          2m
pod/bar-pravega-segmentstore-1                1/1     Running   0          1m
pod/bar-pravega-segmentstore-2                1/1     Running   0          30s

NAME                                            TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)              AGE
service/bar-bookie-headless                 ClusterIP   None          <none>        3181/TCP             2m
service/bar-pravega-controller              ClusterIP   10.23.244.3   <none>        10080/TCP,9090/TCP   2m
service/bar-pravega-segmentstore-headless   ClusterIP   None          <none>        12345/TCP            2m

NAME                                                    DESIRED   CURRENT   READY   AGE
replicaset.apps/bar-pravega-controller-64ff87fc49   1         1         1       2m

NAME                                            DESIRED   CURRENT   AGE
statefulset.apps/bar-bookie                 3         3         2m
statefulset.apps/bar-pravega-segmentstore   3         3         2m
```

By default, a `PravegaCluster` instance is only accessible within the cluster through the Controller `ClusterIP` service. From within the Kubernetes cluster, a client can connect to Pravega at:

```
tcp://<pravega-name>-pravega-controller.<namespace>:9090
```

And the `REST` management interface is available at:

```
http://<pravega-name>-pravega-controller.<namespace>:10080/
```

Check [here](doc/external-access.md) to enable external access to a Pravega cluster.

### Scale a Pravega cluster

You can scale Pravega components independently by modifying their corresponding field in the Pravega resource spec. You can either `kubectl edit` the cluster or `kubectl patch` it. If you edit it, update the number of replicas for BookKeeper, Controller, and/or Segment Store and save the updated spec.

Example of patching the Pravega resource to scale the Segment Store instances to 4.

```
kubectl patch PravegaCluster <pravega-name> --type='json' -p='[{"op": "replace", "path": "/spec/pravega/segmentStoreReplicas", "value": 4}]'
```

### Upgrade a Pravega cluster

Check out the [upgrade guide](doc/upgrade-cluster.md).

### Uninstall the Pravega cluster

```
$ helm delete bar
$ kubectl delete -f pvc.yaml
```

### Uninstall the Operator

> Note that the Pravega clusters managed by the Pravega operator will NOT be deleted even if the operator is uninstalled.

To delete all clusters, delete all cluster CR objects before uninstalling the operator.

```
$ helm delete foo
```

## Configuration
Check out the  [configuration document](doc/configuration.md).


## Development
Check out the  [development guide](doc/development.md).


## Releases  

The latest Pravega releases can be found on the [Github Release](https://github.com/pravega/pravega-operator/releases) project page.

## Troubleshooting

Check out the [troubleshooting document](doc/troubleshooting.md).
