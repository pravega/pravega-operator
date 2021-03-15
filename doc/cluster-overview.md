# Cluster Overview:

Pravega is a storage system for data streams that has an innovative design and an attractive set of features to cope with today’s Stream processing requirements (e.g., event ordering, scalability, performance, etc.). For deploying Pravega, we have to install the following components as a pre-requisite. Pravega can be deployed in different environments like OpenShift, KubeSpray, GKE, EKS etc

Zookeeper

It is a distributed system that provides reliable coordination services, such as consensus and group management. Pravega uses Zookeeper to store specific pieces of metadata as well as to offer a consistent view of data structures used by multiple service instances.This can be insalled via zookeeper-operator

Bookkeeper

It is a distributed and reliable storage system that provides a distributed log abstraction. Bookkeeper excels on achieving low latency, append-only writes. This is the reason why Pravega uses Bookkeeper for journaling: Pravega writes data to Bookkeeper, which provides low latency, persistent, and replicated storage for stream appends. Pravega uses the data in BookKeeper to recover from failures, and that data is truncated once it is flushed to tiered long-term storage.This can be installed via Bookkeeper-operator

LongTermStorage

 It provides long term storage for Stream data.Pravega automatically moves data to LongTermStorage. LongTermStorage can be NFS, ECS or HDFS
