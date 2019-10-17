package jindra

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
	jindra "github.com/kesselborn/jindra/pkg/apis/jindra/v1alpha1"
	core "k8s.io/api/core/v1"
)

func getExamplePipeline(t *testing.T) jindra.JindraPipeline {
	examplePipeline := "../../playground/pipeline-example.yaml"
	yamlData, err := ioutil.ReadFile(examplePipeline)
	if err != nil {
		t.Fatalf("error reading example pipeline file: %s: %s", examplePipeline, err)
	}

	p, err := NewJindraPipeline(yamlData)
	if err != nil {
		t.Fatalf("cannot convert yaml to jindra pipeline: %s", err)
	}

	return p
}

func ok() string {
	return " [OK]"
}

func errMsg(t *testing.T, expected interface{}, got interface{}) string {
	gotString, err := yaml.Marshal(got)
	if err != nil {
		t.Fatalf("error marshalling %#v: %s", gotString, err)
	}
	err = ioutil.WriteFile("/tmp/got", gotString, 0644)
	if err != nil {
		t.Fatalf("error writing diff file /tmp/got: %s", err)
	}

	expectedString, err := yaml.Marshal(expected)
	if err != nil {
		t.Fatalf("error marshalling %#v: %s", expected, err)
	}
	err = ioutil.WriteFile("/tmp/expected", expectedString, 0644)
	if err != nil {
		t.Fatalf("error writing diff file /tmp/expected: %s", err)
	}

	return fmt.Sprintf(" [FAILED]\n\t\texpected %T: %s\n\t\tgot      %T: %s\ncheck with diff /tmp/got /tmp/expected", expected, string(expectedString), got, string(gotString))
}

func TestBasicUnmarshalingTest(t *testing.T) {
	p := getExamplePipeline(t)

	for i, test := range []struct {
		got         interface{}
		expectation interface{}
		desc        string
	}{
		{p.Kind, "JindraPipeline", "correct kind"},
		{p.APIVersion, "jindra.io/v1alpha1", "correct api version"},
		{p.ObjectMeta.Name, "http-fs", "correct name"},
		{len(p.Spec.Resources.Triggers), 1, "number of triggers"},
		{p.Spec.Resources.Triggers[0].Name, "git", "trigger name correct"},
		{len(p.Spec.Resources.Containers), 3, "number of resources"},
		{p.Spec.Resources.Containers[2].Name, "slack", "resource name"},
		{len(p.Spec.Stages), 2, "number of stages"},
		{p.Spec.Stages[1].ObjectMeta.Name, "build-docker-image", "stage name"},
		{p.Spec.OnSuccess.ObjectMeta.Annotations["jindra.io/outputs"], "slack", "on success outputs"},
		{p.Spec.OnError.ObjectMeta.Annotations["jindra.io/outputs"], "slack", "on success outputs"},
	} {
		if reflect.DeepEqual(test.expectation, test.got) {
			t.Logf("\t%2d: %-80s %s", i, test.desc, ok())
		} else {
			t.Fatalf("\t%2d: %-80s %s", i, test.desc, errMsg(t, test.expectation, test.got))
		}
	}

}

func podFileContents(file string) *core.Pod {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	var data core.Pod
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		panic(err)
	}

	return &data
}

func configMapFileContents(file string) *core.ConfigMap {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	jsonContent, err := yaml.YAMLToJSON(content)
	if err != nil {
		panic(err)
	}

	var data core.ConfigMap
	err = json.Unmarshal(jsonContent, &data)
	if err != nil {
		panic(err)
	}

	return &data
}

func TestStageConfigs(t *testing.T) {
	configs, configsErr := pipelineConfigs(getExamplePipeline(t), 42)
	configMap, configMapErr := PipelineRunConfigMap(getExamplePipeline(t), 42)

	for i, test := range []struct {
		got         interface{}
		expectation interface{}
		desc        string
	}{
		{configsErr, nil, "configs creation should not error out"},
		{len(configs), 4, "pipeline should have four configs"},
		{configs["01-build-go-binary.yaml"], *podFileContents("../../playground/jindra.http-fs.42.01-build-go-binary.yaml"), "stage 01 should be correct"},
		{configs["02-build-docker-image.yaml"], *podFileContents("../../playground/jindra.http-fs.42.02-build-docker-image.yaml"), "stage 02 should be correct"},
		{configs["03-on-success.yaml"], *podFileContents("../../playground/jindra.http-fs.42.03-on-success.yaml"), "on success should be correct"},
		{configs["04-on-error.yaml"], *podFileContents("../../playground/jindra.http-fs.42.04-on-error.yaml"), "on error should be correct"},
		{configMapErr, nil, "configmap creation should not error out"},
		{configMap, *configMapFileContents("../../playground/jindra.http-fs.42.stages.yaml"), "configmap should be correct"},
	} {
		if reflect.DeepEqual(test.expectation, test.got) {
			t.Logf("\t%2d: %-80s %s", i, test.desc, ok())
		} else {
			t.Fatalf("\t%2d: %-80s %s", i, test.desc, errMsg(t, test.expectation, test.got))
		}
	}

}

func TestAnnotationEnvConverter(t *testing.T) {
	annotation := `
          slack.params.text=Job succeeded
          slack.params.icon_emoji=":ghost:"
          slack.params.attachments='[{"color":"#00ff00","text":"hihihi"}]'
		  rsync.params.foo="bar"
		  rsync.source.url="rsync://foo.bar"
	`

	expectedSlackEnv := []core.EnvVar{
		{Name: "slack.params.text", Value: "Job succeeded"},
		{Name: "slack.params.icon_emoji", Value: "\":ghost:\""},
		{Name: "slack.params.attachments", Value: "'[{\"color\":\"#00ff00\",\"text\":\"hihihi\"}]'"},
	}

	e := annotationToEnv(annotation)

	if !reflect.DeepEqual(expectedSlackEnv, e["slack"]) {
		t.Fatalf("\t%2d: %-80s %s", 0, "converting env annotation failed", errMsg(t, expectedSlackEnv, e))
	}
}

// TODO: test duplicate resource name
// TODO: more complex case of wait-for
// TODO: test naming of files if no onSuccess / onError exists
