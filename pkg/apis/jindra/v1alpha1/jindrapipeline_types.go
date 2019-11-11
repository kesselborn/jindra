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

// SetDefaults sets sane default values if not set already
func (ppl *JindraPipeline) SetDefaults() []interface{} {
	modifiedItems := []interface{}{}
	// only set name if final is actually set (annotations or containers are set)
	if (len(ppl.Spec.Final.Annotations) > 0 || len(ppl.Spec.Final.Spec.Containers) > 0) && ppl.Spec.Final.Name == "" {
		modifiedItems = append(modifiedItems, "final/name", "final")
		ppl.Spec.Final.Name = "final"
	}

	// only set name if on-error is actually set (annotations or containers are set)
	if (len(ppl.Spec.OnError.Annotations) > 0 || len(ppl.Spec.OnError.Spec.Containers) > 0) && ppl.Spec.OnError.Name == "" {
		modifiedItems = append(modifiedItems, "on-error/name", "on-error")
		ppl.Spec.OnError.Name = "on-error"
	}

	// only set name if on-success is actually set (annotations or containers are set)
	if (len(ppl.Spec.OnSuccess.Annotations) > 0 || len(ppl.Spec.OnSuccess.Spec.Containers) > 0) && ppl.Spec.OnSuccess.Name == "" {
		modifiedItems = append(modifiedItems, "on-success/name", "on-success")
		ppl.Spec.OnSuccess.Name = "on-success"
	}

	for i := 0; i < len(ppl.Spec.Resources.Triggers); i++ {
		if ppl.Spec.Resources.Triggers[i].Schedule == "" {
			modifiedItems = append(modifiedItems, "trigger/"+ppl.Spec.Resources.Triggers[i].Name+"/name/schedule", "/5 * * * *")
			ppl.Spec.Resources.Triggers[i].Schedule = "/5 * * * *"
		}
	}

	if ppl.Annotations[BuildNoOffsetAnnotationKey] == "" {
		modifiedItems = append(modifiedItems, "/build-offset", "0")
		ppl.Annotations[BuildNoOffsetAnnotationKey] = "0"
	}

	for i := 0; i < len(ppl.Spec.Stages); i++ {
		if ppl.Spec.Stages[i].Spec.RestartPolicy == core.RestartPolicy("") {
			modifiedItems = append(modifiedItems, "/stage/"+ppl.Spec.Stages[i].Name+"/restart-policy", "never")
			ppl.Spec.Stages[i].Spec.RestartPolicy = core.RestartPolicyNever
		}

	}
	for _, pod := range []*core.Pod{&ppl.Spec.OnSuccess, &ppl.Spec.OnError, &ppl.Spec.Final} {
		if (len(pod.Spec.Containers) > 0) && pod.Spec.RestartPolicy == core.RestartPolicy("") {
			modifiedItems = append(modifiedItems, "/stage/"+pod.Name+"/restart-policy", "never")
			pod.Spec.RestartPolicy = core.RestartPolicyNever
		}
	}

	return modifiedItems
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
