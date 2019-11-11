package v1alpha1

import (
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var valLog = logf.Log.WithName("jindra-validator")

func (ppl JindraPipeline) triggerHasResource() error {
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

func (ppl JindraPipeline) triggerIsInResourceOfFirstStage() error {
	valLog.Info("validation: triggerIsInResourceOfFirstStage")
	stage1InResourcesArray := strings.Split(ppl.Spec.Stages[0].Annotations[InResourceAnnotationKey], ",")
	stage1InResources := arrayToSet(stage1InResourcesArray)

	for _, trigger := range ppl.Spec.Resources.Triggers {
		if _, ok := stage1InResources[trigger.Name]; !ok && trigger.Name != "" {
			return fmt.Errorf("invalid trigger '%s': every trigger needs to be an input resource of the first stage", trigger.Name)
		}
	}

	return nil
}
func (ppl JindraPipeline) noDuplicateResourceAnnotations() error {
	valLog.Info("validation: noDuplicateResourceAnnotations")
	for _, stage := range ppl.allPods() {
		if duplicate := findDuplicate(strings.Split(stage.Annotations[InResourceAnnotationKey], ",")); duplicate != "" {
			return fmt.Errorf("stage '%s' uses the input resource '%s' twice", stage.Name, duplicate)
		}
		if duplicate := findDuplicate(strings.Split(stage.Annotations[OutResourceAnnotationKey], ",")); duplicate != "" {
			return fmt.Errorf("stage '%s' uses the output resource '%s' twice", stage.Name, duplicate)
		}
	}

	return nil
}

func (ppl JindraPipeline) noDuplicateResourceNames() error {
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

func (ppl JindraPipeline) resourcesExist() error {
	valLog.Info("validation: resourcesExist")
	resourceNames := map[string]bool{"transit": true}

	for _, container := range ppl.Spec.Resources.Containers {
		resourceNames[container.Name] = true
	}

	for _, stage := range ppl.allPods() {
		for _, resource := range strings.Split(stage.Annotations[InResourceAnnotationKey], ",") {
			if _, ok := resourceNames[resource]; !ok && resource != "" {
				return fmt.Errorf("input resource '%s' referenced in stage '%s' does not exist", resource, stage.Name)
			}
		}
		for _, resource := range strings.Split(stage.Annotations[OutResourceAnnotationKey], ",") {
			if _, ok := resourceNames[resource]; !ok && resource != "" {
				return fmt.Errorf("output resource '%s' referenced in stage '%s' does not exist", resource, stage.Name)
			}
		}
	}

	return nil
}

func (ppl JindraPipeline) serviceExist() error {
	valLog.Info("validation: serviceExist")
	for _, stage := range ppl.allPods() {
		if services := strings.Split(stage.Annotations[ServicesAnnotationKey], ","); len(services) > 1 && services[0] != "" {
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

func (ppl JindraPipeline) noOwnerReference() error {
	valLog.Info("validation: noOwnerReference")
	for _, stage := range ppl.allPods() {
		if len(stage.OwnerReferences) > 0 {
			return fmt.Errorf("stage '%s' must not have an owner reference", stage.Name)
		}
	}

	return nil
}

func (ppl JindraPipeline) correctOrNoRestartPolicy() error {
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

func (ppl JindraPipeline) allPods() []core.Pod {
	pods := ppl.Spec.Stages

	for _, pod := range append([]core.Pod{ppl.Spec.OnSuccess}, ppl.Spec.OnError, ppl.Spec.Final) {
		if pod.Name != "" {
			pods = append(pods, pod)
		}
	}

	return pods
}
