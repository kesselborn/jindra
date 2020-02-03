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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/ghodss/yaml"
	core "k8s.io/api/core/v1"
)

const (
	fixtureDir = "../../tests/fixtures"
)

func getExamplePipeline(t *testing.T) Pipeline {
	examplePipeline := path.Join(fixtureDir, "pipeline-example.yaml")
	yamlData, err := ioutil.ReadFile(examplePipeline)
	if err != nil {
		t.Fatalf("error reading example pipeline file: %s: %s", examplePipeline, err)
	}

	p, err := NewPipelineFromYaml(yamlData)
	if err != nil {
		t.Fatalf("cannot convert yaml to jindra pipeline: %s", err)
	}

	return p
}

func jsonFromYamlFile(file string, t *testing.T) []byte {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("error reading file %s: %s", file, err)
	}

	jsonContent, err := yaml.YAMLToJSON(content)
	if err != nil {
		t.Fatalf("error yaml2json file %s: %s", file, err)
	}

	return jsonContent
}

func podFileContents(file string, t *testing.T) *core.Pod {
	var data core.Pod

	err := json.Unmarshal(jsonFromYamlFile(file, t), &data)
	if err != nil {
		t.Fatalf("error yaml unmarshaling file %s: %s", file, err)
	}

	return &data
}

func configMapFileContents(file string, t *testing.T) *core.ConfigMap {
	var data core.ConfigMap

	err := json.Unmarshal(jsonFromYamlFile(file, t), &data)
	if err != nil {
		t.Fatalf("error json unmarshaling data from %s: %s", file, err)
	}

	return &data
}

func secretFileContents(file string, t *testing.T) *core.Secret {
	var data core.Secret

	err := json.Unmarshal(jsonFromYamlFile(file, t), &data)
	if err != nil {
		t.Fatalf("error json unmarshaling data from %s: %s", file, err)
	}

	return &data
}

func emptyErrorWrapper(e error) error {
	if e == nil {
		return errors.New("<nil>")
	}

	return e
}

func ok() string {
	return " [OK]"
}

func collectJindraContainers(containers []core.Container) []core.Container {
	filteredContainers := []core.Container{}

	jindraContainerNames := map[string]bool{
		podwatcherContainerName:    true,
		rsyncContainerName:         true,
		runnerContainerName:        true,
		setSempahoresContainerName: true,
		toolsContainerName:         true,
		transitContainerName:       true,
		watcherContainerName:       true,
	}

	for _, container := range containers {
		if jindraContainerNames[container.Name] {
			filteredContainers = append(filteredContainers, container)
		}
	}

	return filteredContainers
}

func getJindraContainers(pods stagePods) []core.Container {
	containers := []core.Container{}

	for _, pod := range pods {
		containers = append(containers, collectJindraContainers(append(pod.Spec.InitContainers, pod.Spec.Containers...))...)
	}

	return containers
}

func errMsg(t *testing.T, expected interface{}, got interface{}) string {
	gotString, err := yaml.Marshal(got)
	if err != nil {
		t.Fatalf("error marshalling %#v: %s", gotString, err)
	}

	os.Remove("/tmp/got")
	os.Remove("/tmp/expected")

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
