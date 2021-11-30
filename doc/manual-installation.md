# Manual Installation

* [Install the Operator manually](#install-the-operator-manually)
* [Set up LongTermStorage](#set-up-longtermstorage)
* [Install the Pravega cluster manually](#install-the-pravega-cluster-manually)
* [Uninstall the Pravega Cluster manually](#uninstall-the-pravega-cluster-manually)
* [Uninstall the Operator manually](#uninstall-the-operator-manually)

### Install the Operator manually

> Note: If you are running on Google Kubernetes Engine (GKE), please [check this first](development.md#installation-on-google-kubernetes-engine).

In case you don't have cert-manager v0.15.0+, install it from the following link :-

https://cert-manager.io/docs/installation/kubernetes/

Install the issuer and certificate
```
$ kubectl create -f deploy/certificate.yaml
```
Install the webhook
```
$ kubectl create -f deploy/webhook.yaml  
```
Register the Pravega cluster custom resource definition (CRD).
```
$ kubectl create -f deploy/crds/pravega.pravega.io_pravegaclusters_crd.yaml
```
Create the operator role, role binding and service account.
```
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
$ kubectl create -f deploy/service_account.yaml
$ kubectl create -f deploy/version_map.yaml
```
Install the operator.
```
$ kubectl create -f deploy/operator.yaml  
```

### Deploying in Test Mode
 We can enable test mode on operator by passing an argument `-test` in `operator.yaml` file. Operator running in test mode skips minimum replica requirement checks on Pravega components. Test mode provides a bare minimum setup and is not recommended to be used in production environments.

```
containers:
  - name: pravega-operator
    image: pravega/pravega-operator:0.5.0-rc1
    ports:
    - containerPort: 60000
      name: metrics
    command:
    - pravega-operator
    imagePullPolicy: Always
    args: [-test]
```
### Set up LongTermStorage

Pravega requires a long term storage provider known as longtermStorage.

Check out the available [options for longtermStorage](longtermstorage.md) and how to configure it.

In this example we are going to use a `pravega-tier2` PVC using [NFS as the storage backend](longtermstorage.md#use-nfs-as-longtermstorage).

### Install the Pravega cluster manually

Deploying the Pravega cluster.

```
$ kubectl create -f deploy/crds/pravega-cr.yaml
```

Verify that the cluster instances and its components are being created.

```
$ kubectl get PravegaCluster
NAME      VERSION   DESIRED VERSION   DESIRED MEMBERS   READY MEMBERS   AGE
pravega   0.10.0    0.10.0            4                 0               58s
```

**Note:** If we are installing pravega version 0.9.0 or above using operator version 0.5.1 or below, add the below JVM options for controller and segmentstore in addition to the current JVM options.
```
segmentStoreJVMOptions: ["-XX:+UseContainerSupport","-XX:+IgnoreUnrecognizedVMOptions"]

controllerjvmOptions: ["-XX:+UseContainerSupport","-XX:+IgnoreUnrecognizedVMOptions"]
```

### Uninstall the Pravega cluster manually

```
$ kubectl delete -f deploy/crds/pravega-cr.yaml
$ kubectl delete pvc pravega-tier2
```

### Uninstall the Operator manually

> Note that the Pravega clusters managed by the Pravega operator will NOT be deleted even if the operator is uninstalled.

To delete all clusters, delete all cluster CR objects before uninstalling the operator.

```
$ kubectl delete -f deploy
```
