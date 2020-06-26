# Supported Upgrade Paths
Till Operator version **0.4.x** only minor version upgrades were supported e.g. _0.4.0 -> 0.4.1_.

Starting Operator version `0.4.3` we also support major version upgrades for Pravega Operator.

  `0.4.3 --> 0.5.0`

  `0.4.4 -->0.5.0`

# Upgrade Guide

## Upgrading till 0.4.4
Till version 0.4.4, Pravega operator can be upgraded by providing the location to the new charts

`$ helm upgrade <operator release name> <location of new charts>`

The upgrade is handled as a [rolling update](https://kubernetes.io/docs/tutorials/kubernetes-basics/update/update-intro/) by Kubernetes and results in a new operator pod being created and the old one being terminated.

## Upgrading from 0.4.x to 0.5.0

### What changed between 0.4.x and 0.5.0 ?
A lot has changed between operator versions _0.4.x_ and _0.5.0_.

Here is a list of changes and their impact:

1. The Pravega Cluster CRD

Till Operator `0.4.4`, the Pravega CR (version `v1alpha1`) includes Bookkeeper.
Starting version `0.5.0`, Pravega CR does **not** include Bookkeeper.
Bookkeeper is moved out and is now a prerequisite for Pravega deployment.
It can be installed separately either using [Bookkeeper-Operator](https://github.com/pravega/bookkeeper-operator) or some other means.
The `Bookkeeper` field in PravegaCluster Spec in version `v1alpha1` is replaced with field `BookkeeperUri` in version `v1beta1`.

2. CR conversion from `v1alpha1` to `v1beta1`

When upgrading to Operator version `0.5.x`, it is necessary to migrate the Pravega CR Object from the old version `v1alpha1`[with Bookkeeper] to the new version `v1beta1`[without Bookkeeper].
This is done by a _Conversion Webhook_ that gets triggered automatically when Operator is upgraded to 0.5.0.
See:
https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definition-versioning/

The Conversion Webhook is part of Pravega Operator code (_starting 0.5.0-rc1_) and it does the following:

  a. Converts a Pravega CR `v1alpha1 `Object to `v1beta1 `Object by copying all values in the Pravega Spec for Controller, SegmentStore, Tier2 (LongTermStorage) etc...

  b. Creates a new Bookkeeper CR object based on the Bookkeeper [CRD](https://github.com/pravega/bookkeeper-operator/blob/master/deploy/crds/crd.yaml) and populates it with values from the Bookkeeper Spec in Pravega CR Object `v1alpha1`. The name & namespace of this object is as that of Pravega CR object.

  c. Sets owner/controller reference of all Bookkeeper objects _[STS, ConfigMap, PVCs, PDB & Headless Service]_ to the new _Bookkeeper CR Object_ and removes references to the _Pravega CR Object_.
With this, all existing Bookkeeper artifacts are migrated to Bookkeeper CR from Pravega CR. The Bookkeeper CR object so created, should be managed by a Bookkeeper Operator.

  d. Sets owner references of all Pravega Cluster objects [_Controller Deployment, Services, Config Maps, SegmentStore STS, Services, Config Maps etc..._] to APIVersion "pravega.pravega.io/v1beta1" from earlier "pravega.pravega.io/v1alpha1". This makes sure on deletion of Pravega CR Object, all these artifacts are also deleted

  e. Deletes and recreates Bookkeeper and Segment Store STS, since owner references on existing STS cannot be updated.

3. OpenAPIV3Schema Validation

Structural schemas are a requirement for Custom Resources [_starting `apiextensions.k8s.io/v1`_], and not specifying one disables _ConversionWebhook_ feature in version `_apiextensions.k8s.io/v1beta1_`.
As such, the new Pravega Cluster CRD includes OpenAPIV3Schema Validation for Pravega CRs for both versions `v1alpha1` and `v1beta1`.
See: https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#specifying-a-structural-schema

4. Controller Runtime upgrade

Controller Runtime library has been upgraded from v0.1.8 to v0.5.1 as the older one does not support conversion webhooks.
The operator uses hub-spokes model in controller-runtime to achieve version conversion.

See: https://book.kubebuilder.io/multiversion-tutorial/tutorial.html

5. Mutating Webhook replaced with Validating Webhook

The mutating webhook in operator(0.4.x) has been replaced with a validating webhook that validates pravega version information entered by the user.
This can be extended later to add more validations in the near future.
A mutating webhook was no longer needed as _tag_ field in PravegaImageSpec has been removed in `v1beta1` and only "version" field in the Pravega Spec is now supported.

6. `Tier2` renamed to `LongTermStorage` in Pravega Spec

The field "Tier2" in Pravega Spec has been renamed to LongTermStorage in `v1beta1`.
During upgrade the version conversion code takes care of migrating values mentioned inside `Tier2` in `v1alpha1` to field `LongTermStorage` in `v1beta1` as there is no `Tier2` feild in v1beta1.

### Pre-requisites

For upgrading Operator to version 0.5.0, the following must be true:
1. The Kubernetes Server version must be at least 1.15, (WebhookConversion is a beta feature in Kubernetes 1.15)
See: https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definition-versioning/#webhook-conversion

2. Cert-Manager or some other certificate management solution must be deployed for managing webhook service certificates. The upgrade trigger script assume user has [cert-manager](https://cert-manager.io/docs/installation/kubernetes/) installed but any other cert management solution can also be used and script would need to be modified accordingly.

  Installing cert-manager: https://cert-manager.io/docs/installation/kubernetes/

3. [Bookkeeper Operator](https://github.com/pravega/bookkeeper-operator/tree/master/charts/bookkeeper-operator) must be deployed in the same namespace as Pravega Operator, prior to triggering the upgrade.Also, Bookkeeper operator config map should contain the bookkeeper versions of the installed bookkeeper.

4. Install an Issuer and a Certificate (either self-signed or CA signed) in the same namespace as the Pravega Operator (you can refer to the manifest [here](https://github.com/pravega/pravega-operator/blob/master/deploy/certificate.yaml) to create a self-signed certificate in the default namespace). You would need to provide the name of the certificate (webhookCert.certName), the name of the secret created by this certificate (webhookCert.secretName), the tls.crt (webhookCert.crt) and tls.key (webhookCert.key)  inside the corresponding fields in the values.yaml file of the charts for the 0.5.0 operator. The values tls.crt and tls.key are contained in the secret which is created by the certificate (say `selfsigned-cert-tls`) and can be obtained using the following command
```
kubectl get secret selfsigned-cert-tls -o yaml | grep tls.
```

5. Execute the script `pre-upgrade.sh` inside the `scripts` folder. This script patches the `pravega-webhook-svc` with the required annotations and labels.


### Triggering the upgrade

The upgrade to Operator 0.5.0 can be triggered using the following command
```
helm upgrade <operator release name> <location of 0.5.0 charts> --set webhookCert.crt=<tls.crt> --set webhookCert.generate=false --set webhookCert.certName=<cert-name> --set webhookCert.secretName=<secret-name>
```

### How to check for successful upgrade completion

As in earlier cases, upgrading to operator version `0.5.0`, causes the old operator pod to be terminated and a new operator pod (_with image 0.5.0_) to be created.

The new pravega operator immediately detects that the stored version of Pravega CR(`v1alpha1`) is different from the watched version(`v1beta1`) and triggers the conversion webhook for converting the CR from `v1alpha1` to `v1beta1`.
These logs in the new p-operator pod indicate that a conversion was attempted and completed successfully:

"_Converting Pravega CR version from v1alpha1 to v1beta1_"

eventually followed by message:

"_Version migration completed successfully._"

The execution of version conversion code in operator takes from a few secs upto a minute.
But for these changes to take effect on Kubernetes server, it may take upto 8-10 minutes.
During this period, resource requests on the pravegacluster object (`kubectl get` and `kubectl describe` calls) will fail and this is expected.

Once conversion completes on server, these requests should reflect appropriate values and status for Pravega and Bookkeeper CR Objects.

### Impact on Pravega

1. Till Pravega CR conversion has completed on the Kubernetes server (_detected by a successful_ `kubectl get/describe pravegacluster <clustername>`) no **scale/update/delete** operations should be attempted on Pravega/Bookkeeper clusters as this can mess up with the conversion process and leave the cluster in an inconsistent state.

2. If I/O is running on Pravega during this upgrade, for a few minutes (_till Bookkeeper and SegmentStore STS/Services are recreated_) Readers/Writers may be stalled. Nevertheless they will be able to resume again, provided retries are appropriately configured and/or readers that have bailed out due to "RetriesExhaustedException" are restarted again by the client application.

### Migrating out of Pravega CR `v1alpha1`
Operator 0.5.0 watches and reconciles only `v1beta1` Pravega CR and not `v1alpha1`. After all Pravega Clusters in a given K8s Cluster have migrated to version `v1beta1` the `served:` flag for version `v1alpha1` can be set to `false` in the [PravegaCRD](https://github.com/pravega/pravega-operator/blob/v0.5.0-rc1/tools/manifest_files/crd.yaml#L308) to indicate this is no longer a supported Pravega CR version.
