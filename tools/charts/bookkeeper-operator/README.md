# Bookkeeper Operator Helm Chart

Installs [bookkeeper-operator](https://github.com/pravega/bookkeeper-operator) to create/configure/manage Bookkeeper clusters atop Kubernetes.

## Introduction

This chart bootstraps a [bookkeeper-operator](https://github.com/pravega/bookkeeper-operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Bookkeeper Operator on multiple namespaces.

## Prerequisites
  - Kubernetes 1.10+ with Beta APIs
  - Helm 2.10+
  - An existing Apache Zookeeper 3.5 cluster. This can be easily deployed using our [Zookeeper operator](https://github.com/pravega/zookeeper-operator

## Installing the Chart

To install the chart with the release name `my-release`:

```
$ helm install --name my-release bookkeeper-operator
```

The command deploys bookkeeper-operator on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Bookkeeper operator chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `image.repository` | Repository for bookkeeper operator image | `pravega/bookkeeper-operator` |
| `image.tag` | Tag for bookkeeper operator image | `latest` |
| `image.pullPolicy` | Pull policy for bookkeeper operator image | `IfNotPresent` |
| `crd.create` | Create bookkeeper CRD | `true` |
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create service account resources | `true` |
| `serviceAccount.name` | Name for the service account | `bookkeeper-operator` |
