# Deploying Pravega-Operator

Here, we briefly describe how to install [pravega-operator](https://github.com/pravega/pravega-operator) that is used to create/configure/manage Pravega clusters atop Kubernetes.

## Prerequisites
  - Kubernetes 1.15+ with Beta APIs
  - Helm 3.2.1+
  - Cert-Manager v0.15.0+ or some other certificate management solution in order to manage the webhook service certificates. This can be easily deployed by referring to [this](https://cert-manager.io/docs/installation/kubernetes/)
  - An Issuer and a Certificate (either self-signed or CA signed) in the same namespace that the Pravega Operator will be installed (refer to [this certificate.yaml file](../../deploy/certificate.yaml) manifest to create a self-signed certificate in the default namespace)

## Installing Pravega-Operator

> Note: If you are running on Google Kubernetes Engine (GKE), please [check this first](../../doc/development.md#installation-on-google-kubernetes-engine).

To install the pravega-operator chart, use the following commands:

```
$ helm repo add pravega https://charts.pravega.io
$ helm repo update
$ helm install [RELEASE_NAME] pravega/pravega-operator --version=[VERSION] --set webhookCert.certName=[CERT_NAME] --set webhookCert.secretName=[SECRET_NAME]
```
where:

- **[RELEASE_NAME]** is the release name for the pravega-operator chart.
- **[VERSION]** can be any stable release version for pravega-operator from 0.5.0 onwards.
- **[CERT_NAME]** is the name of the certificate created as a prerequisite
- **[SECRET_NAME]** is the name of the secret created by the above certificate

This command deploys a pravega-operator on the Kubernetes cluster in its default configuration. The [configuration](#operator-configuration) section lists the parameters that can be configured during installation.

>Note: If we provide [RELEASE_NAME] same as chart name, deployment name will be same as release-name. But if we are providing a different name for release(other than pravega-operator in this case), deployment name will be [RELEASE_NAME]-[chart-name]. However, deployment name can be overridden by providing --set  fullnameOverride=[DEPLOYMENT_NAME]` along with helm install command

>Note: If the pravega-operator version is 0.4.5, webhookCert.certName and webhookCert.secretName should not be set. Also in this case, bookkeeper operator, cert-manager and the certificate/issuer do not need to be deployed as prerequisites.

## Upgrading Pravega-Operator

For upgrading pravega-operator, please refer [upgrade guide](../../doc/operator-upgrade.md)

## Uninstalling Pravega-Operator

To uninstall/delete the pravega-operator chart, use the following command:

```
$ helm uninstall [RELEASE_NAME]
```

This command removes all the Kubernetes components associated with the chart and deletes the release.

## Operator Configuration

The following table lists the configurable parameters of the pravega-operator chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `image.repository` | Image repository | `pravega/pravega-operator` |
| `image.tag` | Image tag | `0.5.3` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `crd.create` | Create pravega CRD | `true` |
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Name for the service account | `pravega-operator` |
| `testmode.enabled` | Enable test mode | `false` |
| `testmode.version` | Major version number of the alternate pravega image we want the operator to deploy or provide an upgrade path to, if test mode is enabled | `""` |
| `testmode.fromVersion` | Major version number of the alternate pravega image, if we wish to provide an upgrade path from this version to the version mentioned above, if test mode is enabled | `""` |
| `webhookCert.crt` | tls.crt value corresponding to the certificate | |
| `webhookCert.key` | tls.key value corresponding to the certificate | |
| `webhookCert.generate` | Whether to generate the certificate and the issuer (set to false while using self-signed certificates) | `false` |
| `webhookCert.certName` | Name of the certificate, if generate is set to false | `selfsigned-cert` |
| `webhookCert.secretName` | Name of the secret created by the certificate, if generate is set to false | `selfsigned-cert-tls` |
| `watchNamespace` | Namespaces to be watched  | `""` |
