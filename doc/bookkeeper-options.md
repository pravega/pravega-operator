## Bookkeeper options
Bookkeeper has many configuration options.
The available options can be found [here](https://bookkeeper.apache.org/docs/4.7.0/reference/config/) and are expressed through the bookkeeper/options part of the resource specification. 
All values must be expressed as Strings.

Take metrics for example, here we choose codahale as our metrics provider. The default is Prometheus. 
```
options:
  enableStatistics: "true"
  statsProviderClass: "org.apache.bookkeeper.stats.codahale.CodahaleMetricsProvider"
  codahaleStatsGraphiteEndpoint: "graphite.example.com:2003"
  codahaleStatsOutputFrequencySeconds: "30"
```