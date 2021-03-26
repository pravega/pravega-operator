# Script for complete Pravega Installation

The purpose of this script is to sequentially deploy all the dependencies (i.e. Operators, Zookeeper and Bookkeeper) and Pravega Components in the right order.

## Prerequisites

  - Kubernetes 1.15+ with Beta APIs
  - Helm 3.2.1+
  - LongTerm Storage ([options for long term storage](https://github.com/pravega/pravega-operator/blob/master/doc/longtermstorage.md))
  - Cert-Manager v0.15.0+
  - An Issuer and a Certificate (either self-signed or CA signed) for the Bookkeeper Operator (refer to [this](https://github.com/pravega/bookkeeper-operator/blob/master/deploy/certificate.yaml) manifest to create a self-signed certificate in the default namespace)

We use cert-manager for certificate management for webhook services in Kubernetes. In case you plan to use the same, you would need to [install cert-manager](https://cert-manager.io/docs/installation/kubernetes/)

## Deploying the Pravega Cluster

To deploy the Pravega Cluster along with all the required dependencies, run the following command:

```
$ ./pravegacluster.sh deploy [CERT_NAME] [SECRET_NAME]
```
where:
- **[CERT_NAME]** is the name of the bookkeeper operator certificate created as a prerequisite (this is an optional parameter and its default value is `selfsigned-cert-bk`)
- **[SECRET_NAME]** is the name of the secret created by the above certificate (this is an optional parameter and its default value is `selfsigned-cert-tls-bk`)

## Undeploying the Pravega Cluster

To remove the Pravega Cluster along with all its dependencies, run the following command:

```
$ ./pravegacluster.sh destroy
```
