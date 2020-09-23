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

## Installing the Chart

To install the pravega-operator chart, use the following commands:

```
$ helm repo add pravega https://charts.pravega.io
$ helm repo update
$ helm install [RELEASE_NAME] pravega/pravega-operator --version=[VERSION] --set webhookCert.crt=[TLS_CRT] --set webhookCert.generate=false --set webhookCert.certName=[CERT_NAME] --set webhookCert.secretName=[SECRET_NAME]
```
where:
- **[RELEASE_NAME]** is the release name for the pravega-operator chart.
  **[RESOURCE_NAME]** is the name of the pravega-operator resource so created. (If [RELEASE_NAME] contains the string `pravega-operator`, `[RESOURCE_NAME] = [RELEASE_NAME]`, else `[RESOURCE_NAME] = [RELEASE_NAME]-pravega-operator`. The [RESOURCE_NAME] can however be overridden by providing `--set fullnameOverride=[RESOURCE_NAME]` along with the helm install command)
- **[VERSION]** can be any stable release version for pravega-operator from 0.5.0 onwards.
- **[CERT_NAME]** is the name of the certificate created in the previous step
- **[SECRET_NAME]** is the name of the secret created by the above certificate
- **[TLS_CRT]** is contained in the above secret and can be obtained using the command `kubectl get secret [SECRET_NAME] -o yaml | grep tls.crt`

This command deploys a pravega-operator on the Kubernetes cluster in its default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

>Note: If the pravega-operator version is 0.4.5, webhookCert.generate, webhookCert.crt, webhookCert.certName and webhookCert.secretName should not be set. Also in this case, bookkeeper operator, cert-manager and the certificate/issuer do not need to be deployed as prerequisites.

## Uninstalling the Chart

To uninstall/delete the pravega-operator chart, use the following command:

```
$ helm uninstall [RELEASE_NAME]
```

This command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the pravega-operator chart and their default values.

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
