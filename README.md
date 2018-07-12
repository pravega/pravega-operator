# Pravega Operator

Pravega Kubernetes Operator


# Zookeeper Operator
>**This operator is in WIP state and subject to (breaking) changes.**

This Operator runs a Zookeeper 3.5 cluster, and uses Zookeeper dynamic reconfiguration to handle node membership.

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

### Install the Kubernetes resources

```bash
# Create Operator deployment, Roles, Service Account, and Custom Resource Definition for
#   a Pravega cluster.
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

### The PravegaCluster Custom Resource

#### Requirements

Each Pravega cluster requires a running Zookeeper instance.  This must be deployed before deploying the Pravega Cluster
resource and the `example/cr.yaml` file updated with the correct `zookeeperUri`.

#### Deployments
This will deploy a small (3 Bookies, 1 controller, 1 segmentstore) development instance to your Kuberentes cluster using 
a standard PVC for Tier2 Storage (simulating an NFS mount)

With this YAML template you can install a small development Pravega Cluster easily into your Kubernetes cluster:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pravega-tier2
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
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
    segmentStoreReplicas: 1

    image:
      repository: pravega/pravega
      tag: 0.3.0
      pullPolicy: IfNotPresent

    tier2:
      filesystem:
        persistentVolumeClaim:
          claimName: pravega-tier2
```
