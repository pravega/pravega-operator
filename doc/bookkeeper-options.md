## Bookkeeper options

Bookkeeper has many configuration options. The available options can be found [here](https://bookkeeper.apache.org/docs/4.7.0/reference/config/) and are expressed through the `bookkeeper/options` part of the resource specification.

All values must be expressed as Strings.

Take metrics for example, here we choose codahale as our metrics provider. The default is Prometheus.

```
...
spec:
  bookeeper:
    options:
      enableStatistics: "true"
      statsProviderClass: "org.apache.bookkeeper.stats.codahale.CodahaleMetricsProvider"
      codahaleStatsGraphiteEndpoint: "graphite.example.com:2003"
      codahaleStatsOutputFrequencySeconds: "30"
...
```
### Bookkeeper JVM Options

It is also possible to tune the Bookkeeper JVM by passing customized JVM options. Bookkeeper JVM Options
are obvisouly for Bookkeeper JVM whereas aforementioned BookKeeper options are for BookKeeper server configuration.

The format is as follows:
```
...
spec:
  bookkeeper:
    bookkeeperJVMOptions:
      memoryOpts: ["-Xms2g", "-XX:MaxDirectMemorySize=2g"]
      gcOpts: ["-XX:MaxGCPauseMillis=20"]
      gcLoggingOpts: ["-XX:NumberOfGCLogFiles=10"]
      extraOpts: []
...
```
The reason that we are using such detailed names like `memoryOpts` is because the Bookkeeper official [scripts](https://github.com/apache/bookkeeper/blob/master/bin/common.sh#L118) are using those and we need to override it using the same name. JVM options that don't belong to the earlier 3 categories can be mentioned under `extraOpts`.

There are a bunch of default options in the Pravega operator code that is good for general deployment. It is possible to override those default values by just passing the customized options. For example, the default option `"-XX:MaxDirectMemorySize=1g"` can be override by passing `"-XX:MaxDirectMemorySize=2g"` to
the Pravega operator. The operator will detect `MaxDirectMemorySize` and override its default value if it exists. Check [here](https://www.oracle.com/technetwork/java/javase/tech/vmoptions-jsp-140102.html) for more JVM options.

Default memoryOpts:
```
"-Xms1g",
"-XX:MaxDirectMemorySize=1g",
"-XX:+ExitOnOutOfMemoryError",
"-XX:+CrashOnOutOfMemoryError",
"-XX:+HeapDumpOnOutOfMemoryError",
"-XX:HeapDumpPath=" + heapDumpDir,
```
if Pravega version is greater or equal to 0.4, then the followings are also added to the default memoryOpts
```
"-XX:+UnlockExperimentalVMOptions",
"-XX:+UseCGroupMemoryLimitForHeap",
"-XX:MaxRAMFraction=2"
```

Default gcOpts:
```
"-XX:+UseG1GC",
"-XX:MaxGCPauseMillis=10",
"-XX:+ParallelRefProcEnabled",
"-XX:+AggressiveOpts",
"-XX:+DoEscapeAnalysis",
"-XX:ParallelGCThreads=32",
"-XX:ConcGCThreads=32",
"-XX:G1NewSizePercent=50",
"-XX:+DisableExplicitGC",
"-XX:-ResizePLAB",
```

Default gcLoggingOpts:
```
"-XX:+PrintGCDetails",
"-XX:+PrintGCDateStamps",
"-XX:+PrintGCApplicationStoppedTime",
"-XX:+UseGCLogFileRotation",
"-XX:NumberOfGCLogFiles=5",
"-XX:GCLogFileSize=64m",
```
