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
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// log is for logging in this package.
var defLog = logf.Log.WithName("pipeline-defaulter")

// SetDefaults sets sane default values if not set already
func (ppl *Pipeline) SetDefaults() {
	for _, f := range []func(){
		ppl.setDefaultNames,
		ppl.setDefaultTriggerSchedule,
		ppl.setBuildNoOffset,
		ppl.setRestartPolicies,
	} {
		f()
	}
}

func (ppl *Pipeline) setDefaultNames() {
	if ppl.Spec.Final != nil && ppl.Spec.Final.Name == "" {
		defLog.Info("adding name for final stage", "pipeline", ppl.Name, "stage", "final")
		ppl.Spec.Final.Name = "final"
	}

	if ppl.Spec.OnError != nil && ppl.Spec.OnError.Name == "" {
		defLog.Info("adding name for onError stage", "pipeline", ppl.Name, "stage", "onError")
		ppl.Spec.OnError.Name = "on-error"
	}

	if ppl.Spec.OnSuccess != nil && ppl.Spec.OnSuccess.Name == "" {
		defLog.Info("adding name for onError stage", "pipeline", ppl.Name, "stage", "onError")
		ppl.Spec.OnSuccess.Name = "on-success"
	}
}

func (ppl *Pipeline) setDefaultTriggerSchedule() {
	if ppl.Spec.Resources.Triggers != nil {
		for i := 0; i < len(ppl.Spec.Resources.Triggers); i++ {
			trigger := ppl.Spec.Resources.Triggers[i]
			if trigger.Schedule == "" {
				defLog.Info("setting default trigger schedule", "pipeline", ppl.Name, "trigger", trigger.Name)
				ppl.Spec.Resources.Triggers[i].Schedule = "/5 * * * *"
			}
		}
	}
}

func (ppl *Pipeline) setBuildNoOffset() {
	if ppl.Annotations == nil {
		defLog.Info("adding annotations map", "pipeline", ppl.Name)
		ppl.Annotations = map[string]string{}
	}
	if ppl.Annotations[buildNoOffsetAnnotationKey] == "" {
		defLog.Info("adding build no offset", "pipeline", ppl.Name)
		ppl.Annotations[buildNoOffsetAnnotationKey] = "0"
	}

}

func (ppl *Pipeline) setRestartPolicies() {
	for i := 0; i < len(ppl.Spec.Stages); i++ {
		if ppl.Spec.Stages[i].Spec.RestartPolicy == core.RestartPolicy("") {
			stage := ppl.Spec.Stages[i]
			defLog.Info("setting restart policy", "pipeline", ppl.Name, "stage", stage.Name, "policy", core.RestartPolicyNever)
			ppl.Spec.Stages[i].Spec.RestartPolicy = core.RestartPolicyNever
		}
	}

	for _, pod := range []*core.Pod{ppl.Spec.OnSuccess, ppl.Spec.OnError, ppl.Spec.Final} {
		if (pod != nil && len(pod.Spec.Containers) > 0) && pod.Spec.RestartPolicy == core.RestartPolicy("") {
			defLog.Info("setting restart policy", "pipeline", ppl.Name, "stage", pod.Name, "policy", core.RestartPolicyNever)
			pod.Spec.RestartPolicy = core.RestartPolicyNever
		}
	}
}
