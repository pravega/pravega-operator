## AuthImplementations

If we need to enable auth handlers in Pravega controller, modify the pravega manifest and ensure that following fields are present in the yaml file.

```
authImplementations:
  mountPath: "/opt/pravega/pluginlib"
  authHandlers:
 - image: authHandlerImage
   source: "/some_vendor/data/*"
```  

User has to provide an image that contains authhandler implementation and the image should have `/bin/sh` enabled.

Pravega performs dynamic implementations of the Authorization/Authentication [API](https://github.com/pravega/pravega/blob/master/documentation/src/docs/auth/auth-plugin.md). Multiple implementations can co-exist and different plugins can be used by different users.


Alternatively, user can mention the details in init Containers as well to copy the files to controller or segmentstore pods. For using init containers modify the  pravega manifest and ensure that following fields are present.

```
initContainers:
- name: "initscripts"
  image: <image containing scripts>
  command: ["cp /data/scripts/* /opt/scripts"]
  volumeMounts:
  - mountPath: /opt/scripts
    name: scripts
```
Also, we need to ensure that main container is mounted and emptydir volume is created at the pod. This can be achieved by mentioning `emptyDirVolumeMounts` in options as below.

```
options:
  emptyDirVolumeMounts: "scripts=/opt/scripts"
```
