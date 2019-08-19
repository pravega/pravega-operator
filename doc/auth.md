# Enable Authentication

Client can communicate with Pravega in a more secure way by enabling Authentication.
Authentication Handler can be implemented in different ways. For details see [Pravega AUthentication](https://github.com/pravega/pravega/blob/master/documentation/src/docs/auth/auth-plugin.md) for details.

PasswordAuthHandler(default) is the most basic authentication handler.
For using this, user needs to do the following:

1. Create a userPassword file containing `<user>:<password>:<acl>;` on each line for one user.
Delimiter should be `:` with `;` at the end of each line.
Use the   [PasswordCreatorTool](https://github.com/pravega/pravega/blob/master/controller/src/test/java/io/pravega/controller/auth/PasswordFileCreatorTool.java) to create a new file with the password encrypted. Use this file when creating the secret in next step.

Example:

cat userPassword.txt
```
admin:353030303a633132666135376233353937356534613430383430373939343839333733616463363433616532363238653930346230333035393666643961316264616661393a3639376330623663396634343864643262663335326463653062613965336439613864306264323839633037626166663563613166333733653631383732353134643961303435613237653130353633633031653364366565316434626534656565636335663666306465663064376165313765646263656638373764396361:*,READ_UPDATE;
```

2. Create a kubernetes secret with this file:

```
$ kubectl create secret generic password-auth \
  --from-file=./userpass.txt \
```

Ensure secret is created:
```
$ kubectl describe secret password-auth

Name:         password-auth
Namespace:    default
Labels:       <none>
Annotations:  <none>

Type:  Opaque

Data
====
userpass.txt:  418 bytes
```

3. Then specify the secret names in the `authentication` block and these parameters in the `options` block.

```
apiVersion: "pravega.pravega.io/v1alpha1"
kind: "PravegaCluster"
metadata:
  name: "example"
spec:
  authentication:
    enabled: true
    passwordAuthSecret: password-auth
...
  pravega:
    options:
      controller.auth.enabled: "true"
      controller.auth.userPasswordFile: "/etc/auth-passwd-volume/userpass.txt"
      controller.auth.tokenSigningKey: "secret"
      autoScale.authEnabled: "true"
      autoScale.tokenSigningKey: "secret"
      pravega.client.auth.token: "YWRtaW46MTExMV9hYWFh"
      pravega.client.auth.method: "Basic"

...
```

YWRtaW46MTExMV9hYWFh is base64encoded(admin:1111_aaaa) where admin is username and 1111_aaaa is the password to be used for segmentstore to controller communication.
Note that Pravega operator uses `/etc/auth-passwd-volume` as the mounting directory for secrets.

For more security configurations, check [here](https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md).
