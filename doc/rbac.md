## Setting up RBAC for Pravega operator

### Use non-default service accounts

You can optionally configure non-default service accounts for Pravega Controller and Pravega Segment Store pods.

Set the `controllerServiceAccountName` and `segmentStoreServiceAccountName` fields under the `pravega` block.

```
...
spec:
  pravega:
    controllerServiceAccountName: ctrl-service-account
    segmentStoreServiceAccountName: ss-service-account
...
```

If external access is enabled in your Pravega cluster, Segment Store pods will require access to some Kubernetes API endpoints to obtain the external IP and port. Make sure that the service account you are using for the Segment Store has, at least, the following permissions.

```
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: pravega-components
  namespace: "pravega-namespace"
rules:
- apiGroups: ["pravega.pravega.io"]
  resources: ["*"]
  verbs: ["get"]
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get"]
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: pravega-components
rules:
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get"]
```

Replace the `namespace` with your own namespace.

### Installing on a Custom Namespace with RBAC enabled

Create the namespace.

```
$ kubectl create namespace pravega-io
```

Update the namespace configured in the `deploy/role_binding.yaml` file.

```
$ sed -i -e 's/namespace: default/namespace: pravega-io/g' deploy/role_binding.yaml
```

Apply the changes.

```
$ kubectl -n pravega-io apply -f deploy
```

Note that the Pravega operator only monitors the `PravegaCluster` resources which are created in the same namespace, `pravega-io` in this example. Therefore, before creating a `PravegaCluster` resource, make sure an operator exists in that namespace.

```
$ kubectl -n pravega-io create -f example/cr.yaml
```

```
$ kubectl -n pravega-io get pravegaclusters
NAME      AGE
pravega   28m
```

```
$ kubectl -n pravega-io get pods -l pravega_cluster=pravega
NAME                                          READY     STATUS    RESTARTS   AGE
pravega-pravega-controller-6c54fdcdf5-947nw   1/1       Running   0          29m
pravega-pravega-segmentstore-0                1/1       Running   0          29m
pravega-pravega-segmentstore-1                1/1       Running   0          29m
pravega-pravega-segmentstore-2                1/1       Running   0          29m
```
