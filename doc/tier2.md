## Tier 2 Storage

The following Tier 2 storage providers are supported:

- [Filesystem: NFS](#use-nfs-as-tier-2)
- [Filesystem: Google Filestore](#use-google-filestore-storage-as-tier-2)
- [S3: Dell EMC ECS](#use-dell-emc-ecs-as-tier-2)
- [HDFS](#use-hdfs-as-tier-2)

### Use NFS as Tier 2

The following example uses an NFS volume provisioned by the [NFS Server Provisioner](https://github.com/kubernetes/charts/tree/master/stable/nfs-server-provisioner) helm chart to provide Tier 2 storage.

```
$ helm install stable/nfs-server-provisioner
```

Note that the `nfs-server-provisioner` is a toy NFS server and is ONLY intended as a demo and should NOT be used for production deployments.

You can also connect to an existing NFS server by using [NFS Client Provisioner](https://github.com/helm/charts/tree/master/stable/nfs-client-provisioner).

```
helm install --set nfs.server=<address:x.x.x.x> --set nfs.path=</exported/path> --set storageClass.name=nfs --set nfs.mountOptions='{nolock,sec=sys,vers=4.0}' stable/nfs-client-provisioner
```

Verify that the `nfs` storage class is now available.

```
$ kubectl get storageclass
NAME   PROVISIONER                                             AGE
nfs    cluster.local/elevated-leopard-nfs-server-provisioner   24s
...
```

Once the NFS provisioner is installed, you can create a `PersistentVolumeClaim` that will be used as Tier 2 for Pravega. Create a `pvc.yaml` file with the following content.

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

### Use Dell EMC ECS as Tier 2

Pravega can also use an S3-compatible storage backend such as [Dell EMC ECS](https://www.dellemc.com/sr-me/storage/ecs/index.htm) as Tier 2.

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
3. Follow the [instructions to deploy Pravega manually](manual-installation.md#install-the-pravega-cluster-manually) and configure the Tier 2 block in your `PravegaCluster` manifest with your ECS connection details and a reference to the secret above.
    ```
    ...
    spec:
    tier2:
        ecs:
          configUri: http://10.247.10.52:9020?namespace=pravega
          bucket: "shared"
          prefix: "example"
          credentials: ecs-credentials
    ```

#### (Optional) ECS TLS Support
In case the ECS is secured using self-signed TLS/SSL certificate, Pravega must trust the certificate in order to establish secured connection with it.

This is done by adding the ECS certificate into the Java Truststore used by Pravega.

(No action is needed if the ECS's certificate is signed by public CA, as in the case of ECS test drive.)

1. Retrieve the TLS certificate of the ECS server, e.g. "ecs-certificate.pem".

2. Create a file with the secret references to the content and name of the ECS certificate.
    ```
    apiVersion: v1
    kind: Secret
    metadata:
      name: "segmentstore-tls"
    type: Opaque
    stringData:
      ECS_CERTIFICATE_NAME: "ecs-certificate.pem"
    data:
      ecs-certificate.pem: QmFnIEF0dH......JpYnV0ZLS0tLQo=
    ```

    Note: the secret name must be same as the secret name of other SegmentStore TLS materials, so all the materials can be mounted together.

    Assuming that the file is named `ecs-tls.yaml`, apply it so it appends, instead of replaces, the SegmentStore TLS secret.
    ```
    $ kubectl apply -f ecs-tls.yaml
    ```

3. In the Pravega manifest file, specify the name of the secret defined above to "tls/static/segmentStoreSecret". 
    ```
    ...
    kind: "PravegaCluster"
    metadata:
      name: "example"
    spec:
    tls:
      static:
        segmentStoreSecret: "segmentstore-tls"
    ...
    tier2:
        ecs:
          configUri: http://10.247.10.52:9020?namespace=pravega
          bucket: "shared"
          prefix: "example"
          credentials: ecs-credentials
    ```

### Use HDFS as Tier 2

Pravega can also use HDFS as the storage backend for Tier 2. The only requisite is that the HDFS backend must support Append operation.

Follow the [instructions to deploy Pravega manually](manual-installation.md#install-the-pravega-cluster-manually) and configure the Tier 2 block in your `PravegaCluster` manifest with your HDFS connection details.

```
spec:
  tier2:
    hdfs:
      uri: hdfs://10.28.2.14:8020/
      root: /example
      replicationFactor: 3
```
