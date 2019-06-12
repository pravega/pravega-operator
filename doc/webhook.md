## Admission Webhook

[Admission webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) are HTTP callbacks that receive admission requests and do something with them.
There are  two webhooks [ValidatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#validatingadmissionwebhook) and 
[MutatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) which are basically 
doing the same thing except MutatingAdmissionWebhook can modify the requests. In our case, we use MutatingAdmissionWebhook.

In the Pravega operator repo, we are leveraging the webhook implementation from controller-runtime package, here is the [GoDoc](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/webhook). 
In detail, there are two steps that developers need to do 1) create webhook server and 2) implement the handler.
The webhook server registers webhook configuration with the apiserver and creates an HTTP server to route requests to the handlers.
The server is behind a Kubernetes Service and provides a certificate to the apiserver when serving requests. The kubebuilder has a detailed instruction of 
building a webhook, see [here](https://github.com/kubernetes-sigs/kubebuilder/blob/86026527c754a144defa6474af6fb352143b9270/docs/book/beyond_basics/sample_webhook.md).

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

2. Namespace

The Kubernetes service needs to be created in the same namespace with the webhook server. 
Kuberenetes uses the [downawrd api](https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/#the-downward-api)
to expose pod information to containers. Here is how it get passed in the `operator.yaml` file.
```
env:
- name: WEBHOOK_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
```

### What it does
The webhook maintains a compatibility matrix of the Pravega versions. Reuqests will be rejected if the version is not valid or not upgrade compatible 
with the current running version. Also, all the upgrade requests will be rejected if the current cluster is in upgrade status.  


