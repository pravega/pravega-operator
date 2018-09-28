# Pravega Operator

### Project status: alpha

The project is currently alpha. While no breaking API changes are currently planned, we reserve the right to address bugs and change the API before the project is declared stable.

### Overview

[Pravega](http://pravega.io) is an open source distributed storage service implementing Streams. It offers Stream as the main primitive for the foundation of reliable storage systems: *a high-performance, durable, elastic, and unlimited append-only byte stream with strict ordering and consistency*.

The Pravega operator manages Pravega clusters deployed to Kubernetes and automates tasks related to operating a Pravega cluster.

- [x] Create and destroy a Pravega cluster
- [x] Resize cluster
- [ ] Rolling upgrades

> Note that unchecked features are in the roadmap but not available yet.

## Requirements

- Kubernetes 1.8+
- An existing Apache Zookeeper 3.5 cluster. This can be easily deployed using the following [Zookeeper operator](https://github.com/Nuance-Mobility/zookeeper-operator)

## Usage

### Install the operator

> Note: if you are running on Google Kubernetes Engine (GKE), please [check this first](#installation-on-google-kubernetes-engine).

Register the `PravegaCluster` custom resource definition (CRD).

```
$ kubectl create -f deploy/crd.yaml
```

Create the operator role and role binding.

```
$ kubectl create -f deploy/rbac.yaml
```

Deploy the Pravega operator.

```
$ kubectl create -f deploy/operator.yaml
```

Verify that the Pravega operator is running.

```
$ kubectl get deploy
NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
pravega-operator     1         1         1            1           17s
zookeeper-operator   1         1         1            1           12m
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

## Development

### Build the operator image

Requirements:
  - Go 1.10+
  - [Operator SDK](https://github.com/operator-framework/operator-sdk#quick-start)

Use the `operator-sdk` command to build the Pravega operator image.

```
$ operator-sdk build pravega/pravega-operator
```

The Pravega operator image will be available in your Docker environment.

```
$ docker images pravega/pravega-operator
REPOSITORY                 TAG                 IMAGE ID            CREATED             SIZE
pravega/pravega-operator   latest              625dab6fe470        54 seconds ago      37.2MB
```

Optionally push it to a Docker registry.

```
docker tag pravega/pravega-operator [REGISTRY_HOST]:[REGISTRY_PORT]/pravega/pravega-operator
docker push [REGISTRY_HOST]:[REGISTRY_PORT]/pravega/pravega-operator
```

where:

- `[REGISTRY_HOST]` is your registry host or IP (e.g. `registry.example.com`)
- `[REGISTRY_PORT]` is your registry port (e.g. `5000`)

### Using Google Filestore Storage as Tier 2

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


### Tuning Pravega configuration

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
