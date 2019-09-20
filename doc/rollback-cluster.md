# Pravega Cluster Rollback

This document details how manual rollback can be triggered after a Pravega cluster upgrade fails.
Note that a rollback can be triggered only on Upgrade Failure.

## Upgrade Failure

An Upgrade can fail because of following reasons:

1. Incorrect configuration (wrong quota, permissions, limit ranges)
2. Network issues (ImagePullError)
3. K8s Cluster Issues.
4. Application issues (Application runtime misconfiguration or code bugs)

An upgrade failure can manifest through a Pod staying in `Pending` state forever or continuously restarting or crashing (CrashLoopBackOff).
A component deployment failure needs to be tracked and mapped to "Upgrade Failure" for Pravega Cluster.
Here we try to fail-fast by explicitly checking for some common causes for deployment failure like image pull errors or  CrashLoopBackOff State and failing the upgrade if any pod runs into this state during upgrade.

The following Pravega Cluster Status Condition indicates a Failed Upgrade:

```
ClusterConditionType: Error
Status: True
Reason: UpgradeFailed
Message: <Details of exception/cause of failure>
```
After an Upgrade Failure the output of `kubectl describe pravegacluster pravega` would look like this:

```
$> kubectl describe pravegacluster pravega
. . .
Spec:
. . .
Version:        0.6.0-2252.b6f6512
. . .
Status:
. . .
Conditions:
    Last Transition Time:  2019-09-06T09:00:13Z
    Last Update Time:      2019-09-06T09:00:13Z
    Status:                False
    Type:                  Upgrading
    Last Transition Time:  2019-09-06T08:58:40Z
    Last Update Time:      2019-09-06T08:58:40Z
    Status:                False
    Type:                  PodsReady
    Last Transition Time:  2019-09-06T09:00:13Z
    Last Update Time:      2019-09-06T09:00:13Z
    Message:               failed to sync segmentstore version. pod pravega-pravega-segmentstore-0 update failed because of ImagePullBackOff
    Reason:                UpgradeFailed
    Status:                True
    Type:                  Error
  . . .
  Current Version:         0.6.0-2239.6e24df7
. . .
Version History:
    0.6.0-2239.6e24df7
```
where `0.6.0-2252.b6f6512` is the version we tried upgrading to and `0.6.0-2239.6e24df7` is the cluster version prior to triggering the upgrade.

## Manual Rollback Trigger

A Rollback is triggered when a Pravega Cluster is in `UpgradeFailed` Error State and a user manually updates version feild in the PravegaCluster spec to point to the last stable cluster version.

A Rollback involves moving all components in the cluster back to the last stable cluster version. As with upgrades, the operator rolls back one component at a time and one pod at a time to preserve high-availability.

Note:
1. A Rollback to only the last stable cluster version is supported at this point.
2. Changing the cluster spec version to the previous cluster version, when cluster is not in `UpgradeFailed` state, will not trigger a rollback.

## Rollback Implementation

When Rollback is triggered the cluster moves into ClusterCondition `RollbackInProgress`.
Once the Rollback completes, this condition is set to false.

During a Rollback, the Cluster Status should look something like:
```
$> kubectl describe pravegacluster pravega
. . .
Status:
  Conditions:
    Last Transition Time:  2019-09-20T10:41:10Z
    Last Update Time:      2019-09-20T10:41:10Z
    Status:                False
    Type:                  Upgrading
    Last Transition Time:  2019-09-20T10:45:12Z
    Last Update Time:      2019-09-20T10:45:12Z
    Status:                True
    Type:                  PodsReady
    Last Transition Time:  2019-09-20T10:41:10Z
    Last Update Time:      2019-09-20T10:41:10Z
    Message:               failed to sync segmentstore version. pod pravega-pravega-segmentstore-0 update failed because of ImagePullBackOff
    Reason:                UpgradeFailed
    Status:                True
    Type:                  Error
    Last Update Time:      2019-09-20T10:45:12Z
    Message:               1
    Reason:                Updating Bookkeeper
    Status:                True
    Type:                  RollbackInProgress
. . .
```
Here the `RollbackInProgress` condition being `true` indicates that a Rollback is in Progress.
Also `Reason` and `Message` feilds of this condition indicate the component being rolled back and number of updated replicas respectively.

The operator rolls back components following the reverse upgrade order :

1. Pravega Controller
2. Pravega Segment Store
3. BookKeeper

A `versionHistory` field in the PravegaClusterSpec maintains the history of upgrades.

## Rollback Outcome

### Success
If the Rollback completes successfully, the cluster state goes back to condition `PodsReady`, which would mean the cluster is now in a stable state. All other conditions should be `false`.
```
Last Transition Time:  2019-09-20T09:49:26Z
Last Update Time:      2019-09-20T09:49:26Z
Status:                True
Type:                  PodsReady

```

Example:
```
Status:
  Conditions:
    Last Transition Time:  2019-09-20T10:12:04Z
    Last Update Time:      2019-09-20T10:12:04Z
    Status:                False
    Type:                  Upgrading
    Last Transition Time:  2019-09-20T10:11:34Z
    Last Update Time:      2019-09-20T10:11:34Z
    Status:                True
    Type:                  PodsReady
    Last Transition Time:  2019-09-20T10:07:19Z
    Last Update Time:      2019-09-20T10:07:19Z
    Status:                False
    Type:                  Error
    Last Transition Time:  2019-09-20T09:50:57Z
    Last Update Time:      2019-09-20T09:50:57Z
    Status:                False
    Type:                  RollbackInProgress
```

### Failure
If the Rollback Fails, the cluster would move to `Error` state indicated by this cluster condition:
```
ClusterConditionType: Error
Status: True
Reason: RollbackFailed
Message: <Details of exception/cause of failure>
```

Example:
```
Status:
  Conditions:
    Last Transition Time:  2019-09-20T09:46:24Z
    Last Update Time:      2019-09-20T09:46:24Z
    Status:                False
    Type:                  Upgrading
    Last Transition Time:  2019-09-20T09:49:26Z
    Last Update Time:      2019-09-20T09:49:26Z
    Status:                False
    Type:                  PodsReady
    Last Transition Time:  2019-09-20T09:46:24Z
    Last Update Time:      2019-09-20T09:50:57Z
    Message:               failed to sync bookkeeper version. pod pravega-bookie-0 update failed because of ImagePullBackOff
    Reason:                RollbackFailed
    Status:                True
    Type:                  Error
    Last Transition Time:  2019-09-20T09:50:57Z
    Last Update Time:      2019-09-20T09:50:57Z
    Status:                False
    Type:                  RollbackInProgress
```

When a rollback failure happens, manual intervention would be required to resolve this.
After checking and solving the root cause of failure, to bring the cluster back to a stable state, a user can upgrade to:
1. The version to which a user initially intended to upgrade.(when upgrade failure was noticed)
2. To any other supported version based versions of all pods in the cluster.
