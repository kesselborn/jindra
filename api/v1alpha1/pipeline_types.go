/*

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

package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Pipeline is the Schema for the pipelines API
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PipelineList contains a list of Pipeline
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

// PipelineSpec defines the desired state of Pipeline
type PipelineSpec struct {
	// Resources that can be used within the pipeline
	// +optional
	Resources Resources `json:"resources,omitempty"`

	// Definition of the stages of this pipeline. Each state is a pod definition
	// +kubebuilder:validation:EmbeddedResource
	Stages []core.Pod `json:"stages"`

	// Pod that should be executed if all stages finished successfully
	// +optional
	// +kubebuilder:validation:Optional
	OnSuccess *core.Pod `json:"onSuccess,omitempty"`

	// Pod that should be executed if an error ocurred in any stage
	// +optional
	// +kubebuilder:validation:Optional
	OnError *core.Pod `json:"onError,omitempty"`

	// Pod that should be executed if the pipeline finised (regardless whether it was
	// successful or not
	// +optional
	// +kubebuilder:validation:Optional
	Final *core.Pod `json:"final,omitempty"`
}

// Resources defines a pipeline resources for new versions should be done
// +k8s:openapi-gen=true
type Resources struct {
	// +optional
	Triggers []Trigger `json:"triggers"`

	// +kubebuilder:validation:EmbeddedResource
	Containers []core.Container `json:"containers"`
}

// Trigger defines a pipeline trigger and the cron schedule when the checks
// for new versions should be done
type Trigger struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
}

// PipelineStatus defines the observed state of Pipeline
type PipelineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	BuildNo int `json:"buildNo"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
