# Configuring SegmentStore Headless Service Name

By default segmentstore headless service name is configured as  `[PRAVEGA_CLUSTER_NAME]-pravega-segmentstore-headless`.

```
pravega-pravega-segmentstore-headless    ClusterIP    None    <none>    12345/TCP    2d16h

```
But we can configure the headless service name as follows:

```
helm install pravega pravega/pravega --set segmentStore.headlessSvcNameSuffix="segmentstore-svc"
```

After installation services can be listed using `kubectl get svc` command.

```
pravega-segmentstore-svc    ClusterIP    None    <none>    12345/TCP    2d16h

```

# Configuring Segmentsore Statefulset Name

By default segmentstore statefulset name  is configured as  `[PRAVEGA_CLUSTER_NAME]-pravega-segment-store`.

```
pravega-pravega-segment-store    1/1     2d17h

```
But we can configure the segmentstore statefulset name  as follows:

```
helm install pravega pravega/pravega --set segmentStore.stsNameSuffix="segmentstore-sts"
```

After installation sts can be listed using `kubectl get sts` command.

```
pravega-segmentstore-sts        1/1     2d17h

```

# Configuring Controller Service Name

By default controller service name is configured as  `[PRAVEGA_CLUSTER_NAME]-pravega-controller`.

```
pravega-pravega-controller    ClusterIP   10.100.200.173   <none>        10080/TCP,9090/TCP        2d16h

```

But we can configure the controller service name as follows:

```
helm install pravega pravega/pravega --set controller.svcNameSuffix="controllersvc"
```

After installation, services can be listed using `kubectl get svc` command.

```
pravega-controllersvc      ClusterIP   10.100.200.173   <none>        10080/TCP,9090/TCP     2d16h

```
