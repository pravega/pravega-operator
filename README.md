# Pravega Operator

 [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/pravega/pravega-operator?status.svg)](https://godoc.org/github.com/pravega/pravega-operator) [![Build Status](https://travis-ci.org/pravega/pravega-operator.svg?branch=master)](https://travis-ci.org/pravega/pravega-operator) [![Go Report](https://goreportcard.com/badge/github.com/pravega/pravega-operator)](https://goreportcard.com/report/github.com/pravega/pravega-operator) [![Version](https://img.shields.io/github/release/pravega/pravega-operator.svg)](https://github.com/pravega/pravega-operator/releases)

## Table of Contents

 * [Overview](#overview)
    * [Features](#features)
    * [Project status](#project-status)
    * [Requirements](#requirements)
 * [Quickstart](#quickstart)
    * [Install the Operator](#install-the-operator)
    * [Install the Cluster](#install-the-cluster)
    * [Installing on minikube](#installing-on-minikube)
    * [Script for complete Pravega Installation](#script-for-complete-pravega-installation)
    * [Scale the Cluster](#scale-the-cluster)
    * [Upgrade the Cluster](#upgrade-the-cluster)
    * [Upgrade the Operator](#upgrade-the-operator)
 * [Development](#development)
 * [Releases](#releases)
 * [Community](#community)
 * [Troubleshooting](#troubleshooting)

## Overview

[Pravega](http://pravega.io) is an open source distributed storage service implementing Streams. It offers Stream as the main primitive for the foundation of reliable storage systems: *a high-performance, durable, elastic, and unlimited append-only byte stream with strict ordering and consistency*.

The Pravega Operator manages Pravega clusters deployed to Kubernetes and automates tasks related to operating a Pravega cluster.The operator itself is built with the [Operator framework](https://github.com/operator-framework/operator-sdk).

### Features

- [x] Create and destroy a Pravega cluster
- [x] Resize cluster
- [x] Rolling upgrades
- [x] Rollback support
- [x] Support modification of configuration parameters at runtime
- [x] Open API Schema validation
- [x] External Access support

### Project status

The project is currently beta. While no breaking API changes are currently planned, we reserve the right to address bugs and change the API before the project is declared stable.

### Requirements

- Kubernetes 1.15+
- Helm 3.2.1+
- Cert-Manager v0.15.0+ or some other certificate management solution in order to manage the webhook service certificates. This can be easily deployed by referring to [this](https://cert-manager.io/docs/installation/kubernetes/)
- An Issuer and a Certificate (either self-signed or CA signed) in the same namespace that the Pravega Operator will be installed (refer to [this](https://github.com/pravega/pravega-operator/blob/master/deploy/certificate.yaml) manifest to create a self-signed certificate in the default namespace)

## Quickstart

We recommend using our [helm charts](charts) for all installation and upgrades (but not for rollbacks at the moment since helm rollbacks are still experimental). The helm charts for pravega operator (version 0.4.5 onwards) and pravega cluster (version 0.5.0 onwards) are published in [https://charts.pravega.io](https://charts.pravega.io/).

### Install the Operator

> Note: If you are running on Google Kubernetes Engine (GKE), please [check this first](doc/development.md#installation-on-google-kubernetes-engine).

To understand how to deploy a Pravega Operator using helm, refer to [this](charts/pravega-operator#deploying-pravega-operator).

### Install the cluster

To understand how to deploy a pravega cluster using helm, refer to [this](charts/pravega#deploying-a-pravega-cluster).

### Upgrade the Operator

For upgrading the pravega operator check the document [operator-upgrade](doc/operator-upgrade.md).

### Installing on minikube

For installing operator and cluster on minikbe, refer to [minikube](doc/minikube_setup.md).

### Script for complete Pravega Installation

For installing pravega and all the components together, please refer to [this](setup/README.md).

### Scale the Cluster

For scaling the parvega cluster using helm, refer to [this](charts/pravega#updating-the-chart).

### Upgrade the cluster

Check out the [upgrade guide](doc/upgrade-cluster.md).

## Configuration

Check out the [configuration document](doc/configuration.md).

## Development

Check out the [development guide](doc/development.md).

## Releases  

The latest Pravega operator releases can be found on the [Github Release](https://github.com/pravega/pravega-operator/releases) project page.

## Community

Contact the developers and community on [slack](https://pravega-io.slack.com/) ([signup](https://pravega-slack-invite.herokuapp.com/)) if you need any help.

## Troubleshooting

Check out the [troubleshooting document](doc/troubleshooting.md).

## License
Pravega Operator is under Apache 2.0 license. See the [LICENSE](https://github.com/pravega/pravega-operator/blob/master/LICENSE) for details.
