# Pravega cluster rollback

This document shows how to automated rollback of Pravega cluster is implemented by the operator  while preserving the cluster's state and data whenever possible.

## Failing an Upgrade

An Upgrade can fail because of following reasons:

1. Incorrect configuration (wrong quota, permissions, limit ranges)
2. Network issues (ImagePullError)
3. K8s Cluster Issues.
4. Application issues (Application runtime misconfiguration or code bugs)

An upgrade failure can manifest through a Pod to staying in `Pending` state forever or continuously restarting or crashing (CrashLoopBackOff).
A component deployment failure needs to be tracked and mapped to "Upgrade Failure" for Pravega Cluster.
Here we try to fail-fast by explicitly checking for some common causes for deployment failure like image pull errors or  CrashLoopBackOff State and failing the upgrade if any pod runs into this state during upgrade.

The following Pravega Cluster Status Condition indicates an Upgrade Failure:

```
ClusterConditionType: Error
Status: True
Reason: UpgradeFailed
Message: <Details of exception/cause of failure>
```

## Rollback Trigger

A Rollback is triggered by Upgrade Failure condition i.e the Cluster moving to
`ClusterConditionType: Error` and
`Reason:UpgradeFailed` state.

## Rollback Implementation
When Rollback is started cluster moves into ClusterCondition `RollbackInProgress`.
Once Rollback completes this condition is set to false.
The order in which the components are rolled back is the following:

1. BookKeeper
2. Pravega Segment Store
3. Pravega Controller

A new field `versionHistory` has been added to Pravega ClusterStatus to maintain history of previous cluster versions .
```
VersionHistory []string `json:"versionHistory,omitempty"`
```
Currently, operator only supports automated rollback to the previous cluster version.
Later, rollback to any other previous version(s), may be supported.

Rollback involves moving all components in the cluster back to the previous cluster version. As in case of upgrade, operator would rollback one component at a time and one pod at a time to maintain HA.

If Rollback completes successfully, cluster state goes back to `PodsReady` which would mean the cluster is now in a stable state.
If Rollback Fails, cluster would move to state `RollbackError` and User would be prompted for manual intervention.







## Pending tasks


## Prerequisites
