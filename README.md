# Pravega Operator

 [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/pravega/pravega-operator?status.svg)](https://godoc.org/github.com/pravega/pravega-operator) [![Build Status](https://travis-ci.org/pravega/pravega-operator.svg?branch=master)](https://travis-ci.org/pravega/pravega-operator) [![Go Report](https://goreportcard.com/badge/github.com/pravega/pravega-operator)](https://goreportcard.com/report/github.com/pravega/pravega-operator) [![Version](https://img.shields.io/github/release/pravega/pravega-operator.svg)](https://github.com/pravega/pravega-operator/releases)

## Overview

[Pravega](http://pravega.io) is an open source distributed storage service implementing Streams. It offers Stream as the main primitive for the foundation of reliable storage systems: *a high-performance, durable, elastic, and unlimited append-only byte stream with strict ordering and consistency*.

The Pravega Operator manages Pravega clusters deployed to Kubernetes and automates tasks related to operating a Pravega cluster.The operator itself is built with the [Operator framework](https://github.com/operator-framework/operator-sdk).

## Project status

The project is currently beta. While no breaking API changes are currently planned, we reserve the right to address bugs and change the API before the project is declared stable.

## Quickstart

For setting up pravega quickly, check our [complete Pravega Installation script](setup/README.md). For installing pravega on minikube,refer to [minikube](doc/minikube_setup.md).

## Install the Operator

To understand how to deploy a Pravega Operator refer to [Operator Deployment](charts/pravega-operator#deploying-pravega-operator).

## Upgrade the Operator

For upgrading the pravega operator check the document [Operator Upgrade](doc/operator-upgrade.md).

## Features

- [x] [Create and destroy a Pravega cluster](charts/pravega/README.md)
- [x] [Resize cluster](charts/pravega/README.md#updating-pravega-cluster)
- [x] [Rolling upgrades/Rollback](doc/upgrade-cluster.md)
- [x] [Pravega Configuration tuning](doc/configuration.md)
- [x] Input validation

## Development

Check out the [development guide](doc/development.md).

## Releases  

The latest Pravega operator releases can be found on the [Github Release](https://github.com/pravega/pravega-operator/releases) project page.

## Contributing and Community

We thrive to build a welcoming and open community for anyone who wants to use the operator or contribute to it. [Here](CONTRIBUTING.md) we describe how to contribute to pravega operator.Contact the developers and community on [slack](https://pravega-io.slack.com/) ([signup](https://pravega-slack-invite.herokuapp.com/)) if you need any help.

## Troubleshooting

Check out the [pravega troubleshooting](doc/troubleshooting.md#pravega-cluster-issues) for pravega issues and for operator issues [operator troubleshooting](doc/troubleshooting.md#pravega-operator-issues).

## License

Pravega Operator is under Apache 2.0 license. See the [LICENSE](https://github.com/pravega/pravega-operator/blob/master/LICENSE) for details.
