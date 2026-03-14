package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyObject implements runtime.Object
func (in *ManagedService) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopy returns a deep copy of the ManagedService
func (in *ManagedService) DeepCopy() *ManagedService {
	if in == nil {
		return nil
	}
	out := new(ManagedService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all fields into the target ManagedService
func (in *ManagedService) DeepCopyInto(out *ManagedService) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopyObject implements runtime.Object for ManagedServiceList
func (in *ManagedServiceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopy returns a deep copy of the ManagedServiceList
func (in *ManagedServiceList) DeepCopy() *ManagedServiceList {
	if in == nil {
		return nil
	}
	out := new(ManagedServiceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all fields into the target ManagedServiceList
func (in *ManagedServiceList) DeepCopyInto(out *ManagedServiceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ManagedService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}