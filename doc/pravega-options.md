## Pravega options

Pravega has many configuration options for setting up metrics, tuning, etc. The available options can be found [here](https://github.com/pravega/pravega/blob/master/config/config.properties) and are expressed through the pravega/options part of the resource specification.

All values must be expressed as Strings.

```
...
spec:
  pravega:
    options:
      metrics.enableStatistics: "true"
      metrics.statsdHost: "telegraph.default"
      metrics.statsdPort: "8125"
...
```
### Pravega JVM Options

It is also possible to tune the JVM options for Pravega Controller and Segmentstore. Pravega JVM options are for configuring Controller&Segmenstore JVM process whereas Pravega options are for configuring Pravega software.

Here is an example,
```
...
spec:
  pravega:
    controllerJvmOptions: ["-XX:MaxDirectMemorySize=1g"]
    segmentStoreJVMOptions: ["-XX:MaxDirectMemorySize=1g"]
...
```
There are a bunch of default options in the Pravega operator code that is good for general deployment, please check [here](https://github.com/pravega/pravega-operator/blob/master/pkg/controller/pravega/pravega_controller.go#L175) for controller and [here](https://github.com/pravega/pravega-operator/blob/master/pkg/controller/pravega/pravega_segmentstore.go#L161) for segmentstore. It is possible to override those default values by
just passing the customized options. For example, the default option `"-XX:MaxDirectMemorySize=1g"` can be override by passing `"-XX:MaxDirectMemorySize=2g"` to
the Pravega operator. The operator will detect `MaxDirectMemorySize` and override its default value if it exists. Check [here](https://www.oracle.com/technetwork/java/javase/tech/vmoptions-jsp-140102.html) for more JVM options.
