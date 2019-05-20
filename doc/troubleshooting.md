# Troubleshooting

### Table of contents:

* [Helm Error: no available release name found](#helm-error-no-available-release-name-found)
* [NFS volume mount failure: wrong fs type](#nfs-volume-mount-failure-wrong-fs-type)
* [Recover Statefulset when node fails](#recover-statefulset-when-node-fails)
* [Recover Operator when node fails](#recover-operator-when-node-fails)

## Helm Error: no available release name found

When installing a cluster for the first time using `kubeadm`, the initialization defaults to setting up RBAC controlled access, which messes with permissions needed by Tiller to do installations, scan for installed components, and so on. `helm init` works without issue, but `helm list`, `helm install` and other commands do not work.

```
$ helm install stable/nfs-server-provisioner
Error: no available release name found
```
The following workaround can be applied to resolve the issue:

1. Create a service account for the Tiller.

```
kubectl create serviceaccount --namespace kube-system tiller
```

2. Bind that service account to the `cluster-admin` ClusterRole.

```
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
```

3. Add the service account to the Tiller deployment.

```
kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'
```

The above commands should resolve the errors and `helm install` should work correctly.

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

## Recover Operator when node fails

If the Operator pod is deployed on the node that fails, the pod will be rescheduled to a healthy node. However, the Operator will
not function properly because it has a leader election locking mechanism. See [here](https://github.com/operator-framework/operator-sdk/blob/master/doc/proposals/leader-for-life.md).

To make it work, the cluster admin will need to delete the lock by running
```
kubectl delete configmap pravega-operator-lock
```
After that, the new Operator pod will become the leader. If the node comes up later, the extra Operator pod will
be deleted by Deployment controller. 