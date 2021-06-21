## AuthImplementations

If we need to enable auth handlers in Pravega controller, modify the Pravega manifest and ensure that following fields are present in the yaml file.

```
authImplementations:
  mountPath: "/opt/pravega/pluginlib"
  authHandlers:
  - image: authHandlerImage
    source: "/some_vendor/data/*"
```  

User has to provide an image that contains authhandler implementation and the image should have `/bin/sh` enabled. Also need to ensure `source` provided in the `yaml` should be the correct location of `jar` file and should have sufficient permissions to copy the file. `MountPath` should be present in Pravega class path and currently  `/opt/pravega/pluginlib` is added in class path of Pravega controller.

Pravega performs dynamic implementations of the Authorization/Authentication [API](https://github.com/pravega/pravega/blob/master/documentation/src/docs/auth/auth-plugin.md). Multiple implementations can co-exist and different plugins can be used by different users.
