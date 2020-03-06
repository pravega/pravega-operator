# Enable Authentication

Pravega supports pluggable authentication and authorization (referred to as auth for short). For details please see [Pravega Authentication](https://github.com/pravega/pravega/blob/master/documentation/src/docs/auth/auth-plugin.md).

By default, the PasswordAuthHandler plugin is installed on the system.
To use the default `PasswordAuthHandler` plugin for `auth`, the following steps can be followed:

1. Create a file containing `<user>:<password>:<acl>;` with one line per user.
Delimiter should be `:` with `;` at the end of each line.
Use the   [PasswordCreatorTool](https://github.com/pravega/pravega/blob/master/controller/src/test/java/io/pravega/controller/auth/PasswordFileCreatorTool.java) to create a new file with the password encrypted.
Use this file when creating the secret in next step.

Sample encrypted password file:
```
$ cat userdata.txt
admin:353030303a633132666135376233353937356534613430383430373939343839333733616463363433616532363238653930346230333035393666643961316264616661393a3639376330623663396634343864643262663335326463653062613965336439613864306264323839633037626166663563613166333733653631383732353134643961303435613237653130353633633031653364366565316434626534656565636335663666306465663064376165313765646263656638373764396361:*,READ_UPDATE;
```

2. Create a kubernetes secret with this file:

```
$ kubectl create secret generic password-auth \
  --from-file=./userdata.txt \
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
userdata.txt:  418 bytes
```

3. Specify the secret names in the `authentication` block and these parameters in the `options` block.

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
      controller.auth.userPasswordFile: "/etc/auth-passwd-volume/userdata.txt"
      controller.auth.tokenSigningKey: "secret"
      autoScale.authEnabled: "true"
      autoScale.tokenSigningKey: "secret"
      pravega.client.auth.token: "YWRtaW46MTExMV9hYWFh"
      pravega.client.auth.method: "Basic"

...
```

`pravega.client.auth.method` and `pravega.client.auth.token` represent the auth method and token to be used for internal communications from the Segment Store to the Controller.
If you intend to use the default auth plugin, these values are:
```
pravega.client.auth.method: Basic
pravega.client.auth.token: Base64 encoded value of <username>:<pasword>,
```
where username and password are credentials you intend to use.

Note that Pravega operator uses `/etc/auth-passwd-volume` as the mounting directory for secrets.

For more security configurations, please check [here](https://github.com/pravega/pravega/blob/master/documentation/src/docs/security/pravega-security-configurations.md).
