# Enable TLS

Client can communicate with Pravega in a more secure way using TLS. To enable this feature, you will first need to
create secrets for Controller and Segment Store to make the relevant, sensible files available to the backend pods.

```
$ kubectl create secret generic controller-tls \
  --from-file=./controller01.pem \
  --from-file=./ca-cert \
  --from-file=./controller01.key.pem \
  --from-file=./controller01.jks \
  --from-file=./password
```

```
$ kubectl create secret generic segmentstore-tls \
  --from-file=./segmentstore01.pem \
  --from-file=./ca-cert \
  --from-file=./segmentstore01.key.pem \
  --from-file=./segmentstore01.jks \
  --from-file=./password
```

Then specify the secret names in the `tls` block and the TLS parameters in the `options` block.

```
apiVersion: "pravega.pravega.io/v1alpha1"
kind: "PravegaCluster"
metadata:
  name: "example"
spec:
  tls:
    static:
      controllerSecret: "controller-tls"
      segmentStoreSecret: "segmentstore-tls"
...
  pravega:
    options:
      controller.security.tls.enable: "true"
      controller.security.tls.server.certificate.location: "/etc/secret-volume/controller01.pem"
      controller.security.tls.server.privateKey.location: "/etc/secret-volume/controller01.key.pem"
      controller.security.tls.trustStore.location: "/etc/secret-volume/ca-cert"
      controller.security.tls.server.keyStore.location: "/etc/secret-volume/controller01.jks"
      controller.security.tls.server.keyStore.pwd.location: "/etc/secret-volume/password"
      pravegaservice.security.tls.enable: "true"
      pravegaservice.security.tls.server.certificate.location: "/etc/secret-volume/segmentStore01.pem"
      pravegaservice.security.tls.server.privateKey.location: "/etc/secret-volume/segmentStore01.key.pem"
      pravegaservice.security.tls.server.keyStore.location: "/etc/secret-volume/segmentStore01.jks"
      pravegaservice.security.tls.server.keyStore.pwd.location: "/etc/secret-volume/password"
...
```

Note that Pravega operator uses `/etc/secret-volume` as the mounting directory for secrets.

For more security configurations, check [here](https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md).
