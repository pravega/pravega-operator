# Pravega upgrade guide

This document shows how to upgrade a Pravega cluster managed by the operator to a desired version while preserving the cluster's state and data whenever possible.

> **Warning:** The upgrade feature is still considered experimental, so its use is discouraged if you care about your data.


## Overview

The activity diagram below shows the overall upgrade process started by an end-user and performed by the operator.

![pravega k8 upgrade 1](https://user-images.githubusercontent.com/3786750/51993601-7908b000-24af-11e9-8149-82fd1b036630.png)


## Limitations

- The rollback mechanism is on the roadmap but not implemented yet. Check out [this issue](https://github.com/pravega/pravega-operator/issues/153) for tracking. When a cluster fails to upgrade, manual recovery is necessary.
- There is no validation of the configured desired version (Issue #)
-

## Prerequisites

Your Pravega cluster should be in a healthy state. You can check your cluster health by listing it and checking that all members are ready.

```
$ kubectl get PravegaCluster
NAME      VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
example   0.4.0     7                 7               11m
```

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
2. Pravega Controller
3. Pravega Segment Store


The upgrade workflow is as follows:

- The operator will change the `Upgrade` condition to `True` to indicate that the cluster resource has an upgrade in progress.
- For each cluster component, the operator will check if the current component version matches the target version.
  - If it does, it will move to the next component
  - If it doesn't, the operator will trigger the upgrade of that component. This is a process that can span to multiple reconcile iterations. In each iteration, the operator will check the state of the pods. [Check below](#) to see how each component is upgraded.
  - If any of the component pods has errors, the upgrade process will stop (`Upgrade` condition to `False`) and operator will set the `Error` condition to `True` and indicate the reason.
- When all components are upgraded, the `Upgrade` condition will be set to `False` and `status.currentVersion` will be updated to the desired version.


### BookKeeper upgrade


### Pravega Segment Store upgrade


### Pravega Controller upgrade


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

If your upgrade has failed, you can describe the status section of your Pravega cluster to discover why.

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
INFO[5930] error syncing cluster version, need manual intervention. failed to sync bookkeeper version. pod example-bookie-0 is restarting
...
```
