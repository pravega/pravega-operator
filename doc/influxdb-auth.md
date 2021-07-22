# Enable InfluxDB Authentication

Operator supports passing influxdb credentials as secret. It is the recommended approach rather than passing username/password as part of Pravega options.

Steps to configure influxdb authentication are as follows:

1. Create a secret for basic authentication. Below is the sample yaml file

```
apiVersion: v1
kind: Secret
metadata:
  name: secret-basic-auth
type: kubernetes.io/basic-auth
stringData:
  username: admin
  password: t0p-Secret
  ```
2. Modify the Pravega manifest to include the secret name for influxdb

```
influxDBSecret:
  name: "secret-basic-auth"
  path: "/etc/secret"
```

3.  Once Pravega is deployed, secret will be mounted in  `/etc/secret` for controller and segment store pods. If the path is not mentioned, `/etc/influxdb-secret-volume` will be used as default path.
