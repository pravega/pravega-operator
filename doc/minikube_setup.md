# Minikube Setup

To setup minikube locally you can follow the steps mentioned [here](https://github.com/pravega/pravega/wiki/Kubernetes-Based-System-Test-Framework#minikube-setup).

Once minikube setup is complete, `minikube start` will create a minikube VM. However the resources allotted to this VM (i.e cpus , memory and disk space) might not be sufficient for a Pravega Cluster setup. It is recommended to create the minikube VM with 8 cpus, 50 gb memory and 50 gb storage (these values can be modified as per requirement) which can be done in the following way

```
minikube start --cpus=8 --memory=50000mb --disk-size=50000mb
```

## Resource Requirements

Here, we specify the optimal resources that we need to provide to each of the components in order to have a working Pravega setup.

### LongTermStorage
To setup LongTermStorage refer to [this](longtermstorage.md#long-term-storage). The `PersistentVolumeClaim` to consume the volume so created should be provisioned to consume 10Gi.

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
Create a single node Bookkeeper Cluster using the [BookKeeper Operator](https://github.com/pravega/bookkeeper-operator). In order to create a single node Bookkeeper Cluster, create the Bookkeeper Operator in [test mode](https://github.com/pravega/bookkeeper-operator/blob/master/doc/development.md#install-the-operator-in-test-mode). Modify the bookkeeper manifest and ensure that the following fields are present within its `spec`.

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
Finally create a Pravega Cluster comprising of a single SegmentStore and Controller replica using the Pravega Operator.The Operator can be run in `test mode` by enabling `testmode: true` in `values.yaml` file. Operator running in test mode skips minimum replica requirement checks on Pravega components. Modify the pravega manifest and ensure that the following fields are present within its `spec`.

```yaml
spec:
  pravega:
    controllerReplicas: 1
    segmentStoreReplicas: 1

    options:
      bookkeeper.ack.quorum.size: "1"
      bookkeeper.write.quorum.size: "1"
      bookkeeper.ensemble.size: "1"
```
