# Pravega Helm Chart

Installs [Pravega](https://github.com/pravega/pravega) clusters atop Kubernetes.

## Introduction

This chart creates a [Pravega](https://github.com/pravega/pravega) cluster in [Kubernetes](http://kubernetes.io) using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Pravega cluster on multiple namespaces.

## Prerequisites

  - Kubernetes 1.15+ with Beta APIs
  - Helm 3.2.1+
  - An existing Apache Zookeeper 3.6.1 cluster. This can be easily deployed using our [Zookeeper Operator](https://github.com/pravega/zookeeper-operator)
  - An existing Apache Bookkeeper 4.9.2 cluster. This can be easily deployed using our [BookKeeper Operator](https://github.com/pravega/bookkeeper-operator)
  - Pravega Operator. You can install it using its own [Helm chart](https://github.com/pravega/pravega-operator/tree/master/charts/pravega-operator)

## Installing the Chart

To install the chart with the release name `my-release`:

```
$ helm install my-release pravega
```

The command deploys pravega on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```
$ helm uninstall my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Pravega chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `version` | Pravega version | `0.7.0` |
| `tls` | Pravega security configuration passed to the Pravega processes | `{}` |
| `authentication.enabled` | Enable authentication to authorize client communication with Pravega | `false` |
| `authentication.passwordAuthSecret` | Name of Secret containing Password based Authentication Parameters, if authentication is enabled | |
| `zookeeperUri` | Zookeeper client service URI | `zookeeper-client:2181` |
| `bookkeeperUri` | Bookkeeper headless service URI | `bookkeeper-bookie-headless:3181` |
| `externalAccess.enabled` | Enable external access | `false` |
| `externalAccess.type` | External access service type, if external access is enabled (LoadBalancer/NodePort) | `LoadBalancer` |
| `externalAccess.domainName` | External access domain name, if external access is enabled  | |
| `image.repository` | Image repository | `pravega/pravega` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `debugLogging` | Enable debug logging | `false` |
| `serviceAccount.name` | Service account to be used | `pravega-components` |
| `controller.replicas` | Number of controller replicas | `1` |
| `controller.resources.requests.cpu` | CPU requests for controller | `500m` |
| `controller.resources.requests.memory` | Memory requests for controller | `1Gi` |
| `controller.resources.limits.cpu` | CPU limits for controller | `1000m` |
| `controller.resources.limits.memory` | Memory limits for controller | `2Gi` |
| `controller.service.type` | Override the controller service type, if external access is enabled (LoadBalancer/NodePort) | |
| `controller.service.annotations` | Annotations to add to the controller service, if external access is enabled | `{}` |
| `controller.jvmOptions` | JVM Options for controller | `["-Xmx2g", "-XX:MaxDirectMemorySize=2g"]` |
| `segmentStore.replicas` | Number of segmentStore replicas | `1` |
| `segmentStore.secret` | Secret configuration for the segmentStore | `{}` |
| `segmentStore.env` | Name of configmap containing environment variables to be added to the segmentStore | |
| `segmentStore.resources.requests.cpu` | CPU requests for segmentStore | `1000m` |
| `segmentStore.resources.requests.memory` | Memory requests for segmentStore | `4Gi` |
| `segmentStore.resources.limits.cpu` | CPU limits for segmentStore | `2000m` |
| `segmentStore.resources.limits.memory` | Memory limits for segmentStore | `4Gi` |
| `segmentStore.service.type` | Override the segmentStore service type, if external access is enabled (LoadBalancer/NodePort) | |
| `segmentStore.service.annotations` | Annotations to add to the segmentStore service, if external access is enabled | `{}` |
| `segmentStore.service.segmentStoreLoadBalancerIP` |It is used to provide a LoadBalancerIP | |
| `segmentStore.service.segmentStoreExternalTrafficPolicy` | It is used to provide segmentStoreExternalTrafficPolicy  |  |  
| `segmentStore.jvmOptions` | JVM Options for segmentStore | `[]` |
| `storage.longtermStorage.type` | Type of long term storage backend to be used (filesystem/ecs/hdfs) | `filesystem` |
| `storage.longtermStorage.filesystem.pvc` | Name of the pre-created PVC, if long term storage type is filesystem | `pravega-tier2` |
| `storage.longtermStorage.ecs` | Configuration to use a Dell EMC ECS system, if long term storage type is ecs | `{}` |
| `storage.longtermStorage.hdfs` | Configuration to use an HDFS system, if long term storage type is hdfs | `{}` |
| `storage.cache.className` | Storage class for cache volume | `standard` |
| `storage.cache.size` | Storage requests for cache volume | `20Gi` |
| `options` | List of Pravega options | |
