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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"

	"golang.org/x/crypto/ssh"
	core "k8s.io/api/core/v1"
)

func containerNames(p core.Pod) []string {
	names := []string{}
	for _, c := range p.Spec.Containers {
		names = append(names, c.Name)
	}

	return names
}

func initContainerNames(p core.Pod) []string {
	names := []string{}
	for _, c := range p.Spec.InitContainers {
		names = append(names, c.Name)
	}

	return names
}

func resourceNames(p core.Pod) []string {
	nameMarker := map[string]bool{}
	names := []string{}

	for _, name := range append(inResourcesNames(p), outResourcesNames(p)...) {
		if _, ok := nameMarker[name]; !ok {
			names = append(names, name)
			nameMarker[name] = true
		}
	}

	return names
}

func inResourcesNames(p core.Pod) []string {
	resourceNames := []string{}
	if inResources := p.Annotations[inResourceAnnotationKey]; inResources != "" {
		resourceNames = append(resourceNames, strings.Split(inResources, ",")...)
	}

	return resourceNames
}

func outResourcesNames(p core.Pod) []string {
	resourceNames := []string{}
	if inResources := p.Annotations[outResourceAnnotationKey]; inResources != "" {
		resourceNames = append(resourceNames, strings.Split(inResources, ",")...)
	}

	return resourceNames
}

func generateSSHKeyPair() (priv []byte, pub []byte, errdx error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("error generating private key: %s", err)
	}

	priv = pem.EncodeToMemory(&pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicRsaKey, err := ssh.NewPublicKey(privateKey.Public())
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("error generating public key: %s", err)
	}

	return priv, ssh.MarshalAuthorizedKey(publicRsaKey), nil
}

func defaultLabels(name string, buildNo int, stageName string) map[string]string {
	labels := map[string]string{
		"jindra.io/pipeline": name,
		"jindra.io/run":      fmt.Sprintf("%d", buildNo),
	}

	if stageName != "" {
		labels["jindra.io/stage"] = stageName
	}

	return labels
}

func generateWaitForAnnotation(p core.Pod) []string {
	waitFor := []string{}
	services := map[string]bool{}

	for _, s := range strings.Split(p.Annotations[servicesAnnotationKey], ",") {
		services[s] = true
	}

	for _, c := range p.Spec.Containers {
		if !services[c.Name] {
			waitFor = append(waitFor, c.Name)
		}
	}

	if p.Annotations[debugContainerAnnotationKey] == "enable" {
		if !services[debugContainerName] {
			waitFor = append(waitFor, debugContainerName)
		}
	}

	return waitFor
}

func interface2yaml(dataStruct interface{}) string {
	// convert to json-string first in order to respect the `json:"omitempty"` tags in yaml
	jsonTxt, err := json.Marshal(dataStruct)
	if err != nil {
		// TODO: use logger
		fmt.Fprintf(os.Stderr, "error marshalling struct: %s\n", err)
		return ""
	}

	var slimDataStruct interface{} // does not containe empty properties which are marked as 'omitempty'

	err = json.Unmarshal(jsonTxt, &slimDataStruct)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error unmarshalling json text: %s\n", err)
		return ""
	}

	yamlTxt, err := yaml.Marshal(slimDataStruct)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling slim data struct: %s\n", err)
		return ""
	}

	return string(yamlTxt)
}

func annotationToEnv(annotation string) map[string][]core.EnvVar {
	e := map[string][]core.EnvVar{}

	for _, envvar := range strings.Split(annotation, "\n") {
		key := strings.Split(strings.TrimLeft(envvar, " 	"), ".")[0]
		tokens := strings.SplitN(strings.TrimLeft(envvar, " 	"), "=", 2)
		if len(tokens) != 2 {
			continue
		}

		if _, ok := e[key]; ok {
			e[key] = append(e[key], core.EnvVar{Name: tokens[0], Value: tokens[1]})
		} else {
			e[key] = []core.EnvVar{{Name: tokens[0], Value: tokens[1]}}
		}
	}

	return e
}
