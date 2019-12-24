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
	"errors"
	"fmt"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTriggerInResourceExists(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Annotations[imagePullPolicyAnnotationKey] = "xxx"

	err := emptyErrorWrapper(ppl.Validate())
	expected := errors.New("invalid pull policy 'xxx'")

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "pull policy must be empty, Always, IfNotPresent oder Never", errMsg(t, expected.Error(), err.Error()))
	}

}

func TestInvalidPullPolicy(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Resources.Triggers = append(ppl.Spec.Resources.Triggers, Trigger{
		Name:     "xxx",
		Schedule: "* * * * *",
	})

	err := emptyErrorWrapper(ppl.Validate())
	expected := errors.New("there is no resource for trigger 'xxx'")

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "trigger needs a resource", errMsg(t, expected.Error(), err.Error()))
	}

}

func TestTriggerIsInResourceOfFirstStage(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Resources.Triggers = append(ppl.Spec.Resources.Triggers, Trigger{
		Name:     "slack",
		Schedule: "* * * * *",
	})

	err := emptyErrorWrapper(ppl.Validate())
	expected := errors.New("invalid trigger 'slack': every trigger needs to be an input resource of the first stage")

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "trigger needs to be an in-resource of stage 1", errMsg(t, expected.Error(), err.Error()))
	}
}

func TestNoDuplicateInputs(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Stages[0].Annotations[inResourceAnnotationKey] = "slack,git,slack"

	err := emptyErrorWrapper(ppl.Validate())
	expected := errors.New("stage 'build-go-binary' uses the input resource 'slack' twice")

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "input resources must not have duplicates", errMsg(t, expected.Error(), err.Error()))
	}
}

func TestNoDuplicateOutputs(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Stages[0].Annotations[outResourceAnnotationKey] = "slack,git,slack"

	err := emptyErrorWrapper(ppl.Validate())
	expected := errors.New("stage 'build-go-binary' uses the output resource 'slack' twice")

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "input resources must not have duplicates", errMsg(t, expected.Error(), err.Error()))
	}
}

func TestNoDuplicateResourceName(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Resources.Containers[2].Name = ppl.Spec.Resources.Containers[0].Name

	err := emptyErrorWrapper(ppl.Validate())
	expected := fmt.Errorf("resource name '%s' is used twice", ppl.Spec.Resources.Containers[0].Name)

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "resoure names must not be used twice", errMsg(t, expected.Error(), err.Error()))
	}
}

func TestInResourcesExist(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Stages[0].Annotations[inResourceAnnotationKey] = "git,xxx"

	err := emptyErrorWrapper(ppl.Validate())
	expected := fmt.Errorf("input resource '%s' referenced in stage '%s' does not exist", "xxx", ppl.Spec.Stages[0].Name)

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "in resource exists", errMsg(t, expected.Error(), err.Error()))
	}
}

func TestOutResourcesExist(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Stages[1].Annotations[outResourceAnnotationKey] = "git,xxx"

	err := emptyErrorWrapper(ppl.Validate())
	expected := fmt.Errorf("output resource '%s' referenced in stage '%s' does not exist", "xxx", ppl.Spec.Stages[1].Name)

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "out resource exists", errMsg(t, expected.Error(), err.Error()))
	}
}

func TestOnSuccessOutResourceExists(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.OnSuccess.Annotations[outResourceAnnotationKey] = "git,xxx"

	err := emptyErrorWrapper(ppl.Validate())
	expected := fmt.Errorf("output resource '%s' referenced in stage '%s' does not exist", "xxx", ppl.Spec.OnSuccess.Name)

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "out resource exists", errMsg(t, expected.Error(), err.Error()))
	}
}

func TestServiceExists(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Stages[0].Spec.Containers = append(ppl.Spec.Stages[0].Spec.Containers, ppl.Spec.Stages[1].Spec.Containers...)
	ppl.Spec.Stages[0].Annotations[servicesAnnotationKey] = "build-go-binary,xxx"

	err := emptyErrorWrapper(ppl.Validate())
	expected := fmt.Errorf("service container '%s' referenced in stage '%s' does not exist", "xxx", ppl.Spec.Stages[0].Name)

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "service exists", errMsg(t, expected.Error(), err.Error()))
	}
}

func TestNoOwnerReferences(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Stages[0].OwnerReferences = []metav1.OwnerReference{
		{Name: "foo"},
	}

	err := emptyErrorWrapper(ppl.Validate())
	expected := fmt.Errorf("stage '%s' must not have an owner reference", ppl.Spec.Stages[0].Name)

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "no owner reference", errMsg(t, expected.Error(), err.Error()))
	}
}

func TestRestartPolicyIsNeverOrNotSet(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Stages[0].Spec.RestartPolicy = "Always"

	err := emptyErrorWrapper(ppl.Validate())
	expected := fmt.Errorf(`restartPolicy of stage '%s' must not be set or set to "Never"`, ppl.Spec.Stages[0].Name)

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "restartPolicy must be never or empty", errMsg(t, expected.Error(), err.Error()))
	}
}
