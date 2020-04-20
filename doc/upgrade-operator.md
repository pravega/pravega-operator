### Pravega Operator Upgrade
Pravega operator can be upgraded by modifying the image tag using
```
$ kubectl edit <operator deployment name>
```
This approach will work for upgrades upto 0.4.x

Starting from operator version 0.5 onwards, pravega operator is not handling bookies, and that will be handled by bookkeeper-operator.So while upgrading from 0.4.x to 0.5, we have to transfer the ownership of bookkeeper objects to bookkeeper operator. For this we are maintaining 2 versions namely, v1alpha1(Pravega Custom Resource with Bookkeeper) and v1beta1(Pravega Custom Resource without Bookkeeper) inside the crd.  And during upgrade we are triggering conversion webhook that will change the current version to v1beta1.`bookkeeperUri` field is added in v1beta1.Also, Tier2 storage name is changed to `longTermStorage`

More details on upgrade can be  found at  https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definition-versioning/

upgrading operator  from 0.4.x to 0.5.0 can be triggered using the operatorUpgrade.sh script under tools folder. This script does all necessary configuration changes to enable the upgrade and also triggers the upgrade.Before executing the script install cert-manager. This can be done using the steps mentioned here https://cert-manager.io/docs/installation/kubernetes/

#### Executing the script
git clone pravega-operator repo

Change the path to `tools` folder

Execute the script as follows
```
./operatorUpgrade.sh <pravega-operator deployment name> <pravega-operator deployment namespace> <pravega-operator new image-repo/image-tag>

```
Once the script is ran successfully, ensure that operator upgrade is completed and new operator pod is in running state.

Once upgrade is completed, pravega-operator pod will come up with new operator version. Also new existing stateful set for bookies and segment stores will get recreated.

#### Known Issues
 - Currently, more than 1 replica of pravega-operator is not supported.
