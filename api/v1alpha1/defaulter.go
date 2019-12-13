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
)

// SetDefaults sets sane default values if not set already
func (ppl *Pipeline) SetDefaults() []interface{} {
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

	if ppl.Annotations[buildNoOffsetAnnotationKey] == "" {
		modifiedItems = append(modifiedItems, "/build-offset", "0")
		ppl.Annotations[buildNoOffsetAnnotationKey] = "0"
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
