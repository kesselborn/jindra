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
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// stagePods keeps pod configurations for the pipeline run
type stagePods map[string]core.Pod

var nodeAffinity = core.NodeAffinity{
	PreferredDuringSchedulingIgnoredDuringExecution: []core.PreferredSchedulingTerm{
		{
			Preference: core.NodeSelectorTerm{
				MatchExpressions: []core.NodeSelectorRequirement{
					{
						Key:      "kubernetes.io/hostname",
						Operator: core.NodeSelectorOpIn,
						Values:   []string{"${MY_NODE_NAME}"},
					},
				},
			},
			Weight: 1,
		},
	},
}

// NewPipelineFromYaml creates a pipeline object from yaml source code
func NewPipelineFromYaml(yamlData []byte) (Pipeline, error) {
	// convert yaml to json as annotations are only for json in Pipeline
	jsonData, err := yaml.YAMLToJSON(yamlData)
	if err != nil {
		return Pipeline{}, fmt.Errorf("cannot convert yaml to json data: %s", err)
	}

	var p Pipeline
	err = json.Unmarshal(jsonData, &p)
	if err != nil {
		return Pipeline{}, fmt.Errorf("cannot unmarshal json data %s: %s", string(jsonData), err)
	}

	return p, nil
}

// RunnerPod creates the job that runs the pipeline
func (ppl Pipeline) RunnerPod(buildNo int) (core.Pod, error) {
	return core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: defaultLabels(ppl.Name, buildNo, ""),
			Name:   fmt.Sprintf(nameFormatString, ppl.Name, buildNo),
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},

		Spec: core.PodSpec{
			RestartPolicy:      core.RestartPolicyNever,
			ServiceAccountName: runnerServiceAccount,
			Volumes: append(jindraVolumes([]string{"transit"}),
				core.Volume{
					Name: "stages", VolumeSource: core.VolumeSource{
						ConfigMap: &core.ConfigMapVolumeSource{
							LocalObjectReference: core.LocalObjectReference{
								Name: fmt.Sprintf(configMapFormatString, ppl.Name, buildNo),
							},
						},
					},
				},
				core.Volume{
					Name: "rsync", VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: fmt.Sprintf(rsyncSecretFormatString, ppl.Name, buildNo),
							Items: []core.KeyToPath{
								{Key: rsyncSecretPubKey, Path: "./authorized_keys"},
							},
						},
					},
				},
			),
			Containers: []core.Container{
				ppl.jindraRunnerContainer(buildNo),
				ppl.podWatcherContainer(),
				ppl.rsyncServerContainer(),
			},
			InitContainers: []core.Container{
				ppl.semaphoreContainer(),
			},
		},
	}, nil
}

// PipelineRunConfigMap creates the config map that the pipeline job uses
// to create the stage pods
func (ppl Pipeline) PipelineRunConfigMap(buildNo int) (core.ConfigMap, error) {
	configs, err := ppl.generateStagePods(buildNo)
	if err != nil {
		return core.ConfigMap{}, err
	}

	cmData := map[string]string{}
	for key, pod := range configs {
		cmData[key] = interface2yaml(pod)
	}

	return core.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf(configMapFormatString, ppl.Name, buildNo),
			Labels: defaultLabels(ppl.Name, buildNo, ""),
		},
		Data: cmData,
	}, nil
}

// NewRsyncSSHSecret creates a Kubernetes Secret with a public (key: pub) and
// private ssh key (key: priv)
func (ppl Pipeline) NewRsyncSSHSecret(buildNo int) (core.Secret, error) {
	privateKey, publicKey, err := generateSSHKeyPair()
	if err != nil {
		return core.Secret{}, fmt.Errorf("error creating keypair: %s", err)
	}

	return core.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Data: map[string][]byte{
			rsyncSecretPrivateKey: privateKey,
			rsyncSecretPubKey:     publicKey,
		},
		Type: core.SecretType("Opaque"),
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf(rsyncSecretFormatString, ppl.Name, buildNo),
			Labels: defaultLabels(ppl.Name, buildNo, ""),
		},
	}, nil
}

