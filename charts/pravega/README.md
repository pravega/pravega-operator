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
| `externalAccessEnabled` | Enable Pravega external access | `false` |
| `bookkeeper.image.repository` | Image repo for Bookkeeper image | `pravega/bookkeeper` |
| `bookkeeper.replicas` | Replicas for Bookkeeper | `3` |
| `bookkeeper.storage.ledgerVolumeRequest` | Request storage for ledgerVolume | `10Gi` |
| `bookkeeper.storage.journalVolumeRequest` | Request storage for journalVolume | `10Gi` |
| `bookkeeper.storage.indexVolumeRequest` | Request storage for indexVolume | `10Gi` |
| `bookkeeper.autoRecovery`| Enable Bookkeeper autoRecovery | `true` |
| `pravega.image.repository` | Image repo for Pravega image | `pravega/pravega` |
| `pravega.controller.replicas` | Replicas for controller | `1` |
| `pravega.controller.resources` | Resources for controller | `requests.cpu="250m",requests.memory="512Mi",limits.cpu="500m",limits.memory="1Gi" ` |
| `pravega.controller.debugLogging` | Enable debug logging on controller | `false` |
| `pravega.controller.externalAccess.type` | Type of controller service when externalAccess is enabled | `LoadBalancer` |
| `pravega.controller.options` | Pravega Controller Options|  |
| `pravega.controller.jvmOptions` | Pravega Controller JVMOptions|  |
| `pravega.segmentstore.replicas` | Replicas for segmentstore | `1` |
| `pravega.segmentstore.resources` | Resources for controller | `requests.cpu="500m",requests.memory="1Gi",limits.cpu="1",limits.memory="2Gi"` |
| `pravega.segmentstore.debugLogging` | Enable debug logging on segment store| `false` |
| `pravega.segmentstore.externalAccess.type` | Type of segmentstore service | `LoadBalancer` |
| `pravega.segmentstore.options` | Pravega segmentstore options |  |
| `pravega.controller.jvmOptions` | Pravega Controller JVMOptions|  |
| `pravega.segmentstore.cacheVolumeRequest` | Request storage for cacheVolume | `20Gi` |
| `pravega.tier2` | Name of the PVC used for Tier 2 storage | `pravega-tier2` |
