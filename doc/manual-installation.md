## Manual installation

### Install the Operator manually

> Note: If you are running on Google Kubernetes Engine (GKE), please [check this first](#installation-on-google-kubernetes-engine).

Register the Pravega cluster custom resource definition (CRD).

```
$ kubectl create -f deploy/crd.yaml
```

Create the operator role, role binding and service account.

```
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
$ kubectl create -f deploy/service_account.yaml
```

Install the operator.

```
$ kubectl create -f deploy/operator.yaml
```

### Install the Pravega cluster manually

Once the operator is installed, you can use the following YAML template to install a small development Pravega Cluster (3 Bookies, 1 Controller, 3 Segment Stores). Create a `pravega.yaml` file with the following content.

```yaml
apiVersion: "pravega.pravega.io/v1alpha1"
kind: "PravegaCluster"
metadata:
  name: "example"
spec:
  version: 0.4.0
  zookeeperUri: [ZOOKEEPER_HOST]:2181

  bookkeeper:
    replicas: 3
    image:
      repository: pravega/bookkeeper
    autoRecovery: true

  pravega:
    controllerReplicas: 1
    segmentStoreReplicas: 3
    image:
      repository: pravega/pravega
    tier2:
      filesystem:
        persistentVolumeClaim:
          claimName: pravega-tier2
```

where:

- `[ZOOKEEPER_HOST]` is the host or IP address of your Zookeeper deployment.

Check out other sample CR files in the [`example`](../example) directory.

Deploy the Pravega cluster.

```
$ kubectl create -f pravega.yaml
```

Verify that the cluster instances and its components are being created.

```
$ kubectl get PravegaCluster
NAME      VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
example   0.4.0     7                 0               25s
```
