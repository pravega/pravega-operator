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

When the Pravega cluster is deployed and ready, clients will need to connect to the external Controller address and will automatically discover the external address of all Segment Store pods.


```
$ kubectl get svc -lapp=pravega-cluster
NAME                                    TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)                          AGE
example-bookie-headless                 ClusterIP      None            <none>          3181/TCP                         4m
example-pravega-controller              LoadBalancer   10.31.243.62    35.239.48.145   10080:30977/TCP,9090:30532/TCP   4m
example-pravega-segmentstore-0          LoadBalancer   10.31.252.166   34.66.68.236    12345:32614/TCP                  4m
example-pravega-segmentstore-1          LoadBalancer   10.31.250.183   34.66.58.131    12345:31966/TCP                  4m
example-pravega-segmentstore-2          LoadBalancer   10.31.250.233   34.66.231.244   12345:31748/TCP                  4m
example-pravega-segmentstore-headless   ClusterIP      None            <none>          12345/TCP                        4m
```

In the example above, clients will connect to the Pravega Controller at `tcp://35.239.48.145:9090`.

# Using External DNS names for segment store pods

Given a domain suffix for the cluster, each segment store pod can be assigned a DNS name. An external DNS service can resolve this DNS name to the pod externalIP.
Domain name can be provided in the manifest using `domainname` key under `externalAccess`.
Segment store dns name is created by prefixing the domain name with segmentstore pod-name.

1. External Access with External DNS
To enable external access with dns set the folloing in the manifest file:
```
externalAccess:
    enabled: true
    type: LoadBalancer
    domainName: example.com
```
After deployment, you should see:
```$> $kubectl describe svc pravega-pravega-segmentstore-0
. . .
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: pravega-pravega-segmentstore-0.example.com.
    ncp/internal_ip_for_policy: 100.64.65.241
. . .
```
In this case, external dnsname `pravega-pravega-segmentstore-0.example.com` is advertised by segment store for accepting connections from controller and clients. SegmentStore logs show `publishedIPAddress: pravega-pravega-segmentstore-0.example.com`

2. When external_access = true, but 'domainName' is not added in manifest, external-dns annotation is not added to the service and externalIP is advertised by segment store. SegmentStore logs show `publishedIPAddress: 10.240.119.222`

```$> $kubectl describe svc pravega-pravega-segmentstore-0
. . .
metadata:
  annotations:
    ncp/internal_ip_for_policy: 100.64.65.241
. . .

```

3. When external access is disabled, ClusterIP is advertised by Segment Store. 
