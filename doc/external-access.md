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

# Using External DNS names with segmentstore

When external access is enabled, a domain suffix can be provided in the cluster spec.
Each segmentstore external service is assigned a DNS name which is created by prefixing the domain name with segmentstore pod-name.

An [external DNS service](https://github.com/kubernetes-incubator/external-dns) (like AWS Route53, Google CloudDNS) could resolve this DNS name to the pods' externalIP when external clients try to access the segment store service.

External access can be enabled with/without DNS names support.

1. External Access=enabled with DNS names.

Configuration:
```
externalAccess:
    enabled: true
    type: LoadBalancer
    domainName: example.com
```

Post deployment, you should see:

```$> $kubectl describe svc pravega-pravega-segmentstore-0
. . .
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: pravega-pravega-segmentstore-0.example.com.
    ncp/internal_ip_for_policy: 100.64.65.241
. . .
```

In this case, external DNS name `pravega-pravega-segmentstore-0.example.com` is advertised by segment store for accepting connections from controller and clients.
SegmentStore logs show `publishedIPAddress: pravega-pravega-segmentstore-0.example.com`

2. External Access=enabled without DNS names.
When external access is enabled but 'domainName' is not provided in the manifest, external-dns annotation is not added to the segmentstore services and externalIP is advertised by segment store instead of DNS names.
SegmentStore logs show `publishedIPAddress: <SegmentStore Service External IP>`

Configuration:
```
externalAccess:
    enabled: true
    type: LoadBalancer

```

Post deployment, no annotation is seen on SegmentStore external service:
```$> $kubectl describe svc pravega-pravega-segmentstore-0
. . .
metadata:
  annotations:
    ncp/internal_ip_for_policy: 100.64.65.241
. . .

```

3. Overriding Service Type for Controller and SegmentStore external services.
When External Access is enabled, if ServiceType for Controller and SegmentStore is same, it can be specified under `externalAccess` > `type` as shown earlier.
However, if we need to specify different service types for Controller and SegmentStore services, this can be done by adding `controllerExtServiceType` or `segmentStoreExtServiceType` field for Controller and SegmentStore respectively, under the PravegaSpec.
When specified, these values will override the value against `externalAccess`, `type` field, if present.

For example when:
```
externalAccess:
    enabled: true
    type: LoadBalancer
     . . .

pravega:
    controllerExtServiceType: ClusterIP
    segmentStoreExtServiceType: NodePort

```
The ServiceType for Controller would be `ClusterIP` and that for SegmentStore would be `NodePort`.

4. Adding annotations to Controller and SegmentStore services

To add annotations to Controller and SegmentStore Services the `controllerSvcAnnotations` and `segmentStoreSvcAnnotations` feilds can be specified under PravegaSpec.

Example:
```
pravega:
    . . .
    controllerSvcAnnotations:
      service.beta.kubernetes.io/aws-load-balancer-access-log-enabled: "true"

    segmentStoreSvcAnnotations:
      service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-prefix: "abc"
```

Post deployment, these annotations would get added on Controller and SegmentStore services:
```
$> kubectl describe svc pravega-pravega-controller
Name:                     pravega-pravega-controller
Namespace:                default
Labels:                   app=pravega-cluster
                          component=pravega-controller
                          pravega_cluster=pravega
Annotations:              ncp/internal_ip_for_policy: 100.64.192.5
                          service.beta.kubernetes.io/aws-load-balancer-access-log-enabled: true
Selector:                 app=pravega-cluster,component=pravega-controller,pravega_cluster=pravega
Type:                     LoadBalancer
IP:                       10.100.200.76
LoadBalancer Ingress:     10.247.108.102
. . .
```

```
$> k describe svc pravega-pravega-segmentstore-1
Name:                     pravega-pravega-segmentstore-1
Namespace:                default
Labels:                   app=pravega-cluster
                          component=pravega-segmentstore
                          pravega_cluster=pravega
Annotations:              external-dns.alpha.kubernetes.io/hostname: pravega-pravega-segmentstore-1.example.com.
                          ncp/internal_ip_for_policy: 100.64.192.5
                          service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-prefix: abc
Selector:                 statefulset.kubernetes.io/pod-name=pravega-pravega-segmentstore-1
Type:                     LoadBalancer
IP:                       10.100.200.183
LoadBalancer Ingress:     10.247.108.104
. . .
```
# Exposing Segmentstore Service on single IP address and Different ports

For Exposing SegmentStoreservices on the same I/P address we will use MetalLB.
MetalLB hooks into Kubernetes cluster, and provides a network load-balancer implementation. In short, it allows to create Kubernetes services of type “LoadBalancer” in clusters that don’t run on a cloud provider and thus cannot simply hook into paid products to provide load-balancers.

By default, Services do not share an IP address, for providing same IP address to all the services we need to set the following configurations while creating the External Service:

1) Provide annotation key as **metallb.universe.tf/allow-shared-ip** for all the services.

2) All the services which want to share the IP address need to have the same value for the above annotation, for example "shared-ss-ip".

3) The port for all the services should be different

4) All the services should use External Traffic Policy as Cluster

5) Finally, we need to provide the I/P address that we want our service to provide to the segment store pod as spec.loadBalancerIP while creating the service

To enable this we need to provide segmentStoreSvcAnnotations, segmentStoreLoadBalancerIP, segmentStoreExternalTrafficPolicy in the manifest as shown below

Example:
```
pravega:
    . . .
    segmentStoreLoadBalancerIP: "10.243.39.103"
    segmentStoreExternalTrafficPolicy: "cluster"
    segmentStoreSvcAnnotations:
      metallb.universe.tf/allow-shared-ip: "shared-ss-ip"
```
