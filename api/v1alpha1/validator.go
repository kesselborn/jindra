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
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// Validate validates the correctnes of the Pipeline object
func (ppl Pipeline) Validate() error {
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

var valLog = logf.Log.WithName("jindra-validator")

func (ppl Pipeline) triggerHasResource() error {
	valLog.Info("validation: triggerHasResource")
	containerNames := map[string]bool{}
	for _, resource := range ppl.Spec.Resources.Containers {
		containerNames[resource.Name] = true
	}

	for _, trigger := range ppl.Spec.Resources.Triggers {
		if _, ok := containerNames[trigger.Name]; !ok && trigger.Name != "" {
			return fmt.Errorf("there is no resource for trigger '%s'", trigger.Name)
		}
	}

	return nil
}

func (ppl Pipeline) triggerIsInResourceOfFirstStage() error {
	valLog.Info("validation: triggerIsInResourceOfFirstStage")
	stage1InResourcesArray := strings.Split(ppl.Spec.Stages[0].Annotations[inResourceAnnotationKey], ",")
	stage1InResources := arrayToSet(stage1InResourcesArray)

	for _, trigger := range ppl.Spec.Resources.Triggers {
		if _, ok := stage1InResources[trigger.Name]; !ok && trigger.Name != "" {
			return fmt.Errorf("invalid trigger '%s': every trigger needs to be an input resource of the first stage", trigger.Name)
		}
	}

	return nil
}
func (ppl Pipeline) noDuplicateResourceAnnotations() error {
	valLog.Info("validation: noDuplicateResourceAnnotations")
	for _, stage := range ppl.allPods() {
		if duplicate := findDuplicate(strings.Split(stage.Annotations[inResourceAnnotationKey], ",")); duplicate != "" {
			return fmt.Errorf("stage '%s' uses the input resource '%s' twice", stage.Name, duplicate)
		}
		if duplicate := findDuplicate(strings.Split(stage.Annotations[outResourceAnnotationKey], ",")); duplicate != "" {
			return fmt.Errorf("stage '%s' uses the output resource '%s' twice", stage.Name, duplicate)
		}
	}

	return nil
}

func (ppl Pipeline) noDuplicateResourceNames() error {
	valLog.Info("validation: noDuplicateResourceNames")
	containerNames := []string{}
	for _, container := range ppl.Spec.Resources.Containers {
		containerNames = append(containerNames, container.Name)
	}

	if duplicate := findDuplicate(containerNames); duplicate != "" {
		return fmt.Errorf("resource name '%s' is used twice", duplicate)
	}

	return nil
}

func (ppl Pipeline) resourcesExist() error {
	valLog.Info("validation: resourcesExist")
	resourceNames := map[string]bool{"transit": true}

	for _, container := range ppl.Spec.Resources.Containers {
		resourceNames[container.Name] = true
	}

	for _, stage := range ppl.allPods() {
		for _, resource := range strings.Split(stage.Annotations[inResourceAnnotationKey], ",") {
			if _, ok := resourceNames[resource]; !ok && resource != "" {
				return fmt.Errorf("input resource '%s' referenced in stage '%s' does not exist", resource, stage.Name)
			}
		}
		for _, resource := range strings.Split(stage.Annotations[outResourceAnnotationKey], ",") {
			if _, ok := resourceNames[resource]; !ok && resource != "" {
				return fmt.Errorf("output resource '%s' referenced in stage '%s' does not exist", resource, stage.Name)
			}
		}
	}

	return nil
}

func (ppl Pipeline) serviceExist() error {
	valLog.Info("validation: serviceExist")
	for _, stage := range ppl.allPods() {
		if services := strings.Split(stage.Annotations[servicesAnnotationKey], ","); len(services) > 1 && services[0] != "" {
			containers := map[string]bool{}
			for _, container := range stage.Spec.Containers {
				containers[container.Name] = true
			}
			for _, service := range services {
				if _, ok := containers[service]; !ok {
					return fmt.Errorf("service container '%s' referenced in stage '%s' does not exist", service, stage.Name)
				}
			}
		}
	}

	return nil
}

func (ppl Pipeline) noOwnerReference() error {
	valLog.Info("validation: noOwnerReference")
	for _, stage := range ppl.allPods() {
		if len(stage.OwnerReferences) > 0 {
			return fmt.Errorf("stage '%s' must not have an owner reference", stage.Name)
		}
	}

	return nil
}

func (ppl Pipeline) correctOrNoRestartPolicy() error {
	valLog.Info("validation: correctOrNoRestartPolicy")
	for _, stage := range ppl.allPods() {
		if stage.Spec.RestartPolicy != core.RestartPolicyNever && stage.Spec.RestartPolicy != "" {
			return fmt.Errorf(`restartPolicy of stage '%s' must not be set or set to "Never"`, stage.Name)
		}
	}

	return nil
}

func findDuplicate(words []string) string {
	wordSet := map[string]bool{}

	for _, word := range words {
		if wordSet[word] {
			return word
		}
		wordSet[word] = true
	}

	return ""
}

func arrayToSet(words []string) map[string]bool {
	set := map[string]bool{}

	for _, word := range words {
		set[word] = true
	}

	return set
}

func (ppl Pipeline) allPods() []core.Pod {
	pods := ppl.Spec.Stages

	for _, pod := range append([]core.Pod{ppl.Spec.OnSuccess}, ppl.Spec.OnError, ppl.Spec.Final) {
		if pod.Name != "" {
			pods = append(pods, pod)
		}
	}

	return pods
}