func (ppl Pipeline) generateStagePods(buildNo int) (stagePods, error) {
	config := stagePods{}
	ppl.Status.BuildNo = buildNo

	pods := ppl.Spec.Stages
	for _, stage := range []*core.Pod{ppl.Spec.OnSuccess, ppl.Spec.OnError, ppl.Spec.Final} {
		if stage != nil {
			pods = append(pods, *stage)
		}
	}

	for i, stage := range pods {
		if stage.Name == "" && stage.Annotations == nil {
			continue
		}
		setDefaults(&stage, buildNo)
		for k, v := range defaultLabels(ppl.Name, buildNo, stage.Name) {
			stage.Labels[k] = v
		}

		stageName := fmt.Sprintf("%02d-%s", i+1, stage.GetName())
		name := fmt.Sprintf("${MY_NAME}.%s", stageName)
		stage.SetName(name)

		stage.Annotations[waitForAnnotationKey] = strings.Join(generateWaitForAnnotation(stage), ",")

		stage.Spec.Containers = ppl.generateStageContainers(stage, stageName, strings.Join(generateWaitForAnnotation(stage), ","))
		var err error
		if stage.Spec.InitContainers, err = ppl.generateStageInitContainers(stage); err != nil {
			return stagePods{}, fmt.Errorf("error constructing init containers: %s", err)
		}

		stage.Spec.Affinity = &core.Affinity{NodeAffinity: &nodeAffinity}

		defaultMode := int32(256)
		stage.Spec.Volumes = append(stage.Spec.Volumes,
			append(jindraVolumes(resourceNames(stage)), core.Volume{
				Name: "jindra-rsync-ssh-keys",
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName:  fmt.Sprintf(rsyncSecretFormatString, ppl.Name, ppl.Status.BuildNo),
						DefaultMode: &defaultMode,
						Items: []core.KeyToPath{
							core.KeyToPath{Key: rsyncSecretPrivateKey, Path: "./jindra"},
						},
					},
				},
			})...,
		)

		config[stageName+".yaml"] = stage
	}

	return config, nil
}

// generate init container array which consists of:
// - init containers defined in the pipeline
// - a container which copies jindra tools into a shared volume
// - in resource containers
func (ppl Pipeline) generateStageInitContainers(p core.Pod) ([]core.Container, error) {
	createLocksSrc := []string{
		"touch " + path.Join(semaphoresPrefixPath, "steps-running"),
		"touch " + path.Join(semaphoresPrefixPath, "outputs-running"),
	}
	for _, name := range containerNames(p) {
		createLocksSrc = append(createLocksSrc, "touch "+path.Join(semaphoresPrefixPath, "container-"+name))
	}
	toolsMount := core.VolumeMount{Name: toolsMountName, MountPath: toolsPrefixPath}

	initContainers := []core.Container{ppl.getJindraToolsContainer(toolsMount, createLocksSrc)}

	toolsMount.ReadOnly = true

	inResourceEnvs := map[string][]core.EnvVar{}
	annotation, ok := p.Annotations[inResourceEnvAnnotationKey]

	if ok {
		inResourceEnvs = annotationToEnv(annotation)
	}

	debugArgs := []string{}
	if value, ok := p.Annotations[debugResourcesAnnotationKey]; ok && value == "enable" {
		debugArgs = append(debugArgs, "-wait-on-fail", "-debug-out=/tmp/jindra.debug")
	}

	for _, inName := range inResourcesNames(p) {
		c, err := ppl.resourceContainer(inName)
		if err != nil {
			// TODO: use logger
			fmt.Fprintf(os.Stderr, "error creating init container: %s", err)
			continue
		}
		if _, ok := inResourceEnvs[inName]; ok {
			if c.Env == nil {
				c.Env = []core.EnvVar{}
			}
			c.Env = append(c.Env, inResourceEnvs[inName]...)
		}
		c.VolumeMounts = append(c.VolumeMounts, []core.VolumeMount{
			{Name: resourceVolumePrefix + inName, MountPath: path.Join(resourcesPrefixPath, inName)},
			toolsMount,
		}...)
		c.Name = inResourceContainerNamePrefix + c.Name
		c.Args =
			append(
				append(
					append([]string{
						path.Join(toolsPrefixPath, "crij"),
						"-env-prefix=" + inName,
						"-semaphore-file=" + path.Join(semaphoresPrefixPath, "setting-up-pod"),
						"-env-file=" + path.Join(resourcesPrefixPath, inName, resourceEnvFile),
						"-ignore-missing-env-file",
						"-delete-env-file-after-read",
						"-stderr-file=" + path.Join(resourcesPrefixPath, inName, inResourceStderrFile),
						"-stdout-file=" + path.Join(resourcesPrefixPath, inName, inResourceStdoutFile),
					}),
					debugArgs...,
				),
				"/opt/resource/in",
				path.Join(resourcesPrefixPath, inName))
		initContainers = append(initContainers, c)
	}

	podInitContainers := map[string]core.Container{}
	for _, container := range p.Spec.InitContainers {
		podInitContainers[container.Name] = container
	}

	// prepend annotated firstInitContainers in correct order (loop down as we always prepend)
	containers := strings.Split(p.Annotations[firstInitContainers], ",")
	for i := len(containers) - 1; i >= 0; i-- {
		if containers[i] == "" {
			continue
		}
		c, ok := podInitContainers[containers[i]]
		if !ok {
			return initContainers, fmt.Errorf("defined firstInitContainer %s not found in pipeline definition", containers[i])
		}
		initContainers = append([]core.Container{c}, initContainers...)
		delete(podInitContainers, containers[i])
	}

	// append other init containers in the order specified
	for _, c := range p.Spec.InitContainers {
		if _, ok := podInitContainers[c.Name]; ok {
			initContainers = append(initContainers, podInitContainers[c.Name])
		}
	}

	return initContainers, nil
}

