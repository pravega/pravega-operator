# Pravega Operator

### Project status: alpha

The project is currently alpha. While no breaking API changes are currently planned, we reserve the right to address bugs and change the API before the project is declared stable.

## Table of Contents

 * [Overview](#overview)
 * [Requirements](#requirements)
 * [Usage](#usage)    
    * [Installation of the Operator](#install-the-operator)
    * [Deploy a sample Pravega Cluster](#deploy-a-sample-pravega-cluster)
    * [Uninstall the Pravega Cluster](#uninstall-the-pravega-cluster)
    * [Uninstall the Operator](#uninstall-the-operator)
 * [Configuration](#configuration)
    * [Use non-default service accounts](#use-non-default-service-accounts)
    * [Installing on a Custom Namespace with RBAC enabled](#installing-on-a-custom-namespace-with-rbac-enabled)
    * [Tier 2: Google Filestore Storage](#use-google-filestore-storage-as-tier-2)
    * [Tune Pravega Configurations](#tune-pravega-configuration)
 * [Development](#development)
    * [Build the Operator Image](#build-the-operator-image)
    * [Installation on GKE](#installation-on-google-kubernetes-engine)
    * [Direct Access to Cluster](#direct-access-to-the-cluster)
    * [Run the Operator Locally](#run-the-operator-locally)
* [Releases](#releases)
* [Troubleshooting](#troubleshooting)
    * [Helm Error: no available release name found](#helm-error-no-available-release-name-found)
    * [NFS volume mount failure: wrong fs type error](#nfs-volume-mount-failure-wrong-fs-type-error)
## Overview

[Pravega](http://pravega.io) is an open source distributed storage service implementing Streams. It offers Stream as the main primitive for the foundation of reliable storage systems: *a high-performance, durable, elastic, and unlimited append-only byte stream with strict ordering and consistency*.

The Pravega operator manages Pravega clusters deployed to Kubernetes and automates tasks related to operating a Pravega cluster.

- [x] Create and destroy a Pravega cluster
- [x] Resize cluster
- [ ] Rolling upgrades

> Note that unchecked features are in the roadmap but not available yet.

## Requirements

- Kubernetes 1.8+
- An existing Apache Zookeeper 3.5 cluster. This can be easily deployed using our [Zookeeper operator](https://github.com/pravega/zookeeper-operator)

## Usage

### Install the operator

> Note: If you are running on Google Kubernetes Engine (GKE), please [check this first](#installation-on-google-kubernetes-engine).

Run the following command to install the `PravegaCluster` custom resource definition (CRD), create the `pravega-operator` service account, roles, bindings, and the deploy the operator.

```
$ kubectl create -f deploy
```

Verify that the Pravega operator is running.

```
$ kubectl get deploy
NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
pravega-operator     1         1         1            1           17s
```

### Deploy a sample Pravega cluster

Pravega requires a long term storage provider known as Tier 2 storage. The following Tier 2 storage providers are supported:

- Filesystem (NFS)
- [Google Filestore](#using-google-filestore-storage-as-tier-2)
- [DellEMC ECS](https://www.dellemc.com/sr-me/storage/ecs/index.htm)
- HDFS (must support Append operation)

The following example uses an NFS volume provisioned by the [NFS Server Provisioner](https://github.com/kubernetes/charts/tree/master/stable/nfs-server-provisioner) helm chart to provide Tier 2 storage.

```
$ helm install stable/nfs-server-provisioner
```

Verify that the `nfs` storage class is now available.

```
$ kubectl get storageclass
NAME                 PROVISIONER                                             AGE
nfs                  cluster.local/elevated-leopard-nfs-server-provisioner   24s
...
```

> Note: This is ONLY intended as a demo and should NOT be used for production deployments.

Once the NFS server provisioner is installed, you can create a `PersistentVolumeClaim` that will be used as Tier 2 for Pravega. Create a `pvc.yaml` file with the following content.

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: pravega-tier2
spec:
  storageClassName: "nfs"
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 50Gi
```

```
$ kubectl create -f pvc.yaml
```
Use the following YAML template to install a small development Pravega Cluster (3 bookies, 1 controller, 3 segment stores). Create a `pravega.yaml` file with the following content.

```yaml
apiVersion: "pravega.pravega.io/v1alpha1"
kind: "PravegaCluster"
metadata:
  name: "pravega"
spec:
  zookeeperUri: [ZOOKEEPER_HOST]:2181

  bookkeeper:
    image:
      repository: pravega/bookkeeper
      tag: 0.3.2
      pullPolicy: IfNotPresent

    replicas: 3

    storage:
      ledgerVolumeClaimTemplate:
        accessModes: [ "ReadWriteOnce" ]
        storageClassName: "standard"
        resources:
          requests:
            storage: 10Gi

      journalVolumeClaimTemplate:
        accessModes: [ "ReadWriteOnce" ]
        storageClassName: "standard"
        resources:
          requests:
            storage: 10Gi

    autoRecovery: true

  pravega:
    controllerReplicas: 1
    segmentStoreReplicas: 3

    cacheVolumeClaimTemplate:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 20Gi

    image:
      repository: pravega/pravega
      tag: 0.3.2
      pullPolicy: IfNotPresent

    tier2:
      filesystem:
        persistentVolumeClaim:
          claimName: pravega-tier2
```

where:

- `[ZOOKEEPER_HOST]` is the host or IP address of your Zookeeper deployment.

Deploy the Pravega cluster.

```
$ kubectl create -f pravega.yaml
```

Verify that the cluster instances and its components are running.

```
$ kubectl get PravegaCluster
NAME      AGE
pravega   27s
```

```
$ kubectl get all -l pravega_cluster=pravega
NAME                                DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/pravega-pravega-controller   1         1         1            1           1m

NAME                                       DESIRED   CURRENT   READY     AGE
rs/pravega-pravega-controller-7489c9776d   1         1         1         1m

NAME                                DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/pravega-pravega-controller   1         1         1            1           1m

NAME                                       DESIRED   CURRENT   READY     AGE
rs/pravega-pravega-controller-7489c9776d   1         1         1         1m

NAME                                DESIRED   CURRENT   AGE
statefulsets/pravega-bookie         3         3         1m
statefulsets/pravega-segmentstore   3         3         1m

NAME                                             READY     STATUS    RESTARTS   AGE
po/pravega-bookie-0                              1/1       Running   0          1m
po/pravega-bookie-1                              1/1       Running   0          1m
po/pravega-bookie-2                              1/1       Running   0          1m
po/pravega-pravega-controller-7489c9776d-lcw9x   1/1       Running   0          1m
po/pravega-segmentstore-0                        1/1       Running   0          1m
po/pravega-segmentstore-1                        1/1       Running   0          1m
po/pravega-segmentstore-2                        1/1       Running   0          1m

NAME                             TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)              AGE
svc/pravega-bookie-headless      ClusterIP   None           <none>        3181/TCP             1m
svc/pravega-pravega-controller   ClusterIP   10.3.255.239   <none>        10080/TCP,9090/TCP   1m
```

A `PravegaCluster` instance is only accessible WITHIN the cluster (i.e. no outside access is allowed) using the following endpoint in
the PravegaClient.

```
tcp://<cluster-name>-pravega-controller.<namespace>:9090
```

The `REST` management interface is available at:

```
http://<cluster-name>-pravega-controller.<namespace>:10080/
```

[Check this](#direct-access-to-the-cluster) to enable direct access to the cluster for development purposes.

### Uninstall the Pravega cluster

```
$ kubectl delete -f pravega.yaml
$ kubectl delete -f pvc.yaml
```

### Uninstall the operator

> Note that the Pravega clusters managed by the Pravega operator will NOT be deleted even if the operator is uninstalled.

To delete all clusters, delete all cluster CR objects before uninstalling the operator.

```
$ kubectl delete -f deploy
```

## Configuration

### Use non-default service accounts

You can optionally configure non-default service accounts for the Bookkeeper, Pravega Controller, and Pravega Segment Store pods.

For BookKeeper, set the `serviceAccountName` field under the `bookkeeper` block.

```
...
spec:
  bookkeeper:
    serviceAccountName: bk-service-account
...
```

For Pravega, set the `controllerServiceAccountName` and `segmentStoreServiceAccountName` fields under the `pravega` block.

```
...
spec:
  pravega:
    controllerServiceAccountName: ctrl-service-account
    segmentStoreServiceAccountName: ss-service-account
...
```

### Installing on a Custom Namespace with RBAC enabled

Create the namespace.

```
$ kubectl create namespace pravega-io
```

Update the namespace configured in the `deploy/role_binding.yaml` file.

```
$ sed -i -e 's/namespace: default/namespace: pravega-io/g' deploy/role_binding.yaml
```

Apply the changes.

```
$ kubectl -n pravega-io apply -f deploy
```

Note that the Pravega operator only monitors the `PravegaCluster` resources which are created in the same namespace, `pravega-io` in this example. Therefore, before creating a `PravegaCluster` resource, make sure an operator exists in that namespace.

```
$ kubectl -n pravega-io create -f example/cr.yaml
```

```
$ kubectl -n pravega-io get pravegaclusters
NAME      AGE
pravega   28m
```

```
$ kubectl -n pravega-io get pods -l pravega_cluster=pravega
NAME                                          READY     STATUS    RESTARTS   AGE
pravega-bookie-0                              1/1       Running   0          29m
pravega-bookie-1                              1/1       Running   0          29m
pravega-bookie-2                              1/1       Running   0          29m
pravega-pravega-controller-6c54fdcdf5-947nw   1/1       Running   0          29m
pravega-pravega-segmentstore-0                1/1       Running   0          29m
pravega-pravega-segmentstore-1                1/1       Running   0          29m
pravega-pravega-segmentstore-2                1/1       Running   0          29m
```

### Use Google Filestore Storage as Tier 2

1. [Create a Google Filestore](https://console.cloud.google.com/filestore/instances).

> Refer to https://cloud.google.com/filestore/docs/accessing-fileshares for more information


2. Create a `pv.yaml` file with the `PersistentVolume` specification to provide Tier 2 storage.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pravega-volume
spec:
  capacity:
    storage: 1T
  accessModes:
  - ReadWriteMany
  nfs:
    path: /[FILESHARE]
    server: [IP_ADDRESS]
```

where:

- `[FILESHARE]` is the name of the fileshare on the Cloud Filestore instance (e.g. `vol1`)
- `[IP_ADDRESS]` is the IP address for the Cloud Filestore instance (e.g. `10.123.189.202`)


3. Deploy the `PersistentVolume` specification.

```
$ kubectl create -f pv.yaml
```

4. Create and deploy a `PersistentVolumeClaim` to consume the volume created.

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: pravega-tier2
spec:
  storageClassName: ""
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 50Gi
```

```
$ kubectl create -f pvc.yaml
```

Use the same `pravega.yaml` above to deploy the Pravega cluster.


### Tune Pravega configuration

Pravega has many configuration options for setting up metrics, tuning, etc. The available options can be found
[here](https://github.com/pravega/pravega/blob/master/config/config.properties) and are
expressed through the pravega/options part of the resource specification. All values must be expressed as Strings.

```yaml
...
spec:
  pravega:
    options:
      metrics.enableStatistics: "true"
      metrics.statsdHost: "telegraph.default"
      metrics.statsdPort: "8125"
...
```

## Development

### Build the operator image

Requirements:
  - Go 1.10+

Use the `make` command to build the Pravega operator image.

```
$ make build
```
That will generate a Docker image with the format
`<latest_release_tag>-<number_of_commits_after_the_release>` (it will append-dirty if there are uncommitted changes). The image will also be tagged as `latest`.

Example image after running `make build`.

The Pravega operator image will be available in your Docker environment.

```
$ docker images pravega/pravega-operator

REPOSITORY                  TAG            IMAGE ID      CREATED          SIZE        

pravega/pravega-operator    0.1.1-3-dirty  2b2d5bcbedf5  10 minutes ago   41.7MB    

pravega/pravega-operator    latest         2b2d5bcbedf5  10 minutes ago   41.7MB

```

Optionally push it to a Docker registry.

```
docker tag pravega/pravega-operator [REGISTRY_HOST]:[REGISTRY_PORT]/pravega/pravega-operator
docker push [REGISTRY_HOST]:[REGISTRY_PORT]/pravega/pravega-operator
```

where:

- `[REGISTRY_HOST]` is your registry host or IP (e.g. `registry.example.com`)
- `[REGISTRY_PORT]` is your registry port (e.g. `5000`)

### Installation on Google Kubernetes Engine

The Operator requires elevated privileges in order to watch for the custom resources.

According to Google Container Engine docs:

> Ensure the creation of RoleBinding as it grants all the permissions included in the role that we want to create. Because of the way Container Engine checks permissions when we create a Role or ClusterRole.
>
> An example workaround is to create a RoleBinding that gives your Google identity a cluster-admin role before attempting to create additional Role or ClusterRole permissions.
>
> This is a known issue in the Beta release of Role-Based Access Control in Kubernetes and Container Engine version 1.6.

On GKE, the following command must be run before installing the operator, replacing the user with your own details.

```
$ kubectl create clusterrolebinding your-user-cluster-admin-binding --clusterrole=cluster-admin --user=your.google.cloud.email@example.org
```

### Direct access to the cluster

For debugging and development you might want to access the Pravega cluster directly. For example, if you created the cluster with name `pravega` in the `default` namespace you can forward ports of the Pravega controller pod with name `pravega-pravega-controller-68657d67cd-w5x8b` as follows:

```
$ kubectl port-forward -n default pravega-pravega-controller-68657d67cd-w5x8b 9090:9090 10080:10080
```
## Run the operator locally

You can run the operator locally to help with development, testing, and debugging tasks.

The following command will run the operator locally with the default Kubernetes config file present at `$HOME/.kube/config`. Use the `--kubeconfig` flag to provide a different path.

```
$ operator-sdk up local
```
## Releases  

The latest Pravega releases can be found on the [Github Release](https://github.com/pravega/pravega-operator/releases) project page.

## Troubleshooting

### Helm Error: no available release name found

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

### NFS volume mount failure: wrong fs type error

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
