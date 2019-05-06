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
