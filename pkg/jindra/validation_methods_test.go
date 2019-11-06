package jindra

import (
	"errors"
	"reflect"
	"testing"

	jindra "github.com/kesselborn/jindra/pkg/apis/jindra/v1alpha1"
)

func TestTriggerInResourceExists(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Resources.Triggers = append(ppl.Spec.Resources.Triggers, jindra.JindraPipelineResourcesTrigger{
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
	ppl.Spec.Resources.Triggers = append(ppl.Spec.Resources.Triggers, jindra.JindraPipelineResourcesTrigger{
		Name:     "slack",
		Schedule: "* * * * *",
	})

	err := emptyErrorWrapper(ppl.Validate())
	expected := errors.New("invalid trigger 'slack': every trigger needs to be an input resource of the first stage")

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "trigger needs to be an in-resource of stage 1", errMsg(t, expected.Error(), err.Error()))
	}
}
