## Init Containers

Users can mention the details in init containers to perform certain operations before the main container starts. This is supported in controller as well as segmentstore.

To add init containers in controller modify the  Pravega manifest and ensure that following fields are present.

```
controller:
  initContainers:
  - name: "initscripts"
    image: <image name>
    command: ['sh', '-c', 'echo The app is running! && sleep 60']
```

 Also, we can mention the details in init containers for copying files to controller or segmentstore pods. For that modify the Pravega manifest and ensure that following fields are present. We need to ensure that `/bin/sh` is enabled on the image provided and should have sufficient permissions on the files to perform  copy operation.

```
segmentStore:
  initContainers:
  - name: "initscripts"
    image: <image containing scripts>
    command: ['sh', '-c','cp /data/scripts/* /opt/scripts']
    volumeMounts:
    - mountPath: /opt/scripts
      name: scripts
```
Also, we need to ensure that main container is mounted and emptydir volume is created at the pod. This can be achieved by mentioning `emptyDirVolumeMounts` in options as below.

```
options:
  emptyDirVolumeMounts: "scripts=/opt/scripts"
```