// generate container array which consists of:
// - containers defined in the pipeline
// - out resource containers
// - jindra watch container which deletes semaphores once other containers are finished
// - debug container if debugging annotation was set
func (ppl Pipeline) generateStageContainers(p core.Pod, stageName string, waitFor string) []core.Container {
	toolsMount := core.VolumeMount{Name: toolsMountName, MountPath: toolsPrefixPath, ReadOnly: true}
	semaphoreMount := core.VolumeMount{Name: sempahoresMountName, MountPath: semaphoresPrefixPath}

	containers := append(p.Spec.Containers, ppl.watcherContainer(stageName, waitFor, semaphoreMount))

	if p.Annotations[debugContainerAnnotationKey] == "enable" {
		containers = append([]core.Container{ppl.debugContainer(toolsMount, semaphoreMount, resourceNames(p))}, containers...)
	}

	outResourceEnvs := map[string][]core.EnvVar{}
	annotation, ok := p.Annotations[outResourceEnvAnnotationKey]

	debugArgs := []string{}
	if value, ok := p.Annotations[debugResourcesAnnotationKey]; ok && value == "enable" {
		debugArgs = append(debugArgs, "-wait-on-fail", "-debug-out=/tmp/jindra.debug")
	}

	if ok {
		outResourceEnvs = annotationToEnv(annotation)
	}

	for _, outName := range outResourcesNames(p) {
		c, err := ppl.resourceContainer(outName)
		if err != nil {
			// TODO: use logger
			fmt.Fprintf(os.Stderr, "error creating init container: %s", err)
			continue
		}
		if _, ok := outResourceEnvs[outName]; ok {
			if c.Env == nil {
				c.Env = []core.EnvVar{}
			}
			c.Env = append(c.Env, outResourceEnvs[outName]...)
		}
		c.VolumeMounts = append(c.VolumeMounts, []core.VolumeMount{
			{Name: resourceVolumePrefix + outName, MountPath: path.Join(resourcesPrefixPath, outName)},
			toolsMount,
			semaphoreMount,
		}...)
		c.Name = outResourceContainerNamePrefix + c.Name
		c.Args =
			append(
				append(
					append([]string{
						path.Join(toolsPrefixPath, "crij"),
						"-env-prefix=" + outName,
						"-semaphore-file=" + path.Join(semaphoresPrefixPath, "steps-running"),
						"-env-file=" + path.Join(resourcesPrefixPath, outName, resourceEnvFile),
						"-ignore-missing-env-file",
						"-delete-env-file-after-read",
						"-stderr-file=" + path.Join(resourcesPrefixPath, outName, outResourceStderrFile),
						"-stdout-file=" + path.Join(resourcesPrefixPath, outName, outResourceStdoutFile),
					}),
					debugArgs...,
				),
				"/opt/resource/out",
				path.Join(resourcesPrefixPath, outName))

		containers = append(containers, c)
	}

	return containers
}

func (ppl Pipeline) resourceContainer(name string) (core.Container, error) {
	for _, c := range ppl.Spec.Resources.Containers {
		if c.Name == name {
			return c, nil
		}
	}

	if name == "transit" {
		return ppl.transitContainer(), nil
	}

	return core.Container{}, fmt.Errorf("there is no resource with name %s", name)
}

func setDefaults(p *core.Pod, buildNo int) {
	if p.Kind == "" {
		p.Kind = "Pod"
	}
	if p.APIVersion == "" {
		p.APIVersion = "v1"
	}
	if p.Labels == nil {
		p.Labels = map[string]string{}
	}
	if p.Labels["jindra.io/uid"] == "" {
		p.Labels["jindra.io/uid"] = "${MY_UID}"
	}

	if p.Annotations == nil {
		p.Annotations = map[string]string{}
	}

	p.Spec.RestartPolicy = core.RestartPolicyNever

	for i, c := range p.Spec.Containers {
		p.Spec.Containers[i].VolumeMounts = append(p.Spec.Containers[i].VolumeMounts, jindraVolumeMounts(c, resourceNames(*p))...)
	}
	for i, c := range p.Spec.InitContainers {
		p.Spec.InitContainers[i].VolumeMounts = append(p.Spec.InitContainers[i].VolumeMounts, jindraVolumeMounts(c, resourceNames(*p))...)
	}
}

func jindraVolumes(resources []string) []core.Volume {
	volumes := []core.Volume{}
	emptyDirVolumes := []string{toolsMountName, sempahoresMountName}
	for _, name := range resources {
		emptyDirVolumes = append(emptyDirVolumes, resourceVolumePrefix+name)
	}

	for _, name := range emptyDirVolumes {
		volumes = append(volumes, core.Volume{
			Name: name,
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		})
	}

	return volumes
}

func jindraVolumeMounts(c core.Container, resources []string) []core.VolumeMount {
	mounts := []core.VolumeMount{}

	for _, r := range resources {
		mounts = append(mounts, core.VolumeMount{
			Name:      resourceVolumePrefix + r,
			MountPath: path.Join(resourcesPrefixPath, r),
		})
	}

	return mounts
}
