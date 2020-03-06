# Pravega Operator Helm Chart

Installs [pravega-operator](https://github.com/pravega/pravega-operator) to create/configure/manage Pravega clusters atop Kubernetes.

## Introduction

This chart bootstraps a [pravega-operator](https://github.com/pravega/pravega-operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Pravega Operator on multiple namespaces.

## Prerequisites
  - Kubernetes 1.10+ with Beta APIs
  - Helm 2.10+

## Installing the Chart

To install the chart with the release name `my-release`:

```
$ helm install --name my-release pravega-operator
```

The command deploys pravega-operator on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Pravega operator chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `image.repository` | Repository for pravega operator image | `pravega/pravega-operator` |
| `image.tag` | Tag for pravega operator image | `0.3.2` |
| `image.pullPolicy` | Pull policy for pravega operator image | `IfNotPresent` |
| `crd.create` | Create pravega CRD | `true` |
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create service account resources | `true` |
| `serviceAccount.name` | Name for the service account | `pravega-operator` |
