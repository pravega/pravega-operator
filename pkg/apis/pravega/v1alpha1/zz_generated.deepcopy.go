/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

// +build !ignore_autogenerated

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BookkeeperSpec) DeepCopyInto(out *BookkeeperSpec) {
	*out = *in
	out.Image = in.Image
	in.Storage.DeepCopyInto(&out.Storage)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BookkeeperSpec.
func (in *BookkeeperSpec) DeepCopy() *BookkeeperSpec {
	if in == nil {
		return nil
	}
	out := new(BookkeeperSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BookkeeperStorageSpec) DeepCopyInto(out *BookkeeperStorageSpec) {
	*out = *in
	in.LedgerVolumeClaimTemplate.DeepCopyInto(&out.LedgerVolumeClaimTemplate)
	in.JournalVolumeClaimTemplate.DeepCopyInto(&out.JournalVolumeClaimTemplate)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BookkeeperStorageSpec.
func (in *BookkeeperStorageSpec) DeepCopy() *BookkeeperStorageSpec {
	if in == nil {
		return nil
	}
	out := new(BookkeeperStorageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ECSSpec) DeepCopyInto(out *ECSSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ECSSpec.
func (in *ECSSpec) DeepCopy() *ECSSpec {
	if in == nil {
		return nil
	}
	out := new(ECSSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileSystemSpec) DeepCopyInto(out *FileSystemSpec) {
	*out = *in
	out.PersistentVolumeClaim = in.PersistentVolumeClaim
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileSystemSpec.
func (in *FileSystemSpec) DeepCopy() *FileSystemSpec {
	if in == nil {
		return nil
	}
	out := new(FileSystemSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HDFSSpec) DeepCopyInto(out *HDFSSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HDFSSpec.
func (in *HDFSSpec) DeepCopy() *HDFSSpec {
	if in == nil {
		return nil
	}
	out := new(HDFSSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImageSpec) DeepCopyInto(out *ImageSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImageSpec.
func (in *ImageSpec) DeepCopy() *ImageSpec {
	if in == nil {
		return nil
	}
	out := new(ImageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricsSpec) DeepCopyInto(out *MetricsSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricsSpec.
func (in *MetricsSpec) DeepCopy() *MetricsSpec {
	if in == nil {
		return nil
	}
	out := new(MetricsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PravegaCluster) DeepCopyInto(out *PravegaCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PravegaCluster.
func (in *PravegaCluster) DeepCopy() *PravegaCluster {
	if in == nil {
		return nil
	}
	out := new(PravegaCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PravegaCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PravegaClusterList) DeepCopyInto(out *PravegaClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PravegaCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PravegaClusterList.
func (in *PravegaClusterList) DeepCopy() *PravegaClusterList {
	if in == nil {
		return nil
	}
	out := new(PravegaClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PravegaClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PravegaClusterSpec) DeepCopyInto(out *PravegaClusterSpec) {
	*out = *in
	in.Bookkeeper.DeepCopyInto(&out.Bookkeeper)
	in.Pravega.DeepCopyInto(&out.Pravega)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PravegaClusterSpec.
func (in *PravegaClusterSpec) DeepCopy() *PravegaClusterSpec {
	if in == nil {
		return nil
	}
	out := new(PravegaClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PravegaClusterStatus) DeepCopyInto(out *PravegaClusterStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PravegaClusterStatus.
func (in *PravegaClusterStatus) DeepCopy() *PravegaClusterStatus {
	if in == nil {
		return nil
	}
	out := new(PravegaClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PravegaSpec) DeepCopyInto(out *PravegaSpec) {
	*out = *in
	out.Image = in.Image
	out.Metrics = in.Metrics
	if in.Options != nil {
		in, out := &in.Options, &out.Options
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.CacheVolumeClaimTemplate.DeepCopyInto(&out.CacheVolumeClaimTemplate)
	in.Tier2.DeepCopyInto(&out.Tier2)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PravegaSpec.
func (in *PravegaSpec) DeepCopy() *PravegaSpec {
	if in == nil {
		return nil
	}
	out := new(PravegaSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Tier2Spec) DeepCopyInto(out *Tier2Spec) {
	*out = *in
	if in.FileSystem != nil {
		in, out := &in.FileSystem, &out.FileSystem
		if *in == nil {
			*out = nil
		} else {
			*out = new(FileSystemSpec)
			**out = **in
		}
	}
	if in.Ecs != nil {
		in, out := &in.Ecs, &out.Ecs
		if *in == nil {
			*out = nil
		} else {
			*out = new(ECSSpec)
			**out = **in
		}
	}
	if in.Hdfs != nil {
		in, out := &in.Hdfs, &out.Hdfs
		if *in == nil {
			*out = nil
		} else {
			*out = new(HDFSSpec)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tier2Spec.
func (in *Tier2Spec) DeepCopy() *Tier2Spec {
	if in == nil {
		return nil
	}
	out := new(Tier2Spec)
	in.DeepCopyInto(out)
	return out
}
