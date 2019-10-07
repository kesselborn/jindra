package v1alpha1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
)

func getExamplePipeline(t *testing.T) JindraPipeline {
	examplePipeline := "../../../../playground/pipeline-example.yaml"
	yamlData, err := ioutil.ReadFile(examplePipeline)
	if err != nil {
		t.Fatalf("error reading example pipeline file: %s: %s", examplePipeline, err)
	}

	// convert yaml to json as annotations are only for json in JindraPipeline
	jsonData, err := yaml.YAMLToJSON(yamlData)
	if err != nil {
		t.Fatalf("cannot convert yaml to json data: %s", err)
	}

	var p JindraPipeline
	err = json.Unmarshal(jsonData, &p)
	if err != nil {
		t.Fatalf("cannot unmarshal json data %s: %s", string(jsonData), err)
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

func interface2yaml(x interface{}) string {
	b, err := yaml.Marshal(x)
	if err != nil {
		return ""
	}

	return string(b)
}

func fileContents(file string) interface{} {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	var data interface{}
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		panic(err)
	}

	return data
}

func TestStage01Config(t *testing.T) {
	p := getExamplePipeline(t)

	config, err := pipelineConfigs(p, 42)

	t.Logf("%#v", fileContents("../../../../playground/jindra.http-fs-42.01-build-binary.yaml"))

	for i, test := range []struct {
		got         interface{}
		expectation interface{}
		desc        string
	}{
		{err, nil, "should not error out"},
		{len(config), 2, "pipeline should have two stage pods"},
		{config[0]["jindra.http-fs-42.01-build-go-binary.yaml"], fileContents("../../../../playground/jindra.http-fs-42.01-build-binary.yaml"), "stage 01 should be correct"},
	} {
		if reflect.DeepEqual(test.expectation, test.got) {
			t.Logf("\t%2d: %-80s %s", i, test.desc, ok())
		} else {
			t.Fatalf("\t%2d: %-80s %s", i, test.desc, errMsg(t, test.expectation, test.got))
		}
	}

}
