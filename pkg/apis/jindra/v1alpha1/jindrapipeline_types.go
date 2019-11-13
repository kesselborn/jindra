package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JindraPipelineResourcesTrigger defines a pipeline trigger and the cron schedule when the checks
// for new versions should be done
// +k8s:openapi-gen=true
type JindraPipelineResourcesTrigger struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
}

// JindraPipelineResources defines a pipeline resources
// for new versions should be done
// +k8s:openapi-gen=true
type JindraPipelineResources struct {
	Triggers   []JindraPipelineResourcesTrigger `json:"triggers"`
	Containers []core.Container                 `json:"containers"`
}

// JindraPipelineSpec defines the desired state of JindraPipeline
// +k8s:openapi-gen=true
type JindraPipelineSpec struct {
	Resources JindraPipelineResources `json:"resources,omitempty"`
	Stages    []core.Pod              `json:"stages"`
	OnSuccess core.Pod                `json:"onSuccess,omitempty"`
	OnError   core.Pod                `json:"onError,omitempty"`
	Final     core.Pod                `json:"final,omitempty"`
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// JindraPipelineStatus defines the observed state of JindraPipeline
// +k8s:openapi-gen=true
type JindraPipelineStatus struct {
	BuildNo int `json:"buildNo"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JindraPipeline is the Schema for the jindrapipelines API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=jindrapipelines,scope=Namespaced
type JindraPipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JindraPipelineSpec   `json:"spec,omitempty"`
	Status JindraPipelineStatus `json:"status,omitempty"`
}

// Validate validates the correctnes of the JindraPipeline object
func (ppl JindraPipeline) Validate() error {
	for _, f := range []func() error{
		ppl.correctOrNoRestartPolicy,
		ppl.noDuplicateResourceAnnotations,
		ppl.noDuplicateResourceNames,
		ppl.noOwnerReference,
		ppl.resourcesExist,
		ppl.serviceExist,
		ppl.triggerHasResource,
		ppl.triggerIsInResourceOfFirstStage,
	} {
		if err := f(); err != nil {
			return err
		}
	}

	return nil
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JindraPipelineList contains a list of JindraPipeline
type JindraPipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JindraPipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JindraPipeline{}, &JindraPipelineList{})
}
