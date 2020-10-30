## LongTermStorage

The following LongTermStorage storage providers are supported:

- [Filesystem: NFS](#use-nfs-as-longtermstorage)
- [Filesystem: Google Filestore](#use-google-filestore-storage-as-longtermstorage)
- [S3: Dell EMC ECS](#use-dell-emc-ecs-as-longtermstorage)
- [HDFS](#use-hdfs-as-longtermstorage)

### Use NFS as LongTermStorage

The following example uses an NFS volume provisioned by the [NFS Server Provisioner](https://github.com/kubernetes/charts/tree/master/stable/nfs-server-provisioner) helm chart to provide LongTermStorage storage.

```
$ helm repo add stable https://kubernetes-charts.storage.googleapis.com
$ helm repo update
$ helm install stable/nfs-server-provisioner --generate-name
```

Note that the `nfs-server-provisioner` is a toy NFS server and is ONLY intended as a demo and should NOT be used for production deployments.

You can also connect to an existing NFS server by using [NFS Client Provisioner](https://github.com/helm/charts/tree/master/stable/nfs-client-provisioner).

```
helm install --set nfs.server=<address:x.x.x.x> --set nfs.path=</exported/path> --set storageClass.name=nfs --set nfs.mountOptions='{nolock,sec=sys,vers=4.0}' stable/nfs-client-provisioner --generate-name
```

Verify that the `nfs` storage class is now available.

```
$ kubectl get storageclass
NAME   PROVISIONER                                             AGE
nfs    cluster.local/elevated-leopard-nfs-server-provisioner   24s
...
```

Once the NFS provisioner is installed, you can create a `PersistentVolumeClaim` that will be used as LongTermStorage for Pravega. Create a `pvc.yaml` file with the following content.

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

### Use Google Filestore Storage as LongTermStorage

1. [Create a Google Filestore](https://console.cloud.google.com/filestore/instances).

> Refer to https://cloud.google.com/filestore/docs/accessing-fileshares for more information


2. Create a `pv.yaml` file with the `PersistentVolume` specification to provide LongTermStorage storage.

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

### Use Dell EMC ECS as LongTermStorage

Pravega can also use an S3-compatible storage backend such as [Dell EMC ECS](https://www.dellemc.com/sr-me/storage/ecs/index.htm) as LongTermStorage.

1. Create a file with the secret definition containing your access and secret keys.

    ```
    apiVersion: v1
    kind: Secret
    metadata:
      name: ecs-credentials
    type: Opaque
    stringData:
      ACCESS_KEY_ID: QWERTY@ecstestdrive.emc.com
      SECRET_KEY: 0123456789
    ```

2. Assuming that the file is named `ecs-credentials.yaml`.
    ```
    $ kubectl create -f ecs-credentials.yaml
    ```
3. Follow the [instructions to deploy Pravega manually](manual-installation.md#install-the-pravega-cluster-manually) and configure the LongTermStorage block in your `PravegaCluster` manifest with your ECS connection details and a reference to the secret above.
    ```
    ...
    spec:
      longtermStorage:
        ecs:
          configUri: http://10.247.10.52:9020?namespace=pravega
          bucket: "shared"
          prefix: "example"
          credentials: ecs-credentials
    ```

#### (Optional) ECS HTTPS/TLS Support on Kubernetes
Pravega connects ECS endpoint through OpenJDK based HTTP or HTTPS, so by default Pravega as an HTTPS client is configured to verify ECS server certificate.

The ECS server certificate, or its signing CA's certificate, must present in OpenJDK's Truststore, for Pravega to establish HTTPS/TLS connection with ECS endpoint.

Refer to the steps below to add ECS server certificate or CA's certificate into OpenJDK's Truststore:

1. Retrieve CA certificate or the server certificate as file, e.g. "ecs-certificate.pem".

2. Load the certificate into Kubernetes secret:
    ```
    kubectl create secret generic ecs-cert --from-file ./ecs-certificate.pem
    ```
    or create a file directly to contain the certificate content:
    ```
    apiVersion: v1
    kind: Secret
    metadata:
      name: "ecs-cert"
    type: Opaque
    data:
      ecs-certificate.pem: QmFnIEF0dH......JpYnV0ZLS0tLQo=
    ```
    Assuming the above file is named `ecs-tls.yaml`, then create secret using the above file.
    ```
    $ kubectl create -f ecs-tls.yaml
    ```

3. In Pravega manifest, add the secret name defined above into "tls/static/caBundle" section.
    ```
    ...
    kind: "PravegaCluster"
    metadata:
      name: "example"
    spec:
    tls:
      static:
        caBundle: "ecs-cert"
    ...
    longtermStorage:
        ecs:
          configUri: https://10.247.10.52:9021?namespace=pravega
          bucket: "shared"
          prefix: "example"
          credentials: ecs-credentials
    ```
4. Pravega operator then mounts caBundle onto folder "/etc/secret-volume/ca-bundle" in container.

5. Pravega Segmentstore container adds certificates found under "/etc/secret-volume/ca-bundle" into the default OpenJDK Truststore, in order to establish HTTPS/TLS connection with ECS.

#### Update ECS Credentials

There might be an operational need to update ECS credentials for a running Pravega cluster, where the following steps could be taken:

1. Modify Segmentstore configmap, find "EXTENDEDS3_CONFIGURI", and then replace "secretKey" and/or "identity" with new values
    ```
    $ kubectl edit configmap pravega-pravega-segmentstore
    ```

    ```
    ...
    EXTENDEDS3_BUCKET: shared
    EXTENDEDS3_CONFIGURI: https://10.243.86.64:9021?namespace=namespace%26identity=oldUser%26secretKey=oldPassword
    EXTENDEDS3_PREFIX: example
    ...
    ```
2. Delete all (running) Segmentstore pod(s) in one of the two approaches below:
    ```
    $ kubectl delete po -l component=pravega-segmentstore
    ```

    ```
    $ kubectl delete po pravega-pravega-segmentstore-1
    $ kubectl delete po pravega-pravega-segmentstore-2
    $ kubectl delete po pravega-pravega-segmentstore-3
    ...
    ```
    As StatefulSet, new Segementstore pods will be automatically created with the new ECS credentials, since the default upgrade strategy of Segmentstore is `OnDelete` instead of `RollingUpdate`.

    Since ECS supports grace period when both old and new credentials are accepted, Pravega service is technically uninterrupted during the above process.

### Use HDFS as LongTermStorage

Pravega can also use HDFS as the storage backend for LongTermStorage. The only requisite is that the HDFS backend must support Append operation.

Follow the [instructions to deploy Pravega manually](manual-installation.md#install-the-pravega-cluster-manually) and configure the LongTermStorage block in your `PravegaCluster` manifest with your HDFS connection details.

```
spec:
  longtermStorage:
    hdfs:
      uri: hdfs://10.28.2.14:8020/
      root: /example
      replicationFactor: 3
```
