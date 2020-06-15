# Pravega Operator Helm Chart

Installs [pravega-operator](https://github.com/pravega/pravega-operator) to create/configure/manage Pravega clusters atop Kubernetes.

## Introduction

This chart bootstraps a [pravega-operator](https://github.com/pravega/pravega-operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Pravega Operator on multiple namespaces.

## Prerequisites
  - Kubernetes 1.15+ with Beta APIs
  - Helm 3+
  - An existing Apache Zookeeper 3.5 cluster. This can be easily deployed using our [Zookeeper Operator](https://github.com/pravega/zookeeper-operator)
  - An existing Apache Bookkeeper 4.9.2 cluster. This can be easily deployed using our [BookKeeper Operator](https://github.com/pravega/bookkeeper-operator)

## Installing the Chart

To install the chart with the release name `my-release`:

```
$ helm install my-release pravega-operator
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
| `image.tag` | Image tag | `0.5.0` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `crd.create` | Create pravega CRD | `true` |
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Name for the service account | `pravega-operator` |
| `testmode` | Enable test mode | `false` |
| `webhookCert.selfsigned` | Whether to use self-signed certificates | true |
| `webhookCert.crt` | Certificate provided by the CA, if self-signed certificates are not to be used | |
| `webhookCert.key` | Private key provided by the CA, if self-signed certificates are not to be used | |
| `watchNamespace` | Namespaces to be watched  | `""` |
