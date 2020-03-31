# Minikube Setup

To setup minikube locally you can follow the steps mentioned [here](https://github.com/pravega/pravega/wiki/Kubernetes-Based-System-Test-Framework#minikube-setup).

Once minikube setup is complete, `minikube start` will create a minikube VM. However the resources allotted to this VM (i.e cpus , memory and disk space) might not be sufficient for a Pravega Cluster setup. It is recommended to create the minikube VM with 8 cpus, 50 gb memory and 50 gb storage (these values can be modified as per requirement) which can be done in the following way

```
minikube start --cpus=8 --memory=50000mb --disk-size=50000mb
```

## Resource Requirements

Here, we specify the optimal resources that we need to provide to each of the components in order to have a working Pravega setup.

### Tier 2
To setup Tier 2 refer to [this](tier2.md#tier-2-storage). The `PersistentVolumeClaim` to consume the volume so created should be provisioned to consume 10Gi.

```yaml
  resources:
    requests:
      storage: 10Gi
```

### Zookeeper
Create a single node Zookeeper Cluster using the [Zookeeper Operator](https://github.com/pravega/zookeeper-operator). Modify the zookeeper manifest and ensure that the following fields are present within its `spec`.

```yaml
spec:
  replicas: 1
  persistence:
    reclaimPolicy: Delete
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 10Gi
```

### Bookkeeper
Create a single node Bookkeeper Cluster using the [BookKeeper Operator](https://github.com/pravega/bookkeeper-operator). Modify the bookkeeper manifest and ensure that the following fields are present within its `spec`.

```yaml
spec:
  replicas: 1
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

      indexVolumeClaimTemplate:
        accessModes: [ "ReadWriteOnce" ]
        storageClassName: "standard"
        resources:
          requests:
            storage: 10Gi
```

### Pravega
Finally create a Pravega Cluster comprising of a single SegmentStore and Controller replica using the Pravega Operator. Modify the pravega manifest and ensure that the following fields are present within its `spec`.

```yaml
spec:
  pravega:
    controllerReplicas: 1
    segmentStoreReplicas: 1

    options:
      bookkeeper.bkAckQuorumSize: "1"
      bookkeeper.bkWriteQuorumSize: "1"
      bookkeeper.bkEnsembleSize: "1"
```
