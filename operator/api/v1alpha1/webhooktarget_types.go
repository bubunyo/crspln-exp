package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ runtime.Object = &WebhookTarget{}
var _ runtime.Object = &WebhookTargetList{}

type WebhookTargetSpec struct{}

type WebhookTargetStatus struct {
	Endpoint string `json:"endpoint,omitempty"`
	Received int    `json:"received,omitempty"`
}

type WebhookTarget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebhookTargetSpec   `json:"spec,omitempty"`
	Status WebhookTargetStatus `json:"status,omitempty"`
}

type WebhookTargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []WebhookTarget `json:"items"`
}

func (in *WebhookTarget) DeepCopyInto(out *WebhookTarget) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

func (in *WebhookTarget) DeepCopy() *WebhookTarget {
	if in == nil {
		return nil
	}
	out := new(WebhookTarget)
	in.DeepCopyInto(out)
	return out
}

func (in *WebhookTarget) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *WebhookTargetList) DeepCopyInto(out *WebhookTargetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		out.Items = make([]WebhookTarget, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *WebhookTargetList) DeepCopy() *WebhookTargetList {
	if in == nil {
		return nil
	}
	out := new(WebhookTargetList)
	in.DeepCopyInto(out)
	return out
}

func (in *WebhookTargetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
