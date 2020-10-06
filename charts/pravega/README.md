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

To install the pravega chart, use the following commands:

```
$ helm repo add pravega https://charts.pravega.io
$ helm repo update
$ helm install [RELEASE_NAME] pravega/pravega --version=[VERSION] --set zookeeperUri=[ZOOKEEPER_HOST] --set bookkeeperUri=[BOOKKEEPER_SVC] --set storage.longtermStorage.filesystem.pvc=[TIER2_NAME]
```
where:
- **[RELEASE_NAME]** is the release name for the pravega chart.
- **[CLUSTER_NAME]** is the name of the pravega cluster so created. (If [RELEASE_NAME] contains the string `pravega`, `[CLUSTER_NAME] = [RELEASE_NAME]`, else `[CLUSTER_NAME] = [RELEASE_NAME]-pravega`. The [CLUSTER_NAME] can however be overridden by providing `--set fullnameOverride=[CLUSTER_NAME]` along with the helm install command)
  **Note:** You need to ensure that the [CLUSTER_NAME] is the same value as that provided in the [bookkeeper chart configuration](https://github.com/pravega/bookkeeper-operator/tree/master/charts/bookkeeper#configuration), the default value for which is `pravega` and can be achieved by either providing the `[RELEASE_NAME] = pravega` or by providing `--set fullnameOverride=pravega` at the time of installing the pravega chart. On the contrary, the default value of [CLUSTER_NAME] in the bookkeeper charts can also be overridden by providing `--set pravegaClusterName=[CLUSTER_NAME]` at the time of installing the bookkeeper chart)
- **[VERSION]** can be any stable release version for pravega from 0.5.0 onwards.
- **[ZOOKEEPER_HOST]** is the host or IP address of your Zookeeper deployment (e.g. `zookeeper-client:2181`). Multiple Zookeeper URIs can be specified, use a comma-separated list and DO NOT leave any spaces in between (e.g. `zookeeper-0:2181,zookeeper-1:2181,zookeeper-2:2181`).
- **[BOOKKEEPER_SVC]** is the is the name of the headless service of your Bookkeeper deployment (e.g. `bookkeeper-bookie-0.bookkeeper-bookie-headless.default.svc.cluster.local:3181,bookkeeper-bookie-1.bookkeeper-bookie-headless.default.svc.cluster.local:3181,bookkeeper-bookie-2.bookkeeper-bookie-headless.default.svc.cluster.local:3181`).
- **[TIER2_NAME]** is the longtermStorage `PersistentVolumeClaim` name (`pravega-tier2` if you created the PVC using the manifest provided).

This command deploys pravega on the Kubernetes cluster in its default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

>Note: If the underlying pravega operator version is 0.4.5, bookkeeperUri should not be set, and the pravega-bk chart should be used instead of the pravega chart

>Note: If the operator version is 0.5.1 or below and pravega version is 0.9.0 or above, need to set the controller and segment store Jvm options as shown below.
```
helm install [RELEASE_NAME] pravega/pravega --version=[VERSION] --set zookeeperUri=[ZOOKEEPER_HOST] --set bookkeeperUri=[BOOKKEEPER_SVC] --set storage.longtermStorage.filesystem.pvc=[TIER2_NAME] --set controller.jvmOptions[0]="-XX:+UseContainerSupport" --set controller.jvmOptions[1]="-XX:+IgnoreUnrecognizedVMOptions" --set segmentStore.jvmOptions[0]="-XX:+UseContainerSupport" --set segmentStore.jvmOptions[1]="-XX:+UseContainerSupport" --set segmentStore.jvmOptions[2]="-Xmx2g" --set segmentStore.jvmOptions[3]="-XX:MaxDirectMemorySize=2g"
```
Based on the cluster flavours selected ,segementstore memory requirements needs to be adjusted.


## Uninstalling the Chart

To uninstall/delete the pravega chart, use the following command:

```
$ helm uninstall [RELEASE_NAME]
```

This command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the pravega chart and their default values.

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
| `segmentStore.service.loadBalancerIP` |It is used to provide a LoadBalancerIP for the segmentStore service | |
| `segmentStore.service.externalTrafficPolicy` | It is used to provide ExternalTrafficPolicy for the segmentStore service |  |
| `segmentStore.jvmOptions` | JVM Options for segmentStore | `[]` |
| `storage.longtermStorage.type` | Type of long term storage backend to be used (filesystem/ecs/hdfs) | `filesystem` |
| `storage.longtermStorage.filesystem.pvc` | Name of the pre-created PVC, if long term storage type is filesystem | `pravega-tier2` |
| `storage.longtermStorage.ecs` | Configuration to use a Dell EMC ECS system, if long term storage type is ecs | `{}` |
| `storage.longtermStorage.hdfs` | Configuration to use an HDFS system, if long term storage type is hdfs | `{}` |
| `storage.cache.className` | Storage class for cache volume | `` |
| `storage.cache.size` | Storage requests for cache volume | `20Gi` |
| `options` | List of Pravega options | |
