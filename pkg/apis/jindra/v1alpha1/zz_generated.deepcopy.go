// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JindraPipeline) DeepCopyInto(out *JindraPipeline) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JindraPipeline.
func (in *JindraPipeline) DeepCopy() *JindraPipeline {
	if in == nil {
		return nil
	}
	out := new(JindraPipeline)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *JindraPipeline) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JindraPipelineList) DeepCopyInto(out *JindraPipelineList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]JindraPipeline, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JindraPipelineList.
func (in *JindraPipelineList) DeepCopy() *JindraPipelineList {
	if in == nil {
		return nil
	}
	out := new(JindraPipelineList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *JindraPipelineList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JindraPipelineResources) DeepCopyInto(out *JindraPipelineResources) {
	*out = *in
	if in.Triggers != nil {
		in, out := &in.Triggers, &out.Triggers
		*out = make([]JindraPipelineResourcesTrigger, len(*in))
		copy(*out, *in)
	}
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = make([]v1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JindraPipelineResources.
func (in *JindraPipelineResources) DeepCopy() *JindraPipelineResources {
	if in == nil {
		return nil
	}
	out := new(JindraPipelineResources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JindraPipelineResourcesTrigger) DeepCopyInto(out *JindraPipelineResourcesTrigger) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JindraPipelineResourcesTrigger.
func (in *JindraPipelineResourcesTrigger) DeepCopy() *JindraPipelineResourcesTrigger {
	if in == nil {
		return nil
	}
	out := new(JindraPipelineResourcesTrigger)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JindraPipelineSpec) DeepCopyInto(out *JindraPipelineSpec) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
	if in.Stages != nil {
		in, out := &in.Stages, &out.Stages
		*out = make([]v1.Pod, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.OnSuccess.DeepCopyInto(&out.OnSuccess)
	in.OnError.DeepCopyInto(&out.OnError)
	in.Final.DeepCopyInto(&out.Final)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JindraPipelineSpec.
func (in *JindraPipelineSpec) DeepCopy() *JindraPipelineSpec {
	if in == nil {
		return nil
	}
	out := new(JindraPipelineSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JindraPipelineStatus) DeepCopyInto(out *JindraPipelineStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JindraPipelineStatus.
func (in *JindraPipelineStatus) DeepCopy() *JindraPipelineStatus {
	if in == nil {
		return nil
	}
	out := new(JindraPipelineStatus)
	in.DeepCopyInto(out)
	return out
}
