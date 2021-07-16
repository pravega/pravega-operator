# Pravega cluster upgrade

This document shows how to upgrade a Pravega cluster managed by the operator to a desired version while preserving the cluster's state and data whenever possible.

## Overview

The activity diagram below shows the overall upgrade process started by an end-user and performed by the operator.

![pravega k8 upgrade 1](https://user-images.githubusercontent.com/3786750/51993601-7908b000-24af-11e9-8149-82fd1b036630.png)

## Scope

Pravega is made up of multiple components which can be differentiated into two blocks: internal and external components. Internal components refer to the Pravega Controller and the Segment Store. Whereas external components refer to BookKeeper, ZooKeeper and the Tier 2 backend.

The Pravega operator scope is limited to the internal components, and therefore, this upgrade procedure will only affect the aforementioned components.

Check out [Pravega documentation](http://pravega.io/docs/latest/) for more information about Pravega internals.

## Prerequisites

Your Pravega cluster should be in a healthy state. You can check your cluster health by listing it and checking that all members are ready.

```
$ kubectl get PravegaCluster
NAME          VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
bar-pravega   0.4.0     7                 7               11m
```

## Valid Upgrade Paths

Upgrade of pravega cluster to any version will be allowed as long as the user does not try to downgrade the cluster version.

## Trigger an upgrade

### Upgrading via Helm

The upgrade of the pravega cluster from a version **[OLD_VERSION]** to **[NEW_VERSION]** can be triggered via helm using the following command
```
$ helm upgrade [PRAVEGA_RELEASE_NAME] pravega/pravega --version=[NEW_VERSION] --set version=[NEW_VERSION] --reuse-values --timeout 600s
```
**Note:** By specifying the `--reuse-values` option, the configuration of all parameters are retained across upgrades. However if some values need to be modified during the upgrade, the `--set` flag can be used to specify the new configuration for these parameters. Also, by skipping the `reuse-values` flag, the values of all parameters are reset to the default configuration that has been specified in the published charts for version [NEW_VERSION].

**Note:** If the operator version is 0.5.1 or below and we are upgrading pravega version to 0.9.0 or above, we have to set controller and segmentstore JVM options as follows.

```
$ helm upgrade [PRAVEGA_RELEASE_NAME] pravega/pravega --version=[NEW_VERSION] --set version=[NEW_VERSION] --set 'controller.jvmOptions={-XX:+UseContainerSupport,-XX:+IgnoreUnrecognizedVMOptions}' --set 'segmentStore.jvmOptions={-XX:+UseContainerSupport,-XX:+IgnoreUnrecognizedVMOptions,-Xmx2g,-XX:MaxDirectMemorySize=2g}'  --reuse-values --timeout 600s
```
Based on the cluster flavours selected,segmentstore memory requirements needs to be adjusted.

### Upgrading manually

To initiate the upgrade process manually, a user has to update the `spec.version` field on the `PravegaCluster` custom resource. This can be done in three different ways using the `kubectl` command.
1. `kubectl edit PravegaCluster [CLUSTER_NAME]`, modify the `version` value in the YAML resource, save, and exit.
2. If you have the custom resource defined in a local YAML file, e.g. `pravega.yaml`, you can modify the `version` value, and reapply the resource with `kubectl apply -f pravega.yaml`.
3. `kubectl patch PravegaCluster [CLUSTER_NAME] --type='json' -p='[{"op": "replace", "path": "/spec/version", "value": "X.Y.Z"}]'`.

After the `version` field is updated, the operator will detect the version change and it will trigger the upgrade process.

### Upgrade guide

> Note: To trigger an upgrade please edit only the fields mentioned in this upgrade guide. We do not recommend clubbing other edits (for performance or scaling or anything else) along with the upgrade trigger. Those can be done either prior to triggering the upgrade or after the upgrade has completed successfully.

#### Upgrade to Pravega 0.7 or above

When upgrading the Pravega Cluster from any version below 0.7 to version 0.7 or above, there are a few configuration changes that must be made to Pravega manifest either with the upgrade request or prior to starting the upgrade.

1. Ensure that sufficient resources are allocated to segmentstore pods when moving to Pravega version 0.7 or later.
```
segmentStoreResources:
  requests:
    memory: "4Gi"
    cpu: "2000m"
  limits:
    memory: "16Gi"
    cpu: "8000m"
```

2. Distribute the pod's memory (POD_MEM_LIMIT) between JVM Heap and Direct Memory. For instance, if POD_MEM_LIMIT=16GB then we can set 4GB for JVM and the rest for Direct Memory (12GB) i.e. POD_MEM_LIMIT (16GB) = JVM Heap (4GB) + Direct Memory (12GB).
We need to ensure that the sum of JVM Heap and Direct Memory is not higher than the pod memory limit. In general, we can keep the JVM Heap fixed to 4GB and make the Direct Memory as the variable part.
These two options can be configured through the following field of the manifest file
```
segmentStoreJVMOptions: ["-Xmx4g", "-XX:MaxDirectMemorySize=12g"]
```

3. The cache should be configured at least 1 or 2 GB below the Direct Memory value provided since the Direct Memory is used by other components as well (like Netty). This value is configured in the pravega options part of the manifest file
```
options:
  pravegaservice.cache.size.max: "11811160064"
```

To summarize the way in which the segmentstore pod memory is distributed:

```
POD_MEM_LIMIT = JVM Heap + Direct Memory
Direct Memory = pravegaservice.cache.size.max + 1GB/2GB (other uses)
```
**Note:** If we are upgrading pravega version to 0.9 or above using operator version 0.5.1 or below, add the below JVM options for controller and segmentstore in addition to the current JVM options.
```
segmentStoreJVMOptions: ["-XX:+UseContainerSupport","-XX:+IgnoreUnrecognizedVMOptions"]

controllerjvmOptions: ["-XX:+UseContainerSupport","-XX:+IgnoreUnrecognizedVMOptions"]
```

## Upgrade process

![pravega operator component update](https://user-images.githubusercontent.com/3786750/51993862-f3d1cb00-24af-11e9-857d-281eceb7fd90.png)

Once an upgrade request has been received, the operator will apply the rolling upgrade to all the components that are part of the Pravega cluster.

The order in which the components will be upgraded is the following:

1. Pravega Segment Store
2. Pravega Controller

The upgrade workflow is as follows:

- The operator will change the `Upgrade` condition to `True` to indicate that the cluster resource has an upgrade in progress.
- For each cluster component, the operator will check if the current component version matches the target version.
  - If it does, it will move to the next component
  - If it doesn't, the operator will trigger the upgrade of that component. This is a process that can span to multiple reconcile iterations. In each iteration, the operator will check the state of the pods. Check below to understand how each component is upgraded.
  - If any of the component pods has errors, the upgrade process will stop (`Upgrade` condition to `False`) and operator will set the `Error` condition to `True` and indicate the reason.
- When all components are upgraded, the `Upgrade` condition will be set to `False` and `status.currentVersion` will be updated to the desired version.

### Pravega Segment Store upgrade

Pravega Segment Store is the first component to be upgraded. The Segment Store is deployed as a [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) due to its requirements on:

- Stable network names: the `StatefulSet` provides pods with a predictable name and a [Headless service](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) creates DNS records for pods to be reachable by clients. If a pod is recreated or migrated to a different node, clients will continue to be able to reach the pod despite changing its IP address. As Segment Store pods need to be individually accessed by clients, so having a stable network identifier provided by the Statefulset and a headless service is very convenient.

Statefulset [upgrade strategy](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#update-strategies) is configured in the `updateStrategy` field. It supports two type of strategies.

- `RollingUpdate`. The statefulset will automatically apply a rolling upgrade to the pods.
- `OnDelete`. The statefulset will not automatically upgrade pods. Pods will be updated when they are recreated after being deleted.

In both cases, the upgrade is initiated when the Pod template is updated.

For Segment Store, the operator uses an `OnDelete` strategy. With `RollingUpdate` strategy, you can only check the upgrade status once all pods get upgraded. On the other hand, with `OnDelete` you can keep updating pods one by one and keep checking the application status to make sure that the upgrade is working fine. This allows the operator to have control over the upgrade process and perform verifications and other actions before and after a Segment Store pod is upgraded. The operator might also need to apply migrations when upgrading to a certain version.

Segment Store upgrade process is as follows:

1. Statefulset Pod template is updated to the new image and tag according to the Pravega version.
2. Pick one outdated pod
3. Apply pre-upgrade actions and verifications
4. Delete the pod. The pod is recreated with an updated spec and version
5. Wait for the pod to become ready. If it fails to start or times out, the upgrade is cancelled. Check [Recovering from a failed upgrade](#recovering-from-a-failed-upgrade)
6. Apply post-upgrade actions and verifications
7. If all pods are updated, Segment Store upgrade is completed. Otherwise, go to 2.

### Pravega Controller upgrade

The Controller is the last one to be upgraded. As opposed to the Segment Store, the Controller is a stateless component, meaning that it doesn't need to store data on a volume and it doesn't need to have a stable identify. Controller pods are frontended with a service that load balances requests to pods. Due to this nature, the Controller is deployed as a Kubernetes [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/).

The Controller upgrade is also triggered by updating the Pod template, and a `RollingUpgrade` strategy is applied as we don't need to apply any verification or action other than waiting for the Pod to become ready after being upgraded.


### Monitor the upgrade process

You can monitor your upgrade process by listing the Pravega clusters. If a desired version is shown, it means that the operator is working on updating the version.

```
$ kubectl get PravegaCluster
NAME          VERSION   DESIRED VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
bar-pravega   0.4.0     0.5.0             5                 4               1h
```

When the upgrade process has finished, the version will be updated.

```
$ kubectl get PravegaCluster
NAME          VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
bar-pravega   0.5.0     5                 5               1h
```

The command `kubectl describe` can be used to track progress of the upgrade.
```
$ kubectl describe PravegaCluster bar-pravega
...
Status:
  Conditions:
    Status:                True
    Type:                  Upgrading
    Reason:                Updating Segmentstore
    Message:               1
    Last Transition Time:  2019-04-01T19:42:37+02:00
    Last Update Time:      2019-04-01T19:42:37+02:00
    Status:                False
    Type:                  PodsReady
    Last Transition Time:  2019-04-01T19:43:08+02:00
    Last Update Time:      2019-04-01T19:43:08+02:00
    Status:                False
    Type:                  Error
...  

```
The `Reason` field in Upgrading Condition shows the component currently being upgraded and `Message` field reflects number of successfully upgraded replicas in this component.

If upgrade has failed, please check the `Status` section to understand the reason for failure.

```
$ kubectl describe PravegaCluster bar-pravega
...
Status:
  Conditions:
    Status:                False
    Type:                  Upgrading
    Last Transition Time:  2019-04-01T19:42:37+02:00
    Last Update Time:      2019-04-01T19:42:37+02:00
    Status:                False
    Type:                  PodsReady
    Last Transition Time:  2019-04-01T19:43:08+02:00
    Last Update Time:      2019-04-01T19:43:08+02:00
    Message:               failed to sync segmentstore version. pod bar-pravega-pravega-segmentstore-0 is restarting
    Reason:                UpgradeFailed
    Status:                True
    Type:                  Error
  Current Replicas:        5
  Current Version:         0.4.0
  Members:
    Ready:
      bar-pravega-pravega-controller-64ff87fc49-kqp9k
      bar-pravega-pravega-segmentstore-1
      bar-pravega-pravega-segmentstore-2
      bar-pravega-pravega-segmentstore-3
    Unready:
      bar-pravega-pravega-segmentstore-0
  Ready Replicas:  4
  Replicas:        5
```

You can also find useful information at the operator logs.

```
...
INFO[5884] syncing cluster version from 0.4.0 to 0.5.0-1
INFO[5885] Reconciling PravegaCluster default/bar-pravega
INFO[5886] updating statefulset (bar-pravega-pravega-segmentstore) template image to 'pravega/pravega:0.5.0-1'
INFO[5896] Reconciling PravegaCluster default/bar-pravega
INFO[5897] upgrading pod: bar-pravega-pravega-segmentstore-0
INFO[5899] Reconciling PravegaCluster default/bar-pravega
INFO[5900] statefulset (bar-pravega-pravega-segmentstore) status: 1 updated, 3 ready, 4 target
INFO[5929] Reconciling PravegaCluster default/bar-pravega
INFO[5930] statefulset (bar-pravega-pravega-segmentstore) status: 1 updated, 3 ready, 4 target
INFO[5930] error syncing cluster version, upgrade failed. failed to sync segmentstore version. pod bar-pravega-pravega-segmentstore-0 is restarting
...
```

### Recovering from a failed upgrade

See [Rollback](rollback-cluster.md)
