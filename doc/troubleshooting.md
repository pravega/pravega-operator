# Troubleshooting

## Pravega Cluster Issues

* [Certificate Error: Internal error occurred: failed calling webhook](#certificate-error-internal-error-occurred-failed-calling-webhook)
* [Segment store in CrashLoopBackOff](#segment-store-in-crashloopbackoff)
* [Controller pod not in ready state](#controller-pod-not-in-ready-state)
* [NFS volume mount failure: wrong fs type](#nfs-volume-mount-failure-wrong-fs-type)
* [Recover Statefulset when node fails](#recover-statefulset-when-node-fails)
* [External-IP details truncated in older Kubectl Client Versions](#external-ip-details-truncated-in-older-kubectl-client-versions)
* [Logs missing when Pravega upgrades](#logs-missing-when-Pravega-upgrades)
* [Collecting logs for crashed pod](#collecting-logs-for-crashed-pod)
* [Filtering kubectl Events](#filtering-kubectl-events)
* [Unrecognized VM option](#unrecognized-vm-option)

## Pravega operator Issues
* [Operator pod in container creating state](#operator-pod-in-container-creating-state)
* [Recover Operator when node fails](#recover-operator-when-node-fails)

## Certificate Error: Internal error occurred: failed calling webhook

While installing pravega, if we get the error as  below,
```
helm install pravega charts/pravega
Error: Internal error occurred: failed calling webhook "pravegawebhook.pravega.io": Post https://pravega-webhook-svc.default.svc:443/validate-pravega-pravega-io-v1beta1-pravegacluster?timeout=30s: x509: certificate signed by unknown authority
```
We need to ensure that certificates are installed before installing the operator. Please refer [prerequisite](../charts/pravega-operator/README.md#Prerequisites)

## Segment store in CrashLoopBackOff

If segmentstore goes in `CrashLoopBackOff` state while installing pravega and if segmentstore pod logs shows as below

```
2021-03-17 13:02:39,930 848  [main] ERROR o.a.b.client.BookieWatcherImpl - Failed to get bookie list :
org.apache.bookkeeper.client.BKException$ZKException: Error while using ZooKeeper
        at org.apache.bookkeeper.discover.ZKRegistrationClient.lambda$getChildren$0(ZKRegistrationClient.java:212)
        at org.apache.bookkeeper.zookeeper.ZooKeeperClient$25$1.processResult(ZooKeeperClient.java:1174)
        at org.apache.zookeeper.ClientCnxn$EventThread.processEvent(ClientCnxn.java:630)
        at org.apache.zookeeper.ClientCnxn$EventThread.run(ClientCnxn.java:510)
2021-03-17 13:02:39,931 849  [main] ERROR i.p.s.server.host.ServiceStarter - Could not start the Service, Aborting.
io.pravega.segmentstore.storage.DataLogNotAvailableException: Unable to establish connection to ZooKeeper or BookKeeper.
        at io.pravega.segmentstore.storage.impl.bookkeeper.BookKeeperLogFactory.initialize(BookKeeperLogFactory.java:116)
        at io.pravega.segmentstore.server.store.ServiceBuilder.initialize(ServiceBuilder.java:240)
        at io.pravega.segmentstore.server.host.ServiceStarter.start(ServiceStarter.java:95)
        at io.pravega.segmentstore.server.host.ServiceStarter.main(ServiceStarter.java:275)
Caused by: org.apache.bookkeeper.client.BKException$ZKException: Error while using ZooKeeper
        at org.apache.bookkeeper.discover.ZKRegistrationClient.lambda$getChildren$0(ZKRegistrationClient.java:212)
        at org.apache.bookkeeper.zookeeper.ZooKeeperClient$25$1.processResult(ZooKeeperClient.java:1174)
        at org.apache.zookeeper.ClientCnxn$EventThread.processEvent(ClientCnxn.java:630)
        at org.apache.zookeeper.ClientCnxn$EventThread.run(ClientCnxn.java:510)
```
This issue is happening because name of the pravega cluster (which can be seen with the output of `kubectl get pravegacluster`) is not matching `PRAVEGA_CLUSTER_NAME` provided in the bookkeeper configmap. Due to that it is looking at a wrong path for ledger location in znode. In order to resolve this issue, we have ensure that the cluster name is same as that in [bookkeeper configmap](https://github.com/pravega/bookkeeper-operator/blob/master/deploy/config_map.yaml)

## Controller pod not in ready state

While installing pravega, if the controller pod goes in `0/1` state as below, it can be due to meta data mismatch in znode

```
pravega-pravega-controller-68d68796f4-m5w7m             0/1     Running            0          15m
```
To resolve this issue, we have to ensure that zookeeper, bookkeeper and longterm storage are recreated before doing the pravega installation.

## NFS volume mount failure: wrong fs type

If you experience `wrong fs type` issues when pods are trying to mount NFS volumes like in the `kubectl describe po/pravega-segmentstore-0` snippet below, make sure that all Kubernetes node have the `nfs-common` system package installed. You can just try to run the `mount.nfs` command to make sure NFS support is installed in your system.

In PKS, make sure to use [`v1.2.3`](https://docs.pivotal.io/runtimes/pks/1-2/release-notes.html#v1.2.3) or newer. Older versions of PKS won't have NFS support installed in Kubernetes nodes.

```
Events:
  Type     Reason       Age                        From                                           Message
  ----     ------       ----                       ----                                           -------
  Warning  FailedMount  10m (x222 over 10h)        kubelet, 53931b0d-18f4-49fd-a105-49b1fea3f468  Unable to mount volumes for pod "nautilus-segmentstore-0_nautilus-pravega(79167f33-f73b-11e8-936a-005056aeca39)": timeout expired waiting for volumes to attach or mount for pod "nautilus-pravega"/"nautilus-segmentstore-0". list of unmounted volumes=[tier2]. list of unattached volumes=[cache tier2 pravega-segment-store-token-fvxql]
  Warning  FailedMount  <invalid> (x343 over 10h)  kubelet, 53931b0d-18f4-49fd-a105-49b1fea3f468  (combined from similar events): MountVolume.SetUp failed for volume "pvc-6fa77d63-f73b-11e8-936a-005056aeca39" : mount failed: exit status 32
Mounting command: systemd-run
Mounting arguments: --description=Kubernetes transient mount for   /var/lib/kubelet/pods/79167f33-f73b-11e8-936a-005056aeca39/volumes/kubernetes.io~nfs/pvc-6fa77d63-f73b-11e8-936a-005056aeca39 --scope -- mount -t nfs -o vers=4.1 10.100.200.247:/export/pvc-6fa77d63-f73b-11e8-936a-005056aeca39 /var/lib/kubelet/pods/79167f33-f73b-11e8-936a-005056aeca39/volumes/kubernetes.io~nfs/pvc-6fa77d63-f73b-11e8-936a-005056aeca39
Output: Running scope as unit run-rc77b988cdec041f6aa91c8ddd8455587.scope.
mount: wrong fs type, bad option, bad superblock on 10.100.200.247:/export/pvc-6fa77d63-f73b-11e8-936a-005056aeca39,
       missing codepage or helper program, or other error
       (for several filesystems (e.g. nfs, cifs) you might
       need a /sbin/mount.<type> helper program)

       In some cases useful info is found in syslog - try
       dmesg | tail or so.
```

## Recover Statefulset when node fails

When a node failure happens, unlike Deployment Pod, the Statefulset Pod on that failed node will not be rescheduled to other available nodes automatically.
This is because Kubernetes guarantees at most once execution of a Statefulset. See the [design](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/pod-safety.md).

If the failed node is not coming back, the cluster admin can manually recover the lost pod of Statefulset.
To do that, the cluster admin can delete the failed node object in the apiserver by running
```
kubectl delete node <node name>
```
After the failed node is deleted from Kubernetes, the Statefulset pods on that node will be rescheduled to other available nodes.

## External-IP details truncated in older Kubectl Client Versions

When Pravega is deployed with `external-access enabled`, an External-IP is assigned to its controller and segment store services, which is used by clients to access it. The External-IP details can be viewed in the output of the `kubectl get svc`.
However, when using kubectl client version `v1.10.x` or lower, the External-IP for the controller and segment store services appears truncated in the output.

```
# kubectl get svc
NAME                                    TYPE           CLUSTER-IP       EXTERNAL-IP        PORT(S)                          AGE
kubernetes                              ClusterIP      10.100.200.1     <none>             443/TCP                          6d
pravega-bookie-headless                 ClusterIP      None             <none>             3181/TCP                         6d
pravega-pravega-controller              LoadBalancer   10.100.200.11    10.240.124.15...   10080:31391/TCP,9090:30301/TCP   6d
pravega-pravega-segmentstore-0          LoadBalancer   10.100.200.59    10.240.124.15...   12345:30597/TCP                  6d
pravega-pravega-segmentstore-1          LoadBalancer   10.100.200.42    10.240.124.15...   12345:30840/TCP                  6d
pravega-pravega-segmentstore-2          LoadBalancer   10.100.200.83    10.240.124.15...   12345:31170/TCP                  6d
pravega-pravega-segmentstore-headless   ClusterIP      None             <none>             12345/TCP                        6d
pravega-zk-client                       ClusterIP      10.100.200.120   <none>             2181/TCP                         6d
pravega-zk-headless                     ClusterIP      None             <none>             2888/TCP,3888/TCP                6d

```

This problem has however been solved in kubectl client version `v1.11.0` onwards.

```
# kubectl get svc
NAME                                    TYPE           CLUSTER-IP       EXTERNAL-IP                     PORT(S)                          AGE
kubernetes                              ClusterIP      10.100.200.1     <none>                          443/TCP                          6d20h
pravega-bookie-headless                 ClusterIP      None             <none>                          3181/TCP                         6d3h
pravega-pravega-controller              LoadBalancer   10.100.200.11    10.240.124.155,100.64.112.185   10080:31391/TCP,9090:30301/TCP   6d3h
pravega-pravega-segmentstore-0          LoadBalancer   10.100.200.59    10.240.124.156,100.64.112.185   12345:30597/TCP                  6d3h
pravega-pravega-segmentstore-1          LoadBalancer   10.100.200.42    10.240.124.157,100.64.112.185   12345:30840/TCP                  6d3h
pravega-pravega-segmentstore-2          LoadBalancer   10.100.200.83    10.240.124.158,100.64.112.185   12345:31170/TCP                  6d3h
pravega-pravega-segmentstore-headless   ClusterIP      None             <none>                          12345/TCP                        6d3h
pravega-zk-client                       ClusterIP      10.100.200.120   <none>                          2181/TCP                         6d3h
pravega-zk-headless                     ClusterIP      None             <none>                          2888/TCP,3888/TCP                6d3h

```

Also, while using kubectl client version `v1.10.x` or lower, the complete External-IP can still be viewed by doing a `kubectl describe svc` for the concerned service.

```
# kubectl describe svc pravega-pravega-controller
Name:                     pravega-pravega-controller
Namespace:                default
Labels:                   app=pravega-cluster
                          component=pravega-controller
                          pravega_cluster=pravega
Annotations:              ncp/internal_ip_for_policy=100.64.161.119
Selector:                 app=pravega-cluster,component=pravega-controller,pravega_cluster=pravega
Type:                     LoadBalancer
IP:                       10.100.200.34
LoadBalancer Ingress:     10.247.114.149, 100.64.161.119
Port:                     rest  10080/TCP
TargetPort:               10080/TCPc
NodePort:                 rest  32097/TCP
Endpoints:                
Port:                     grpc  9090/TCP
TargetPort:               9090/TCP
NodePort:                 grpc  32705/TCP
Endpoints:                
Session Affinity:         None
External Traffic Policy:  Cluster
Events:                   <none>
```

## Logs missing when Pravega upgrades

Users may find Pravega logs for old pods to be missing post upgrade. This is because the operator uses the Kubernetes
[rolling update](https://kubernetes.io/docs/tutorials/stateful-application/basic-stateful-set/#updating-statefulsets)
strategy to upgrade pod one at a time. This strategy will use a new replicaset for the update, it will kill one pod in the
old replicaset and start a pod in the new replicaset in the meantime. So after upgrading, users are actually using a new
replicaset, thus the logs for the old pod cannot be obtained using `kubectl logs`.

## Collecting logs for crashed pod

For collecting logs for crashed pod, use the below command

```
kubectl logs <podname> --previous
```

## Filtering kubectl Events

For filtering kubernetes events, please use the below command,

```
kubectl get events --namespace default --field-selector involvedObject.name=<name of the object> --sort-by=.metadata.creationTimestamp
```

## Unrecognized VM option

While Installing pravega, if the pods didnt come up with error as below,

```
Unrecognized VM option 'PrintGCDateStamps'
Error: Could not create the Java Virtual Machine.
Error: A fatal exception has occurred. Program will exit.
```
This is happening because some of default JVM options in operator is not supported by Java version used by pravega. This can be resolved by setting the JVM option `IgnoreUnrecognizedVMOptions` as below.

```
helm install [RELEASE_NAME] pravega/pravega --version=[VERSION] --set zookeeperUri=[ZOOKEEPER_HOST] --set bookkeeperUri=[BOOKKEEPER_SVC] --set storage.longtermStorage.filesystem.pvc=[TIER2_NAME] --set 'controller.jvmOptions={-XX:+UseContainerSupport,-XX:+IgnoreUnrecognizedVMOptions}' --set 'segmentStore.jvmOptions={-XX:+UseContainerSupport,-XX:+IgnoreUnrecognizedVMOptions,-Xmx2g,-XX:MaxDirectMemorySize=2g}'
```

## Operator pod in container creating state

while installing operator, if the operator pod goes in `ContainerCreating` state for long time, make sure certificates are installed correctly.Please refer [prerequisite](../charts/pravega-operator/README.md#Prerequisites)

## Recover Operator when node fails

If the Operator pod is deployed on the node that fails, the pod will be rescheduled to a healthy node. However, the Operator will
not function properly because it has a leader election locking mechanism. See [here](https://github.com/operator-framework/operator-sdk/blob/v0.17.x/doc/proposals/leader-for-life.md).

To make it work, the cluster admin will need to delete the lock by running

```
kubectl delete configmap pravega-operator-lock
```
After that, the new Operator pod will become the leader. If the node comes up later, the extra Operator pod will
be deleted by Deployment controller.
