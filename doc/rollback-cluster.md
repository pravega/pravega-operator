# Pravega cluster rollback

This document shows how to automated rollback of Pravega cluster is implemented by the operator  while preserving the cluster's state and data whenever possible.

## Failing an Upgrade

An Upgrade can fail because of following reasons:

1. Incorrect configuration (wrong quota, permissions, limit ranges)
2. Network issues (ImagePullError)
3. K8s Cluster Issues.
4. Application issues (Application runtime misconfiguration or code bugs)

An upgrade failure can manifest through a Pod to staying in `Pending` state forever or having continous restarts after moving to Running state (CrashLoopBackOff).
Here we try to fail-fast by explicitly checking for some common causes for upgrade failure like `ErrImagePull` and failing the upgrade if any pod faces this issue during upgrade.
We also have a time threshold within which deployment to a pod should complete. If it does not, then we fail the upgrade.
To indicate upgrade failure we set the folling condition on PravegaCluster status:

```
ClusterConditionType: Error
Status: True
Reason: UpgradeFailed
Message: <Details of exception/cause of failure>
```

## Rollback Trigger

Rollback is triggered by the PravegaCluster moving to `ClusterConditionType: Error` with `Reason:UpgradeFailed` state.

## Rollback implementation
When Rollback is started cluster moves into ClusterCondition `RollbackInProgress`.
Once Rollback completes this condition is set to false.

A new data structure is added clusterStatus to maintain all previous cluster versions .
```
VersionHistory []string `json:"versionHistory,omitempty"`
```
For now, operator would support automated rollback only to the previous cluster version. Later, operator may support rollback to any supported previous version, but this would need to be invoked manually.

Rollback involves moving each component in the cluster back to its previous cluster version. As in case of upgrade, operator would rollback one component at a time and one pod at a time.

If Rollback completes successfully, cluster state would be set back to `PodsReady` which would mean the cluster is now in a stable state.
If Rollback Fails, cluster would move to state `RollbackError` and User would be prompted for manual intervention.







## Pending tasks


## Prerequisites
