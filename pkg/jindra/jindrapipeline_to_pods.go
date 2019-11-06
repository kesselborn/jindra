package jindra

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	jindra "github.com/kesselborn/jindra/pkg/apis/jindra/v1alpha1"
	batch "k8s.io/api/batch/v1"
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
func NewPipelineFromYaml(yamlData []byte) (jindra.JindraPipeline, error) {
	// convert yaml to json as annotations are only for json in JindraPipeline
	jsonData, err := yaml.YAMLToJSON(yamlData)
	if err != nil {
		return jindra.JindraPipeline{}, fmt.Errorf("cannot convert yaml to json data: %s", err)
	}

	var p jindra.JindraPipeline
	err = json.Unmarshal(jsonData, &p)
	if err != nil {
		return jindra.JindraPipeline{}, fmt.Errorf("cannot unmarshal json data %s: %s", string(jsonData), err)
	}

	return p, nil
}

// PipelineRunJob creates the job that runs the pipeline
func PipelineRunJob(ppl jindra.JindraPipeline, buildNo int) (batch.Job, error) {
	backoffLimit := int32(0)

	return batch.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: defaultLabels(ppl.Name, buildNo, ""),
			Name:   fmt.Sprintf(nameFormatString, ppl.Name, buildNo),
		},
		Spec: batch.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: defaultLabels(ppl.Name, buildNo, ""),
					Name:   fmt.Sprintf(nameFormatString, ppl.Name, buildNo),
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
						{
							Name:            runnerContainerName,
							Image:           runnerImage,
							ImagePullPolicy: core.PullAlways,
							Env: []core.EnvVar{
								{Name: "MY_IP", ValueFrom: &core.EnvVarSource{FieldRef: &core.ObjectFieldSelector{FieldPath: "status.podIP"}}},
								{Name: "MY_NAME", ValueFrom: &core.EnvVarSource{FieldRef: &core.ObjectFieldSelector{FieldPath: "metadata.name"}}},
								{Name: "MY_NAMESPACE", ValueFrom: &core.EnvVarSource{FieldRef: &core.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
								{Name: "MY_NODE_NAME", ValueFrom: &core.EnvVarSource{FieldRef: &core.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
								{Name: "MY_UID", ValueFrom: &core.EnvVarSource{FieldRef: &core.ObjectFieldSelector{FieldPath: "metadata.uid"}}},

								{Name: "CONFIG_MAP_NAME_FORMAT_STRING", Value: configMapFormatString},
								{Name: "JINDRA_PIPELINE_NAME", Value: ppl.Name},
								{Name: "JINDRA_PIPELINE_RUN_NO", Value: fmt.Sprintf("%d", buildNo)},
								{Name: "JINDRA_SEMAPHORE_MOUNT_PATH", Value: semaphoresPrefixPath},
								{Name: "JINDRA_STAGES_MOUNT_PATH", Value: "/jindra/stages"},
								{Name: "OUT_RESOURCE_ANNOTATION_KEY", Value: outResourceAnnotationKey},
								{Name: "OUT_RESOURCE_CONTAINER_NAME_PREFIX", Value: outResourceContainerNamePrefix},
								{Name: "PIPELINE_LABEL_KEY", Value: pipelineLabelKey},
								{Name: "STAGES_RUNNING_SEMAPHORE", Value: path.Join(semaphoresPrefixPath, stagesRunningSemaphore)},
								{Name: "RSYNC_KEY_NAME_FORMAT_STRING", Value: rsyncSecretFormatString},
								{Name: "RUN_LABEL_KEY", Value: runLabelKey},
								{Name: "WAIT_FOR_ANNOTATION_KEY", Value: waitForAnnotationKey},
							},
							VolumeMounts: append(jindraVolumeMounts(core.Container{}, []string{"transit"}),
								core.VolumeMount{MountPath: semaphoresPrefixPath, Name: sempahoresMountName},
								core.VolumeMount{MountPath: jindraStagesMountPath, Name: "stages"},
							),
						},
						{
							Name:            podwatcherContainerName,
							Image:           podwatcherImage,
							ImagePullPolicy: core.PullAlways,
							Env: []core.EnvVar{
								{Name: "STAGES_RUNNING_SEMAPHORE", Value: path.Join(semaphoresPrefixPath, stagesRunningSemaphore)},
							},
							Args: []string{
								"/bin/sh",
								"-c",
								"/k8s-pod-watcher --debug --semaphore-file ${STAGES_RUNNING_SEMAPHORE}",
							},
							VolumeMounts: []core.VolumeMount{
								core.VolumeMount{MountPath: semaphoresPrefixPath, Name: sempahoresMountName},
							},
						},
						{
							Name:            rsyncContainerName,
							Image:           rsyncImage,
							ImagePullPolicy: core.PullAlways,
							VolumeMounts: []core.VolumeMount{
								core.VolumeMount{MountPath: "/mnt/ssh", Name: "rsync"},
								core.VolumeMount{MountPath: semaphoresPrefixPath, Name: sempahoresMountName},
							},
							Env: []core.EnvVar{
								{Name: "SSH_ENABLE_ROOT", Value: "true"},
								{Name: "STAGES_RUNNING_SEMAPHORE", Value: path.Join(semaphoresPrefixPath, stagesRunningSemaphore)},
							},
						},
					},
					InitContainers: []core.Container{
						{
							Name:            setSempahoresContainerName,
							Image:           setSemaphoresImage,
							ImagePullPolicy: core.PullIfNotPresent,
							Command:         []string{"sh", "-xc", "touch " + path.Join(semaphoresPrefixPath, stagesRunningSemaphore)},
							VolumeMounts: []core.VolumeMount{
								core.VolumeMount{MountPath: semaphoresPrefixPath, Name: sempahoresMountName},
							},
						},
					},
				},
			},
		},
	}, nil
}

