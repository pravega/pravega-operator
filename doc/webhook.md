## Admission Webhook

[Admission webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) are HTTP callbacks that receive admission requests and do something with them.
There are  two webhooks [ValidatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#validatingadmissionwebhook) and 
[MutatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) which are basically 
doing the same thing except MutatingAdmissionWebhook can modify the requests. In our case, we use MutatingAdmissionWebhook because it can validate requests as well as mutating them. E.g. clear the image tag 
if version is specified.

In the Pravega operator repo, we are leveraging the webhook implementation from controller-runtime package, here is the [GoDoc](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/webhook). 
In detail, there are two steps that developers need to do 1) create webhook server and 2) implement the handler.
The webhook server registers webhook configuration with the apiserver and creates an HTTP server to route requests to the handlers.
The server is behind a Kubernetes Service and provides a certificate to the apiserver when serving requests. The kubebuilder has a detailed instruction of 
building a webhook, see [here](https://github.com/kubernetes-sigs/kubebuilder/blob/86026527c754a144defa6474af6fb352143b9270/docs/book/beyond_basics/sample_webhook.md).

The webhook feature itself is enabled by default but it can be disabled if `webhook=false` is specified when installing the 
operator locally using `operator-sdk up local`. E.g. ` operator-sdk up local --operator-flags -webhook=false`. The use case of this is that webhook needs to be
disabled when developing the operator locally since webhook can only be deployed in Kubernetes environment. 

### How to deploy
The webhook is deployed along with the Pravega operator, thus there is no extra steps needed. However, there are some configurations that are necessary to make webhook work.

1. Permission

It is necessary to have permissions for `admissionregistration.k8s.io/v1beta1` resource to configure the webhook. The below is
an example of the additional permission
```
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  verbs:
  - '*'
```

2. Webhook service label selector

The webhook will deploy a Kubernetes service. This service will need to select the operator pod as its backend.
The way to select is using Kubernetes label selector and user will need to specify `"component": "pravega-operator"` as the label
when deploying the Pravega operator deployment. 
```

### What it does
The webhook maintains a compatibility matrix of the Pravega versions. Reuqests will be rejected if the version is not valid or not upgrade compatible 
with the current running version. Also, all the upgrade requests will be rejected if the current cluster is in upgrade status.  


