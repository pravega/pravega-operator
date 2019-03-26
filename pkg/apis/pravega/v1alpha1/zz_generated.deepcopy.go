// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BookkeeperSpec) DeepCopyInto(out *BookkeeperSpec) {
	*out = *in
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		*out = new(BookkeeperStorageSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.AutoRecovery != nil {
		in, out := &in.AutoRecovery, &out.AutoRecovery
		*out = new(bool)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.Options != nil {
		in, out := &in.Options, &out.Options
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
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
	if in.LedgerVolumeClaimTemplate != nil {
		in, out := &in.LedgerVolumeClaimTemplate, &out.LedgerVolumeClaimTemplate
		*out = new(v1.PersistentVolumeClaimSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.JournalVolumeClaimTemplate != nil {
		in, out := &in.JournalVolumeClaimTemplate, &out.JournalVolumeClaimTemplate
		*out = new(v1.PersistentVolumeClaimSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.IndexVolumeClaimTemplate != nil {
		in, out := &in.IndexVolumeClaimTemplate, &out.IndexVolumeClaimTemplate
		*out = new(v1.PersistentVolumeClaimSpec)
		(*in).DeepCopyInto(*out)
	}
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
func (in *ClusterCondition) DeepCopyInto(out *ClusterCondition) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterCondition.
func (in *ClusterCondition) DeepCopy() *ClusterCondition {
	if in == nil {
		return nil
	}
	out := new(ClusterCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSpec) DeepCopyInto(out *ClusterSpec) {
	*out = *in
	if in.ExternalAccess != nil {
		in, out := &in.ExternalAccess, &out.ExternalAccess
		*out = new(ExternalAccess)
		**out = **in
	}
	if in.Bookkeeper != nil {
		in, out := &in.Bookkeeper, &out.Bookkeeper
		*out = new(BookkeeperSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Pravega != nil {
		in, out := &in.Pravega, &out.Pravega
		*out = new(PravegaSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSpec.
func (in *ClusterSpec) DeepCopy() *ClusterSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterStatus) DeepCopyInto(out *ClusterStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ClusterCondition, len(*in))
		copy(*out, *in)
	}
	in.Members.DeepCopyInto(&out.Members)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterStatus.
func (in *ClusterStatus) DeepCopy() *ClusterStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterStatus)
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
func (in *ExternalAccess) DeepCopyInto(out *ExternalAccess) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalAccess.
func (in *ExternalAccess) DeepCopy() *ExternalAccess {
	if in == nil {
		return nil
	}
	out := new(ExternalAccess)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileSystemSpec) DeepCopyInto(out *FileSystemSpec) {
	*out = *in
	if in.PersistentVolumeClaim != nil {
		in, out := &in.PersistentVolumeClaim, &out.PersistentVolumeClaim
		*out = new(v1.PersistentVolumeClaimVolumeSource)
		**out = **in
	}
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
func (in *MembersStatus) DeepCopyInto(out *MembersStatus) {
	*out = *in
	if in.Ready != nil {
		in, out := &in.Ready, &out.Ready
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Unready != nil {
		in, out := &in.Unready, &out.Unready
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MembersStatus.
func (in *MembersStatus) DeepCopy() *MembersStatus {
	if in == nil {
		return nil
	}
	out := new(MembersStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PravegaCluster) DeepCopyInto(out *PravegaCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
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
	}
	return nil
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
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PravegaSpec) DeepCopyInto(out *PravegaSpec) {
	*out = *in
	if in.Options != nil {
		in, out := &in.Options, &out.Options
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.CacheVolumeClaimTemplate != nil {
		in, out := &in.CacheVolumeClaimTemplate, &out.CacheVolumeClaimTemplate
		*out = new(v1.PersistentVolumeClaimSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Tier2 != nil {
		in, out := &in.Tier2, &out.Tier2
		*out = new(Tier2Spec)
		(*in).DeepCopyInto(*out)
	}
	if in.ControllerResources != nil {
		in, out := &in.ControllerResources, &out.ControllerResources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.SegmentStoreResources != nil {
		in, out := &in.SegmentStoreResources, &out.SegmentStoreResources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
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
		*out = new(FileSystemSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Ecs != nil {
		in, out := &in.Ecs, &out.Ecs
		*out = new(ECSSpec)
		**out = **in
	}
	if in.Hdfs != nil {
		in, out := &in.Hdfs, &out.Hdfs
		*out = new(HDFSSpec)
		**out = **in
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
