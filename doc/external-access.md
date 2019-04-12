# Enable external access

By default, a Pravega cluster uses `ClusterIP` services which are only accessible from within Kubernetes. However, when creating the Pravega cluster resource, you can opt to enable external access.

In Pravega, clients initiate the communication with the Pravega Controller, which is a stateless component frontended by a Kubernetes service that load-balances the requests to the backend pods. Then, clients discover the individual Segment Store instances to which they directly read and write data to. Clients need to be able to reach each and every Segment Store pod in the Pravega cluster.

If your Pravega cluster needs to be consumed by clients from outside Kubernetes (or from another Kubernetes deployment), you can enable external access in two ways, depending on your environment constraints and requirements. Both ways will create one service for all Controllers, and one service for each Segment Store pod.

1. Via [`LoadBalancer`](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer) service type.
2. Via [`NodePort`](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport) service type.

You can read more about service types in the [Kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types) to understand which one fits your use case.

When external access is enabled, Segment Store pods need to query the Kubernetes API to find out which is their external IP and port depending on the service type. Therefore, you need to make sure that the service accounts configured have the right permissions, otherwise Segment Store pods will be unable to bootstrap and will crash.

Below you can find example resources to create a service account, give it the minimum required permissions to obtain the external address, and configure and enable it on the `PravegaCluster` manifest.

1. Create a service account for Pravega components.

```
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pravega-components
```

2. Create `Role` and `ClusterRole` with the minimum required permissions. Make sure to update the `default` namespace if you are deploying Pravega to a custom namespace.

```
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: pravega-components
  namespace: "default"
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

3. Bind roles to the service account.

```
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: pravega-components
subjects:
- kind: ServiceAccount
  name: pravega-components
roleRef:
  kind: Role
  name: pravega-components
  apiGroup: rbac.authorization.k8s.io
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: pravega-components
subjects:
- kind: ServiceAccount
  name: pravega-components
  namespace: default
roleRef:
  kind: ClusterRole
  name: pravega-components
  apiGroup: rbac.authorization.k8s.io
```

4. Configure the service account and enable external on the `PravegaCluster` manifest.

```
apiVersion: "pravega.pravega.io/v1alpha1"
kind: "PravegaCluster"
metadata:
  name: "example"
spec:
  externalAccess:
    enabled: true
    type: LoadBalancer

  bookkeeper:
    serviceAccountName: pravega-components
...
  pravega:
    controllerServiceAccountName: pravega-components
    segmentStoreServiceAccountName: pravega-components
...
```
