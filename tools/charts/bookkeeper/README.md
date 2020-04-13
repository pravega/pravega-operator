# Bookkeeper Helm Chart

Installs Bookkeeper clusters atop Kubernetes.

## Introduction

This chart creates a Bookkeeper cluster in [Kubernetes](http://kubernetes.io) using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Bookkeeper cluster on multiple namespaces.

## Prerequisites

  - Kubernetes 1.10+ with Beta APIs
  - Helm 2.10+
  - An existing Apache Zookeeper 3.5 cluster. This can be easily deployed using our [Zookeeper operator](https://github.com/pravega/zookeeper-operator
  - Bookkeeper Operator. You can install it using its own [Helm chart](https://github.com/pravega/pravega-operator/tree/master/charts/pravega-operator)

## Installing the Chart

To install the chart with the release name `my-release`:

```
$ helm install --name my-release bookkeeper
```

The command deploys bookkeeper on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Bookkeeper chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `version` | Version for Bookkeeper cluster | `latest` |
| `zookeeperUri` | Zookeeper service address | `zk-client:2181` |
| `image.repository` | Repository for Bookkeeper image | `pravega/bookkeeper` |
| `image.pullPolicy` | Pull policy for Bookkeeper image | `IfNotPresent` |
| `replicas` | Replicas for Bookkeeper | `3` |
| `autoRecovery`| Enable Bookkeeper autoRecovery | `true` |
