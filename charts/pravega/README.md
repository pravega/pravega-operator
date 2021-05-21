# Deploying Pravega

Here, we briefly describe how to [install](#installing-pravega-cluster)/[update](#updating-pravega-cluster)/[uninstall](#uninstalling-pravega-cluster)/[configure](#pravega-configuration) pravega clusters atop kubernetes.

## Prerequisites

  - Kubernetes 1.15+ with Beta APIs
  - Helm 3.2.1+
  - An existing Apache Zookeeper 3.6.1 cluster. This can be easily deployed using our [Zookeeper Operator](https://github.com/pravega/zookeeper-operator)
  - An existing Apache Bookkeeper 4.9.2 cluster. This can be easily deployed using our [BookKeeper Operator](https://github.com/pravega/bookkeeper-operator)
  - Pravega Operator. Please refer [this](../../charts/pravega-operator/README.md)
  - LongTerm Storage ([options for long term storage](../../doc/longtermstorage.md))

## Installing Pravega Cluster

To install the pravega cluster, use the following commands:

```
$ helm repo add pravega https://charts.pravega.io
$ helm repo update
$ helm install [RELEASE_NAME] pravega/pravega --version=[VERSION] --set zookeeperUri=[ZOOKEEPER_SVC] --set bookkeeperUri=[BOOKKEEPER_SVC] --set storage.longtermStorage.filesystem.pvc=[TIER2_NAME]
```
where:

- **[RELEASE_NAME]** is the release name for the pravega chart.
- **[VERSION]** can be any stable release version for pravega from 0.5.0 onwards.
- **[ZOOKEEPER_SVC]** is the name of client service of your Zookeeper deployment (e.g. `zookeeper-client:2181`). Multiple Zookeeper URIs can be specified, use a comma-separated list and DO NOT leave any spaces in between (e.g. `zookeeper-0:2181,zookeeper-1:2181,zookeeper-2:2181`).
- **[BOOKKEEPER_SVC]** is the name of the headless service of your Bookkeeper deployment (e.g. `bookkeeper-bookie-0.bookkeeper-bookie-headless.default.svc.cluster.local:3181,bookkeeper-bookie-1.bookkeeper-bookie-headless.default.svc.cluster.local:3181,bookkeeper-bookie-2.bookkeeper-bookie-headless.default.svc.cluster.local:3181`).
- **[TIER2_NAME]** is the longtermStorage `PersistentVolumeClaim` name (`pravega-tier2` if you created the PVC using the manifest provided).

>Note:  You need to ensure that the [CLUSTER_NAME] is the same value as that provided in the [bookkeeper chart configuration](https://github.com/pravega/bookkeeper-operator/tree/master/charts/bookkeeper#configuration),the default value for which is `pravega` and can be achieved by either providing the `[RELEASE_NAME] = pravega` or by providing `--set fullnameOverride=pravega` at the time of installing the pravega chart. On the contrary, the default value of [CLUSTER_NAME] in the bookkeeper charts can also be overridden by providing `--set pravegaClusterName=[CLUSTER_NAME]` at the time of installing the bookkeeper chart)

>Note: If we provide [RELEASE_NAME] same as chart name, cluster name will be same as release-name. But if we are providing a different name for release(other than pravega in this case), cluster name will be [RELEASE_NAME]-[chart-name]. However, cluster name can be overridden by providing `--set  fullnameOverride=[CLUSTER_NAME]` along with helm install command.

This command deploys pravega on the Kubernetes cluster in its default configuration. The [configuration](#pravega-configuration) section lists the parameters that can be configured during installation.

>Note: If the underlying pravega operator version is 0.4.5, bookkeeperUri should not be set, and the pravega-bk chart should be used instead of the pravega chart

>Note: If the operator version is 0.5.1 or below and pravega version is 0.9.0 or above, need to set the controller and segment store JVM options as shown below.
```
helm install [RELEASE_NAME] pravega/pravega --version=[VERSION] --set zookeeperUri=[ZOOKEEPER_SVC] --set bookkeeperUri=[BOOKKEEPER_SVC] --set storage.longtermStorage.filesystem.pvc=[TIER2_NAME] --set 'controller.jvmOptions={-XX:+UseContainerSupport,-XX:+IgnoreUnrecognizedVMOptions}' --set 'segmentStore.jvmOptions={-XX:+UseContainerSupport,-XX:+IgnoreUnrecognizedVMOptions,-Xmx2g,-XX:MaxDirectMemorySize=2g}'
```

Based on the [cluster flavours](values) selected,segmentstore memory requirements needs to be adjusted.

Once installation is successful, all cluster members should become ready.
```
$ kubectl get all -l pravega_cluster=bar-pravega
NAME                                          READY   STATUS    RESTARTS   AGE
pod/bar-pravega-controller-64ff87fc49-kqp9k   1/1     Running   0          2m
pod/bar-pravega-segmentstore-0                1/1     Running   0          2m
pod/bar-pravega-segmentstore-1                1/1     Running   0          1m
pod/bar-pravega-segmentstore-2                1/1     Running   0          30s

NAME                                        TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)              AGE
service/bar-pravega-controller              ClusterIP   10.23.244.3   <none>        10080/TCP,9090/TCP   2m
service/bar-pravega-segmentstore-headless   ClusterIP   None          <none>        12345/TCP            2m

NAME                                                    DESIRED   CURRENT   READY   AGE
replicaset.apps/bar-pravega-controller-64ff87fc49       1         1         1       2m

NAME                                            DESIRED   CURRENT   AGE
statefulset.apps/bar-pravega-segmentstore       3         3         2m
```

By default, a `PravegaCluster` instance is only accessible within the cluster through the Controller `ClusterIP` service. From within the Kubernetes cluster, a client can connect to Pravega at:

```
tcp://[CLUSTER_NAME]-pravega-controller.[NAMESPACE]:9090
```

And the `REST` management interface is available at:

```
http://[CLUSTER_NAME]-pravega-controller.[NAMESPACE]:10080/
```

Check out the [external access documentation](../../doc/external-access.md) if your clients need to connect to Pravega from outside Kubernetes.

Check out the [exposing Segmentstore service on single IP address](../..//doc/external-access.md#exposing-segmentstore-service-on-single-ip-address-and-different-ports) if your clients need to connect to all Pravega Segmentstore instances on the same IP address from outside Kubernetes.

## Updating Pravega Cluster

For updating the pravega cluster, use the following command

```
helm upgrade [RELEASE_NAME]  --version=[VERSION]  --set controller.replicas=2 --set segmentstore.replicas=3
```
 we can also update other configurable parameters at run time. For changing log level to DEBUG use the below command.

 ```
  helm upgrade pravega charts/pravega --set-string options."log\.level"="DEBUG"
  ```
Please refer [upgrade](../../doc/upgrade-cluster.md) for upgrading cluster versions.

## Uninstalling Pravega Cluster

To uninstall/delete the pravega chart, use the following command:

```
$ helm uninstall [RELEASE_NAME]
```

This command removes all the Kubernetes components associated with the chart and deletes the release.

## Pravega Configuration

The following table lists the configurable parameters of the pravega chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `version` | Pravega version | `0.9.0` |
| `tls` | Pravega security configuration passed to the Pravega processes | `{}` |
| `authentication.enabled` | Enable authentication to authorize client communication with Pravega | `false` |
| `authentication.passwordAuthSecret` | Name of Secret containing Password based Authentication Parameters, if authentication is enabled | |
| `authentication.segmentStoreTokenSecret` | Name of Secret containing tokenSigningkey for the ss, if authentication is enabled | |
| `authentication.controllerTokenSecret` | Name of Secret containing tokenSigningkey for controller, if authentication is enabled | |
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
| `controller.maxUnavailableReplicas` | Maximum number of unavailable replicas possible for controller pdb | |
| `controller.resources.requests.cpu` | CPU requests for controller | `500m` |
| `controller.resources.requests.memory` | Memory requests for controller | `1Gi` |
| `controller.resources.limits.cpu` | CPU limits for controller | `1000m` |
| `controller.resources.limits.memory` | Memory limits for controller | `2Gi` |
| `controller.securityContext` | Holds pod-level security attributes and common container settings for controller | `{}` |
| `controller.affinity` | Specifies scheduling constraints on controller pods | `{}` |
| `controller.service.type` | Override the controller service type, if external access is enabled (LoadBalancer/NodePort) | |
| `controller.service.annotations` | Annotations to add to the controller service, if external access is enabled | `{}` |
| `controller.labels` | Labels to add to the controller pods | `{}` |
| `controller.annotations` | Annotations to add to the controller pods | `{}` |
| `controller.jvmOptions` | JVM Options for controller | `["-Xms512m", "-XX:+ExitOnOutOfMemoryError", "-XX:+CrashOnOutOfMemoryError", "-XX:+HeapDumpOnOutOfMemoryError", "-XX:HeapDumpPath=/tmp/dumpfile/heap", "-XX:MaxRAMPercentage=50.0", "-XX:+UseContainerSupport", "-XX:+PrintExtendedThreadInfo"]` |
| `controller.svcNameSuffix` | suffix for controller service name | `pravega-controller` |
| `controller.initContainers` | Init Conatiners to add to controller pods | `[]` |
| `segmentStore.replicas` | Number of segmentStore replicas | `1` |
| `segmentStore.maxUnavailableReplicas` | Maximum number of unavailable replicas possible for segmentStore pdb | |
| `segmentStore.secret` | Secret configuration for the segmentStore | `{}` |
| `segmentStore.env` | Name of configmap containing environment variables to be added to the segmentStore | |
| `segmentStore.resources.requests.cpu` | CPU requests for segmentStore | `1000m` |
| `segmentStore.resources.requests.memory` | Memory requests for segmentStore | `4Gi` |
| `segmentStore.resources.limits.cpu` | CPU limits for segmentStore | `2000m` |
| `segmentStore.resources.limits.memory` | Memory limits for segmentStore | `4Gi` |
| `segmentStore.securityContext` | Holds pod-level security attributes and common container settings for segmentStore | `{}` |
| `segmentStore.affinity | Specifies scheduling constraints on segmentStore pods | `{}` |
| `segmentStore.service.type` | Override the segmentStore service type, if external access is enabled (LoadBalancer/NodePort) | |
| `segmentStore.service.annotations` | Annotations to add to the segmentStore service, if external access is enabled | `{}` |
| `segmentStore.service.loadBalancerIP` | It is used to provide a LoadBalancerIP for the segmentStore service | |
| `segmentStore.service.externalTrafficPolicy` | It is used to provide ExternalTrafficPolicy for the segmentStore service |  |
| `segmentStore.jvmOptions` | JVM Options for segmentStore | `["-Xms1g", "-Xmx2g", "-XX:MaxDirectMemorySize=2g", "-XX:+ExitOnOutOfMemoryError", "-XX:+CrashOnOutOfMemoryError", "-XX:+HeapDumpOnOutOfMemoryError", "-XX:HeapDumpPath=/tmp/dumpfile/heap", "-XX:MaxRAMPercentage=50.0", "-XX:+UseContainerSupport", "-XX:+PrintExtendedThreadInfo"]` |
| `segmentStore.stsNameSuffix` | suffix for segmentStore sts name | `pravega-segment-store` |
| `segmentStore.headlessSvcNameSuffix` | suffix for segmentStore headless service name | `pravega-segmentstore-headless` |
| `segmentStore.labels` | Labels to add to the segmentStore pods | `{}` |
| `segmentStore.annotations` | Annotations to add to the segmentStore pods | `{}` |
| `segmentStore.initContainers` | Init Conatiners to add to the segmentStore pods | `[]` |
| `storage.longtermStorage.type` | Type of long term storage backend to be used (filesystem/ecs/hdfs) | `filesystem` |
| `storage.longtermStorage.filesystem.pvc` | Name of the pre-created PVC, if long term storage type is filesystem | `pravega-tier2` |
| `storage.longtermStorage.ecs` | Configuration to use a Dell EMC ECS system, if long term storage type is ecs | `{}` |
| `storage.longtermStorage.hdfs` | Configuration to use an HDFS system, if long term storage type is hdfs | `{}` |
| `storage.cache.className` | Storage class for cache volume | `` |
| `storage.cache.size` | Storage requests for cache volume | `20Gi` |
| `options` | List of Pravega options | |
