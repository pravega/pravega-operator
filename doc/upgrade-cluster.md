# Pravega cluster upgrade

This document shows how to upgrade a Pravega cluster managed by the operator to a desired version while preserving the cluster's state and data whenever possible.

> **Warning:** The upgrade feature is still considered experimental, so its use is discouraged if you care about your data.

## Overview

The activity diagram below shows the overall upgrade process started by an end-user and performed by the operator.

![pravega k8 upgrade 1](https://user-images.githubusercontent.com/3786750/51993601-7908b000-24af-11e9-8149-82fd1b036630.png)

## Scope

Pravega is made up multiple components which can be differentiated into two blocks: internal and external components. Internal components refer to the Pravega Controller, the Segment Store, and BookKeeper. Whereas external components refer to ZooKeeper and the Tier 2 backend.

The Pravega operator scope is limited to internal components, and therefore, this upgrade procedure will only affect the aforementioned components.

Check out [Pravega documentation](http://pravega.io/docs/latest/) for more information about Pravega internals.

## Pending tasks

- There is no validation of the configured desired version. Check out [this issue](https://github.com/pravega/pravega-operator/issues/156)


## Prerequisites

Your Pravega cluster should be in a healthy state. You can check your cluster health by listing it and checking that all members are ready.

```
$ kubectl get PravegaCluster
NAME      VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
example   0.4.0     7                 7               11m
```

## Upgrade Path Matrix

| BASE VERSION | TARGET VERSION                   |
| ------------ | ----------------                 |
| 0.1.0        | 0.1.0                            |
| 0.2.0        | 0.2.0                            |
| 0.3.0        | 0.3.0, 0.3.1, 0.3.2              |
| 0.3.1        | 0.3.1, 0.3.2                     |
| 0.3.2        | 0.3.2                            |
| 0.4.0        | 0.4.0                            |
| 0.5.0        | 0.5.0, 0.5.1, 0.6.0, 0.6.1, 0.7.0|
| 0.5.1        | 0.5.1, 0.6.0, 0.6.1, 0.7.0       |
| 0.6.0        | 0.6.0, 0.6.1, 0.7.0              |
| 0.6.1        | 0.6.1, 0.7.0                     |
| 0.7.0        | 0.7.0                            |

## Trigger an upgrade

To initiate an upgrade process, a user has to update the `spec.version` field on the `PravegaCluster` custom resource. This can be done in three different ways using the `kubectl` command.
1. `kubectl edit PravegaCluster <name>`, modify the `version` value in the YAML resource, save, and exit.
2. If you have the custom resource defined in a local YAML file, e.g. `pravega.yaml`, you can modify the `version` value, and reapply the resource with `kubectl apply -f pravega.yaml`.
3. `kubectl patch PravegaCluster <name> --type='json' -p='[{"op": "replace", "path": "/spec/version", "value": "X.Y.Z"}]'`.

After the `version` field is updated, the operator will detect the version change and it will trigger the upgrade process.

## Upgrade process

![pravega operator component update](https://user-images.githubusercontent.com/3786750/51993862-f3d1cb00-24af-11e9-857d-281eceb7fd90.png)

Once an upgrade request has been received, the operator will apply the rolling upgrade to all the components that are part of the Pravega cluster.

The order in which the components will be upgraded is the following:

1. BookKeeper
2. Pravega Segment Store
3. Pravega Controller

The upgrade workflow is as follows:

- The operator will change the `Upgrade` condition to `True` to indicate that the cluster resource has an upgrade in progress.
- For each cluster component, the operator will check if the current component version matches the target version.
  - If it does, it will move to the next component
  - If it doesn't, the operator will trigger the upgrade of that component. This is a process that can span to multiple reconcile iterations. In each iteration, the operator will check the state of the pods. Check below to understand how each component is upgraded.
  - If any of the component pods has errors, the upgrade process will stop (`Upgrade` condition to `False`) and operator will set the `Error` condition to `True` and indicate the reason.
- When all components are upgraded, the `Upgrade` condition will be set to `False` and `status.currentVersion` will be updated to the desired version.


### BookKeeper upgrade

BookKeeper is the first component to be upgraded as recommended in [BookKeeper documentation](https://bookkeeper.apache.org/docs/4.7.3/admin/upgrade/).

In Kubernetes, BookKeeper is deployed as a [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) due to its requirements on:

- Persistent storage: each bookie has three persistent volume for ledgers, journals, and indices. If a pod is migrated or recreated (e.g. when it's upgraded), the data in those volumes will remain untouched.
- Stable network names: the `StatefulSet` provides pods with a predictable name and a [Headless service](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) creates DNS records for pods to be reachable by clients. If a pod is recreated or migrated to a different node, clients will continue to be able to reach the pod despite changing its IP address.

Statefulset [upgrade strategy](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#update-strategies) is configured in the `updateStrategy` field. It supports two type of strategies.

- `RollingUpdate`. The statefulset will automatically apply a rolling upgrade to the pods.
- `OnDelete`. The statefulset will not automatically upgrade pods. Pods will be updated when they are recreated after being deleted.

In both cases, the upgrade is initiated when the Pod template is updated.

For BookKeeper, the operator uses an `OnDelete` strategy. With `RollingUpdate` strategy, you can only check the upgrade status once all pods get upgraded. On the other hand, with `OnDelete` you can keep updating pod one by one and keep checking the application status to make sure the upgrade working fine. This allows the operator to have control over the upgrade process and perform verifications and actions before and after a BookKeeper pod is upgraded. For example, checking that there are no under-replicated ledgers before upgrading the next pod. Also, the operator might be need to apply migrations when upgrading to a certain version.

BookKeeper upgrade process is as follows:

1. Statefulset Pod template is updated to the new image and tag according to the Pravega version.
2. Pick one outdated pod
3. Apply pre-upgrade actions and verifications
4. Delete the pod. The pod is recreated with an updated spec and version
5. Wait for the pod to become ready. If it fails to start or times out, the upgrade is cancelled. Check [Recovering from a failed upgrade](#recovering-from-a-failed-upgrade)
6. Apply post-upgrade actions and verifications
7. If all pods are updated, BookKeeper upgrade is completed. Otherwise, go to 2.


### Pravega Segment Store upgrade

Pravega Segment Store, as a consumer of BookKeeper, is the second component to be upgraded. As BookKeeper, Segment Store is also deployed as a Statefulset due to similar reasons.

Segment Store instances need access to a persistent volume to store the cache. Losing the data on the cache would not be critical because the Segment Store would be able to recover based on the data in BookKeeper, but it would have a performance impact that we want to avoid when migrating, recreating, or upgrading Segment Store pods.

Also, Segment Store pods need to be individually accessed by clients, so having a stable network identifier provided by the Statefulset and a headless service is very convenient.

Same as Bookkeeper, we use `OnDelete` strategy for Segment Store. The reason that we don't use `RollingUpdate` strategy here is that we found it convenient to manage the upgrade and rollback in the same fashion. Using `RollingUpdate` will introduce Kubernetes rollback mechanism which will cause trouble to our implementation.

### Pravega Controller upgrade

The Controller is the last one to be upgraded. As opposed to BookKeeper and Segment Store, the Controller is a stateless component, meaning that it doesn't need to storage data on a volume and it doesn't need to have a stable identify. Controller pods are frontended with a service that load balances requests to pods. Due to this nature, the Controller is deployed as a Kubernetes [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/).

The Controller upgrade is also triggered by updating the Pod template, and a `RollingUpgrade` strategy is applied as we don't need to apply any verification or action other than waiting for the Pod to become ready after being upgraded.


### Monitor the upgrade process

You can monitor your upgrade process by listing the Pravega clusters. If a desired version is shown, it means that the operator is working on updating the version.

```
$ kubectl get PravegaCluster
NAME      VERSION   DESIRED VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
example   0.4.0     0.5.0             8                 7               1h
```

When the upgrade process has finished, the version will be updated.

```
$ kubectl get PravegaCluster
NAME      VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
example   0.5.0     8                 8               1h
```

The command `kubectl describe` can be used to track progress of the upgrade.
```
$ kubectl describe PravegaCluster example
...
Status:
  Conditions:
    Status:                True
    Type:                  Upgrading
    Reason:                Updating BookKeeper
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
$ kubectl describe PravegaCluster example
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
    Message:               failed to sync bookkeeper version. pod example-bookie-0 is restarting
    Reason:                UpgradeFailed
    Status:                True
    Type:                  Error
  Current Replicas:        8
  Current Version:         0.4.0
  Members:
    Ready:
      example-bookie-1
      example-bookie-2
      example-pravega-controller-64ff87fc49-kqp9k
      example-pravega-segmentstore-0
      example-pravega-segmentstore-1
      example-pravega-segmentstore-2
      example-pravega-segmentstore-3
    Unready:
      example-bookie-0
  Ready Replicas:  7
  Replicas:        8
```

You can also find useful information at the operator logs.

```
...
INFO[5884] syncing cluster version from 0.4.0 to 0.5.0-1
INFO[5885] Reconciling PravegaCluster default/example
INFO[5886] updating statefulset (example-bookie) template image to 'adrianmo/bookkeeper:0.5.0-1'
INFO[5896] Reconciling PravegaCluster default/example
INFO[5897] statefulset (example-bookie) status: 0 updated, 3 ready, 3 target
INFO[5897] upgrading pod: example-bookie-0
INFO[5899] Reconciling PravegaCluster default/example
INFO[5900] statefulset (example-bookie) status: 1 updated, 2 ready, 3 target
INFO[5929] Reconciling PravegaCluster default/example
INFO[5930] statefulset (example-bookie) status: 1 updated, 2 ready, 3 target
INFO[5930] error syncing cluster version, upgrade failed. failed to sync bookkeeper version. pod example-bookie-0 is restarting
...
```

### Recovering from a failed upgrade

See [Rollback](rollback-cluster.md)
