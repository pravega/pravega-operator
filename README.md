# Pravega Operator
>**This operator is in WIP state and subject to (breaking) changes.**

This Operator Provisions a [Pravega Cluster](https://github.com/pravega/pravega).

The operator itself is built with the: https://github.com/operator-framework/operator-sdk

## Build Requirements:
 - Install the Operator SDK first: https://github.com/operator-framework/operator-sdk#quick-start

## Usage:

```bash
mkdir -p $GOPATH/src/github.com/pravega
cd $GOPATH/src/github.com/pravega
git clone git@github.com:pravega/pravega-operator.git
cd pravega-operator
```

### Get the operator Docker image

#### a. Build the image yourself

```bash
operator-sdk build pravega/pravega-operator
docker tag pravega/pravega-operator ${your-operator-image-tag}:latest
docker push ${your-operator-image-tag}:latest
```

#### b. Use the image from Docker Hub

```bash
# No addition steps needed
```

### Install the Operator

The operator and required resources can be installed using the yaml files in the deploy directory:
```bash
$ kubectl apply -f deploy

# View the pravega-operator Pod
$ kubectl get pod
NAME                                  READY     STATUS              RESTARTS   AGE
pravega-operator-6787869796-mxqjv      1/1       Running             0          1m
```

#### Installation on Google GKE
The Operator requires elevated privileges in order to watch for the custom resources.  

According to Google Container Engine docs:
>Because of the way Container Engine checks permissions when you create a Role or ClusterRole, you must first create a RoleBinding that grants you all of the permissions included in the role you want to create.
>
> An example workaround is to create a RoleBinding that gives your Google identity a cluster-admin role before attempting to create additional Role or ClusterRole permissions.
>
> This is a known issue in the Beta release of Role-Based Access Control in Kubernetes and Container Engine version 1.6.

On Google GKE the following command must be run before installing the operator, replacing the user with your own details.

```kubectl create clusterrolebinding your-user-cluster-admin-binding --clusterrole=cluster-admin --user=your.google.cloud.email@example.org```

### Requirements

There are several required components that must exist before the operator can be used to provision a PravegaCluster:

#### Zookeeper

Pravega requires an existing Apache Zookeeper 3.5 cluster .  Which can easily be deployed using the [Pravega Zookeeper operator](https://github.com/pravega/zookeeper-operator).  
The ZookeeperURI for the cluster is provided as part of the PravegaCluster resource.

Note that the Zookeeper instance can be shared between multiple PravegaCluster instances.

#### Tier2 Storage
Pravega requires a long term storage provider known as Tier2 storage.  Several Tier2 storage providers are supported:

- Filesystem (NFS)
  - Google Filestore (please refer to https://console.cloud.google.com/filestore)
- DellEMC ECS
- HDFS (must support Append operation)
  
An instance of a Pravega cluster supports only one type of Tier2 storage which is configured during cluster provisioning and
cannot be changed once provisioned.  The required provider is configured using the `Pravega/Tier2` section of the 
PravegaCluster resource.  You must provide one and only one type of storage configuration.

### Example

#### NFS Storage

The following example uses an NFS PVC provisioned by the [NFS Server Provisioner](https://github.com/kubernetes/charts/tree/master/stable/nfs-server-provisioner) 
helm chart to provide Tier2 storage:

```
helm install stable/nfs-server-provisioner
```

Note that this is ONLY intended as a demo and should NOT be used for production deployments.

#### Deployment

With this YAML template you can install a small development Pravega Cluster (3 Bookies, 1 controller, 3 segmentstore) easily 
into your Kubernetes cluster. The cluster will be provisioned into the same namespace as the operator.

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
---
apiVersion: "pravega.pravega.io/v1alpha1"
kind: "PravegaCluster"
metadata:
  name: "example"
spec:
  zookeeperUri: example-client:2181

  bookkeeper:
    image:
      repository: pravega/bookkeeper
      tag: 0.3.0
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
      tag: 0.3.0
      pullPolicy: IfNotPresent

    tier2:
      filesystem:
        persistentVolumeClaim:
          claimName: pravega-tier2
```

#### NFS: Google Filestore Storage
Create a Persistent Volume (refer to https://cloud.google.com/filestore/docs/accessing-fileshares)  to provide Tier2 storage:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pravega
spec:
  capacity:
    storage: 1T
  accessModes:
  - ReadWriteMany
  nfs:
    path: /vol1
    server: 10.123.189.202
```

Deploy the persistent volume specification:
```kubectl create -f pv.yaml```

Note: the "10.123.189.202:/vol1" is the Filestore that is created previously, and this is ONLY intended as a demo and should NOT be used for production deployments.

#### Deployment

With this YAML template you can install a small development Pravega Cluster (3 Bookies, 1 controller, 3 segmentstore) easily 
into your Kubernetes cluster. The cluster will be provisioned into the same namespace as the operator.

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
---
apiVersion: "pravega.pravega.io/v1alpha1"
kind: "PravegaCluster"
metadata:
  name: "example"
spec:
  zookeeperUri: example-client:2181

  bookkeeper:
    image:
      repository: pravega/bookkeeper
      tag: 0.3.0
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
      tag: 0.3.0
      pullPolicy: IfNotPresent

    tier2:
      filesystem:
        persistentVolumeClaim:
          claimName: pravega-tier2
```

After creating, you can view the cluster:

```
# View the cluster instance
$ kubectl get PravegaCluster
NAME      AGE
example   2h

# View what it's made of
$ kubectl get all -l pravega_cluster=example
NAME                                              READY     STATUS    RESTARTS   AGE
pod/example-bookie-0                              1/1       Running   0          2h
pod/example-bookie-1                              1/1       Running   0          2h
pod/example-bookie-2                              1/1       Running   0          2h
pod/example-pravega-controller-6f58c4f464-2jg8f   1/1       Running   0          2h
pod/example-segmentstore-0                        1/1       Running   0          2h
pod/example-segmentstore-1                        1/1       Running   0          2h
pod/example-segmentstore-2                        1/1       Running   0          2h

NAME                                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)              AGE
service/example-pravega-controller   ClusterIP   10.39.254.134   <none>        10080/TCP,9090/TCP   2h

NAME                                               DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/example-pravega-controller   1         1         1            1           2h

NAME                                                          DESIRED   CURRENT   READY     AGE
replicaset.extensions/example-pravega-controller-6f58c4f464   1         1         1         2h

NAME                                         DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/example-pravega-controller   1         1         1            1           2h

NAME                                                    DESIRED   CURRENT   READY     AGE
replicaset.apps/example-pravega-controller-6f58c4f464   1         1         1         2h

NAME                                    DESIRED   CURRENT   AGE
statefulset.apps/example-bookie         3         3         2h
statefulset.apps/example-segmentstore   3         3         2h

# There are a few other things here, like a configmap, etc...

```

#### Using The Pravega Instance

A PravegaCluster instance is only accessible WITHIN the cluster (i.e. no outside access) using the following endpoint in 
the PravegaClient:

```
tcp://<cluster-name>-pravega-controller.<namespace>:9090
```

The REST management interface is available at:
```
http://<cluster-name>-pravega-controller.<namespace>:10080/
```

#### Pravega Configuration

Pravega has many configuration options for setting up metrics, tuning, etc.  The available options can be found
[here](https://github.com/pravega/pravega/blob/3f5b65084ae17e74c8ef8e6a40e78e61fa98737b/config/config.properties) and are
expressed through the pravega/options part of the resource specification.  All values must be expressed as Strings.

```
spec:
    pravega:
        options:
          metrics.enableStatistics: "true"
          metrics.statsdHost: "telegraph.default"
          metrics.statsdPort: "8125"
```
