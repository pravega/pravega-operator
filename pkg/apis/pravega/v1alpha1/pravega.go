package v1alpha1

import (
	"k8s.io/api/core/v1"
)

type PravegaSpec struct {
	ControllerReplicas       int32                        `json:",controllerReplicas"`
	SegmentStoreReplicas     int32                        `json:",segmentStoreReplicas"`
	DebugLogging             bool                         `json:",debugLogging"`
	Image                    ImageSpec                    `json:",image"`
	Metrics                  MetricsSpec                  `json:",metrics"`
	Options                  map[string]string            `json:",options"`
	CacheVolumeClaimTemplate v1.PersistentVolumeClaimSpec `json:",cacheVolumeClaimTemplate"`
	Tier2                    Tier2Spec                    `json:",tier2"`
}

type MetricsSpec struct {
}

type Tier2Spec struct {
	FileSystem *FileSystemSpec `json:",filesystem"`
	Ecs        *ECSSpec        `json:",ecs"`
	Hdfs       *HDFSSpec       `json:",hdfs"`
}

type FileSystemSpec struct {
	PersistentVolumeClaim v1.PersistentVolumeClaimVolumeSource `json:",persistentVolumeClaim"`
}

type ECSSpec struct {
	Uri         string `json:",uri"`
	Bucket      string `json:",bucket"`
	Root        string `json:",root"`
	Namespace   string `json:",namespace"`
	Credentials string `json:",credentials"`
}

type HDFSSpec struct {
	Uri               string `json:",uri"`
	Root              string `json:",root"`
	ReplicationFactor int32  `json:",replicationFactor"`
}
