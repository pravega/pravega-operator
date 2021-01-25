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
      controller.security.auth.enable: "true"
      controller.security.pwdAuthHandler.accountsDb.location: "/etc/auth-passwd-volume/userdata.txt"
      controller.security.auth.delegationToken.signingKey.basis: "secret"
      autoScale.controller.connect.security.auth.enable: "true"
      autoScale.security.auth.token.signingKey.basis: "secret"
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

Pravega Operator Supports Passing of Auth Parametes as Secret which are mounted as file in both Segment Store and Controller (Operator mounts these secrets in Segementstore pod it's mounted at `/etc/ss-auth-volume` and in Controller pod at `/etc/controller-auth-volume` respectively)

Note that Pravega has to use this feature and start Picking below specified values from file insted of jvm properties which it currently does.
(This is not implemented at pravega end currently)

Below is how we can create secret and expose them as file for Auth related properties:-

1. Create a File containg `controller.security.auth.delegationToken.signingKey.basis`  as `delegationToken.signingKey.basis` which represent the tokensigning key used to connect to the controller:

```
$ cat controllerauthdata.txt
delegationToken.signingKey.basis: "secret"
```

2. Create a kubernetes secret with this file:

```
$ kubectl create secret generic controllertokensecret \
  --from-file=./controllerauthdata.txt \
```

Ensure Secret is created:-

```
$ kubectl describe secret controllertokensecret
Name:         controllertokensecret
Namespace:    default
Labels:       <none>
Annotations:  <none>

Type:  Opaque

Data
====
controllerauthdata.txt:  67 bytes

```

3. Create a File containg `autoScale.security.auth.token.signingKey.basis` as `delegationToken.signingKey.basis` which represent the tokensigning key used to connect to the Segmentstore along with other 3 values `pravega.client.auth.method`, `pravega.client.auth.token` and `controller.connect.auth.credentials.dynamic` as `controller.connect.auth.params` which contains all the 3 values in the same order seprated by semicolon:

```
$ cat segmentstoreauthdata.txt
delegationToken.signingKey.basis: "secret"
controller.connect.auth.params: {method};{token};{dynamic}
```
4. Create a kubernetes secret with this file:

```
$ kubectl create secret generic sstokensecret \
  --from-file=./segmentstoreauthdata.txt \
```
Ensure Secret is created:-

```
$ kubectl describe secret sstokensecret
Name:         sstokensecret
Namespace:    default
Labels:       <none>
Annotations:  <none>

Type:  Opaque

Data
====
segmentstoreauthdata.txt:  106 bytes

```

5. Use these secrets instead of specifying the values in the option:-

```
spec:
  authentication:
    enabled: true
    passwordAuthSecret: password-auth
    segmentStoreTokenSecret: sstokensecret
    controllerTokenSecret: controllertokensecret
```    
