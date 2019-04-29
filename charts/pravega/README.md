# Pravega Helm Chart

Installs [Pravega](https://github.com/pravega/pravega) clusters atop Kubernetes.

## Introduction

This chart creates a [Pravega](https://github.com/pravega/pravega) cluster in [Kubernetes](http://kubernetes.io) using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Pravega cluster on multiple namespaces.

## Prerequisites

  - Kubernetes 1.10+ with Beta APIs
  - Helm 2.10+
  - Pravega Operator. You can install it using its own [Helm chart](https://github.com/pravega/pravega-operator/tree/master/charts/pravega-operator)

## Installing the Chart

To install the chart with the release name `my-release`:

```
$ helm install --name my-release pravega
```

The command deploys pravega on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Pravega chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `version` | Version for Pravega cluster | `0.5.0` |
| `zookeeperUri` | Zookeeper service address | `zk-client:2181` |
| `externalAccess.enabled` | Enable Pravega external access | `false` |
| `externalAccess.type` | Pravega external access type | `LoadBalancer` |
| `bookkeeper.image.repository` | Image repo for Bookkeeper image | `pravega/bookkeeper` |
| `bookkeeper.replicas` | Replicas for Bookkeeper | `3` |
| `bookkeeper.storage.ledgerVolumeRequest` | Request storage for ledgerVolume | `10Gi` |
| `bookkeeper.storage.journalVolumeRequest` | Request storage for journalVolume | `10Gi` |
| `bookkeeper.storage.indexVolumeRequest` | Request storage for indexVolume | `10Gi` |
| `bookkeeper.autoRecovery`| Enable Bookkeeper autoRecovery | `true` |
| `pravega.image.repository` | Image repo for Pravega image | `pravega/pravega` |
| `pravega.controllerReplicas` | Replicas for controller | `1` |
| `pravega.segmentStoreReplicas` | Replicas for segmentStore | `1` |
| `pravega.debugLogging` | Enable debug logging | `false` |
| `pravega.cacheVolumeRequest` | Request storage for cacheVolume | `20Gi` |
| `pravega.tier2` | Name of the PVC used for Tier 2 storage | `pravega-tier2` |
