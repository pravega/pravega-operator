# Pravega Operator Helm Chart

Installs [pravega-operator](https://github.com/pravega/pravega-operator) to create/configure/manage Pravega clusters atop Kubernetes.

## Introduction

This chart bootstraps a [pravega-operator](https://github.com/pravega/pravega-operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Pravega Operator on multiple namespaces.

## Prerequisites
  - Kubernetes 1.15+ with Beta APIs
  - Helm 3.2.1+
  - An existing Apache Zookeeper 3.6.1 cluster. This can be easily deployed using our [Zookeeper Operator](https://github.com/pravega/zookeeper-operator)
  - An existing Apache Bookkeeper 4.9.2 cluster. This can be easily deployed using our [BookKeeper Operator](https://github.com/pravega/bookkeeper-operator)
  - Cert-Manager v0.15.0+ or some other certificate management solution in order to manage the webhook service certificates. This can be easily deployed by referring to [this](https://cert-manager.io/docs/installation/kubernetes/)
  - An Issuer and a Certificate (either self-signed or CA signed) in the same namespace that the Pravega Operator will be installed (refer to [this](https://github.com/pravega/pravega-operator/blob/master/deploy/certificate.yaml) manifest to create a self-signed certificate in the default namespace)
  > The name of the certificate (*webhookCert.certName*), the name of the secret created by this certificate (*webhookCert.secretName*), the tls.crt (*webhookCert.crt*) and tls.key (*webhookCert.key*) need to be specified against the corresponding fields in the values.yaml file, or can be provided with the install command as shown [here](#installing-the-chart).
  The values *tls.crt* and *tls.key* are contained in the secret which is created by the certificate and can be obtained using the following command
  ```
  kubectl get secret <secret-name> -o yaml | grep tls.
  ```

## Installing the Chart

To install the chart with the release name `my-release`:

```
$ helm install my-release pravega-operator --set webhookCert.crt=<tls.crt> --set webhookCert.generate=false --set webhookCert.certName=<cert-name> --set webhookCert.secretName=<secret-name>
```

The command deploys pravega-operator on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```
$ helm uninstall my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Pravega operator chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `image.repository` | Image repository | `pravega/pravega-operator` |
| `image.tag` | Image tag | `0.5.1` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `crd.create` | Create pravega CRD | `true` |
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Name for the service account | `pravega-operator` |
| `testmode.enabled` | Enable test mode | `false` |
| `testmode.version` | Major version number of the alternate pravega image we want the operator to deploy, if test mode is enabled | `""` |
| `webhookCert.crt` | tls.crt value corresponding to the certificate | |
| `webhookCert.key` | tls.key value corresponding to the certificate | |
| `webhookCert.generate` | Whether to generate the certificate and the issuer (set to false while using self-signed certificates) | `false` |
| `webhookCert.certName` | Name of the certificate, if generate is set to false | `selfsigned-cert` |
| `webhookCert.secretName` | Name of the secret created by the certificate, if generate is set to false | `selfsigned-cert-tls` |
| `watchNamespace` | Namespaces to be watched  | `""` |