// PipelineRunConfigMap creates the config map that the pipeline job uses
// to create the stage pods
func PipelineRunConfigMap(ppl jindra.JindraPipeline, buildNo int) (core.ConfigMap, error) {
	configs, err := generateStagePods(ppl, buildNo)
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
func NewRsyncSSHSecret(ppl jindra.JindraPipeline, buildNo int) (core.Secret, error) {
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

func generateStagePods(ppl jindra.JindraPipeline, buildNo int) (stagePods, error) {
	config := stagePods{}
	ppl.Status.BuildNo = buildNo

	for i, stage := range append(ppl.Spec.Stages, ppl.Spec.OnSuccess, ppl.Spec.OnError, ppl.Spec.Final) {
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

		stage.Spec.Containers = generateStageContainers(stage, stageName, strings.Join(generateWaitForAnnotation(stage), ","), ppl)
		var err error
		if stage.Spec.InitContainers, err = generateStageInitContainers(stage, ppl); err != nil {
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
func generateStageInitContainers(p core.Pod, ppl jindra.JindraPipeline) ([]core.Container, error) {
	createLocksSrc := []string{
		"touch " + path.Join(semaphoresPrefixPath, "steps-running"),
		"touch " + path.Join(semaphoresPrefixPath, "outputs-running"),
	}
	for _, name := range containerNames(p) {
		createLocksSrc = append(createLocksSrc, "touch "+path.Join(semaphoresPrefixPath, "container-"+name))
	}
	toolsMount := core.VolumeMount{Name: toolsMountName, MountPath: toolsPrefixPath}

	initContainers := []core.Container{
		{
			Name:            "get-jindra-tools",
			Image:           "jindra/tools",
			ImagePullPolicy: "Always",
			VolumeMounts: []core.VolumeMount{
				{Name: sempahoresMountName, MountPath: semaphoresPrefixPath},
				toolsMount,
			},
			Command: []string{"sh", "-xc", `cp /jindra/contrib/* ` + toolsPrefixPath + `

# create a few semaphores which can be used to block outputs
# until main steps are finished
` + strings.Join(createLocksSrc, "\n")},
		},
	}

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
		c, err := resourceContainer(ppl, inName)
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
	firstInitContainers := strings.Split(p.Annotations[firstInitContainers], ",")
	for i := len(firstInitContainers) - 1; i >= 0; i-- {
		if firstInitContainers[i] == "" {
			continue
		}
		c, ok := podInitContainers[firstInitContainers[i]]
		if !ok {
			return initContainers, fmt.Errorf("defined firstInitContainer %s not found in pipeline definition", firstInitContainers[i])
		}
		initContainers = append([]core.Container{c}, initContainers...)
		delete(podInitContainers, firstInitContainers[i])
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
func generateStageContainers(p core.Pod, stageName string, waitFor string, ppl jindra.JindraPipeline) []core.Container {
	toolsMount := core.VolumeMount{Name: toolsMountName, MountPath: toolsPrefixPath, ReadOnly: true}
	semaphoreMount := core.VolumeMount{Name: sempahoresMountName, MountPath: semaphoresPrefixPath}

	containers := append(p.Spec.Containers, watcherContainer(stageName, waitFor, semaphoreMount))

	if p.Annotations[debugContainerAnnotationKey] == "enable" {
		containers = append([]core.Container{debugContainer(toolsMount, semaphoreMount, resourceNames(p))}, containers...)
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
		c, err := resourceContainer(ppl, outName)
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

func resourceContainer(ppl jindra.JindraPipeline, name string) (core.Container, error) {
	for _, c := range ppl.Spec.Resources.Containers {
		if c.Name == name {
			return c, nil
		}
	}

	if name == "transit" {
		return core.Container{
			Name:  "transit",
			Image: "mrsixw/concourse-rsync-resource",
			Env: []core.EnvVar{
				{Name: "transit.params.rsync_opts", Value: `["--delete", "--recursive"]`},
				{Name: "transit.source.server", Value: "${MY_IP}"},
				{Name: "transit.source.base_dir", Value: "/tmp"},
				{Name: "transit.source.user", Value: "root"},
				{Name: "transit.source.disable_version_path", Value: "true"},
				{Name: "transit.version", Value: `{"ref":"tmp"}`},
				{Name: "transit.source.private_key", ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						Key:                  rsyncSecretPrivateKey,
						LocalObjectReference: core.LocalObjectReference{Name: fmt.Sprintf(rsyncSecretFormatString, ppl.Name, ppl.Status.BuildNo)},
					},
				}},
			},
		}, nil
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
