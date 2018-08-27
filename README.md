# Pravega Operator
Pravega Operator, the easiest and best way to deploy Pravega in Kubernetes. Pravega operator is built on top of a common set of Kubernetes APIs by providing a great automation experience. The Pravega operator performs packaging, deploying and managing the Kubernetes application.The Pravega operator is developed using the CoreOS operator framework [SDK](https://github.com/operator-framework/operator-sdk).

**Note:** The Development of Pravega operator is a WIP and it is expected that breaking changes to the API will be made in the upcoming releases.

## key Features of Pravega Operator

The Pravega Operator directly addresses the challenges of running Pravega on Kubernetes, and will offer the following features across all Platform components:

#### Manages the configuration for Pravega in Kubernetes:

* Automatic deployment of Pravega clusters across various platforms. (It eliminates the burden of using configmaps and appropriate config values for deploying the Pravega clusters.) 
* Automatic configurations are performed for zookeeper deployment (zookeeper.connect, log.dirs).
* The Pravega Operator automatically handles the ordinal index that the Kubernetes StatefulSet assigns to each Pravega pod while deploying the Pravega cluster.
* Automatic enabling of SASL authentication mechanisms.
* Persistent storage is maintained by mounting persistent volumes are mounted for every pod and managed well across failures using kubectl commands.
* Automatic configuration fields of Pravega pod like Memory, CPU, and Disks is performed.
* The entire complexity of running the stateful data service like Pravega in Kubernetes can be easily overcome by the Pravega operator.
* Automatic Monitoring is performed by using metric for alerting.

# Deployment of Pravega Operator

The Pravega Operator Provisions a [Pravega Cluster](https://github.com/pravega/pravega).
The operator is developed using the [operator-sdk](https://github.com/operator-framework/operator-sdk).


## Build Requirements:
 
 Install the Operator SDK from the following: https://github.com/operator-framework/operator-sdk#quick-start

### Usage:

```bash
mkdir -p $GOPATH/src/github.com/pravega
cd $GOPATH/src/github.com/pravega
git clone git@github.com:pravega/pravega-operator.git
cd pravega-operator
```

#### Get the operator Docker image

##### a. Build the image yourself

```bash
operator-sdk build pravega/pravega-operator
docker tag pravega/pravega-operator ${your-operator-image-tag}:latest
docker push ${your-operator-image-tag}:latest
```

##### b. Use the image from Docker Hub

```bash
# No additional steps are required to use the image from the Docker Hub.
```

#### Install the Operator

The operator and required resources can be installed using the `yaml` files available in the deploy directory:
```bash
$ kubectl apply -f deploy
```

View the pravega-operator Pod by using the following command:
```
$ kubectl get pod
NAME                                  READY     STATUS              RESTARTS   AGE
pravega-operator-6787869796-mxqjv      1/1       Running             0          1m
```

## Installation on Google GKE
The Operator requires elevated privileges in order to watch for the custom resources.  

According to Google Container Engine docs:
>Ensure the creation of RoleBinding as it grants all the permissions included in the role that we want to create. Because of the way Container Engine checks permissions when we create a Role or ClusterRole. 
> 
> An example workaround is to create a RoleBinding that gives your Google identity a cluster-admin role before attempting to create additional Role or ClusterRole permissions.
> 
> This is a known issue in the Beta release of Role-Based Access Control in Kubernetes and Container Engine version 1.6.

On Google GKE the following command must be run before installing the operator, replacing the user with your own details.

```kubectl create clusterrolebinding your-user-cluster-admin-binding --clusterrole=cluster-admin --user=your.google.cloud.email@example.org```

# Prerequisite for provisioning Pravega Cluster

The following components must exist before the operator can be used to provision a PravegaCluster.

## Zookeeper

Pravega requires an existing Apache Zookeeper 3.5 cluster. This can be easily deployed using the [Pravega Zookeeper operator](https://github.com/pravega/zookeeper-operator). 

The ZookeeperURI for the cluster is provided as part of the PravegaCluster resource.

The operator itself is built with the: https://github.com/operator-framework/operator-sdk.

## Build Requirements:

Install the Operator SDK: https://github.com/operator-framework/operator-sdk#quick-start

### Usage:

```bash
mkdir -p $GOPATH/src/github.com/pravega
cd $GOPATH/src/github.com/pravega
git clone git@github.com:pravega/zookeeper-operator.git
cd zookeeper-operator
```
### Get the operator Docker image

#### a. Build the image yourself

```bash
operator-sdk build pravega/zookeeper-operator
docker tag pravega/zookeeper-operator ${your-operator-image-tag}:latest
docker push ${your-operator-image-tag}:latest
```
#### b. Use the image from Docker Hub
```
# No additional steps are required to use the image from Docker Hub.
```

### Install the Kubernetes Resources

Ensure to enable cluster role bindings if we are running on GKE:
```bash

$ kubectl create clusterrolebinding your-user-cluster-admin-binding --clusterrole=cluster-admin --user=<your.google.cloud.email@example.org>
```
Install the operator components  and create Operator deployment, Roles, Service Account, and Custom Resource Definition for
a Zookeeper cluster as follows:

```bash
$ kubectl apply -f deploy
```
View the zookeeper-operator Pod using the following commands:
```bash
$ kubectl get pod
NAME                                  READY     STATUS              RESTARTS   AGE
zookeeper-operator-5c7b8cfd85-ttb5g   1/1       Running             0          5m
```

### The Zookeeper Custom Resource

Using the follwoing `YAML` template, install a 3 node Zookeeper Cluster easily into the Kubernetes cluster:
```bash
apiVersion: "zookeeper.pravega.io/v1beta1"
kind: "ZookeeperCluster"
metadata:
  name: "example"
spec:
  size: 3
```
View the cluster and its components using the following commands:

```bash
$ kubectl get zk
NAME      AGE
example   2s

$ kubectl get all -l app=example
NAME            READY     STATUS              RESTARTS   AGE
pod/example-0   1/1       Running             0          51m
pod/example-1   1/1       Running             0          55m
pod/example-2   1/1       Running             0          58m

NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/example-client     ClusterIP   x.x.x.x         <none>        2181/TCP            51m
service/example-headless   ClusterIP   None            <none>        2888/TCP,3888/TCP   51m

NAME                       DESIRED   CURRENT   AGE
statefulset.apps/example   3         3         58m

# There are a few other things here, like a configmap, poddisruptionbudget, etc...
```


**Note:** The Zookeeper instance can be shared between multiple PravegaCluster instances.

### Tier2 Storage
Pravega requires a long term storage provider known as Tier2 storage.  Several Tier2 storage providers are supported:

- Filesystem (NFS)
- DellEMC ECS
- HDFS (must support Append operation)

An instance of a Pravega cluster supports only one type of Tier2 storage which is configured during cluster provisioning and
cannot be changed once provisioned.  The required provider is configured using the `Pravega/Tier2` section of the 
PravegaCluster resource.  You must provide one and only one type of storage configuration.

#### Example

#### NFS Storage

The following example uses an NFS PVC provisioned by the [NFS Server Provisioner](https://github.com/kubernetes/charts/tree/master/stable/nfs-server-provisioner) 
helm chart to provide Tier2 storage:

```
helm install stable/nfs-server-provisioner
```

**Note:** This is ONLY intended as a demo and should NOT be used for production deployments.

#### Deployment

Using the following `YAML` template we can easily install a small development Pravega Cluster (3 Bookies, 1 controller, 3 segmentstore)
in our Kubernetes cluster. The cluster will be provisioned in to the same namespace as the resource.

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

View the cluster instance and its components using the following command:

```
$ kubectl get PravegaCluster
NAME      AGE
example   2h

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
