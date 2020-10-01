## Admission Webhook

[Admission webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) are HTTP callbacks that receive admission requests and do something with them.
There are  two webhooks [ValidatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#validatingadmissionwebhook) and
[MutatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) which are basically doing the same thing except MutatingAdmissionWebhook can modify the requests. In our case, we are using a ValidatingAdmissionWebhook so that it can reject requests to enforce custom policies (which in our case is to ensure that the user is unable to install an invalid pravega version or upgrade to any unsupported pravega version).

In the Pravega operator repo, we are leveraging the webhook implementation from controller-runtime package, here is the [GoDoc](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/webhook).

If you want to implement admission webhooks for your CRD, the only thing you need to do is to implement the `Defaulter` and (or) the `Validator` interface. Kubebuilder takes care of the rest for you, such as:
- Creating the webhook server.
- Ensuring the server has been added in the manager.
- Creating handlers for your webhooks.
- Registering each handler with a path in your server.
The webhook server registers webhook configuration with the apiserver and creates an HTTP server to route requests to the handlers.
The server is behind a Kubernetes Service and provides a certificate to the apiserver when serving requests.
The kubebuilder has a detailed instruction of building a webhook, see [here](https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html)

The webhook feature itself is enabled by default but it can be disabled if `webhook=false` is specified when installing the
operator locally using `operator-sdk run --local`. E.g. `operator-sdk run --local --operator-flags -webhook=false`. The use case of this is that webhook needs to be disabled when developing the operator locally since webhook can only be deployed in Kubernetes environment.

### How to deploy
The ValidatingAdmissionWebhook and the webhook service should be deployed using the provided manifest `webhook.yaml` while deploying the Pravega Operator. However, there are some configurations that are necessary to make webhook work.

1. Permission

It is necessary to have permissions for `admissionregistration.k8s.io/v1beta1` resource to configure the webhook. The below is
an example of the additional permission
```
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - validatingwebhookconfigurations
  verbs:
  - '*'
```

2. Webhook service label selector

The webhook will deploy a Kubernetes service. This service will need to select the operator pod as its backend.
The way to select is using Kubernetes label selector and user will need to specify `"component": "pravega-operator"` as the label
when deploying the Pravega operator deployment.

### What it does
The webhook maintains a compatibility matrix of the Pravega versions. Requests will be rejected if the version is not valid or not upgrade compatible with the current running version. Also, all the upgrade requests will be rejected if the current cluster is in upgrade status.  
