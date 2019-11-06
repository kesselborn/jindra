package v1alpha1

import (
	"fmt"
	"strings"
)

func (ppl JindraPipeline) validateTriggerHasResource() error {
	containerNames := map[string]bool{}
	for _, resource := range ppl.Spec.Resources.Containers {
		containerNames[resource.Name] = true
	}

	for _, trigger := range ppl.Spec.Resources.Triggers {
		if _, ok := containerNames[trigger.Name]; !ok {
			return fmt.Errorf("there is no resource for trigger '%s'", trigger.Name)
		}
	}

	return nil
}

func (ppl JindraPipeline) validateTriggerIsInResourceOfFirstStage() error {
	stage1InResourcesArray := strings.Split(ppl.Spec.Stages[0].Annotations[InResourceAnnotationKey], ",")
	stage1InResources := map[string]bool{}

	for _, res := range stage1InResourcesArray {
		stage1InResources[res] = true
	}

	for _, trigger := range ppl.Spec.Resources.Triggers {
		if _, ok := stage1InResources[trigger.Name]; !ok {
			return fmt.Errorf("invalid trigger '%s': every trigger needs to be an input resource of the first stage", trigger.Name)
		}
	}

	return nil
}
