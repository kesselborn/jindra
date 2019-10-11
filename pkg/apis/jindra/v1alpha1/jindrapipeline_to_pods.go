package v1alpha1

import (
	"fmt"
	"os"
	"path"
	"strings"

	core "k8s.io/api/core/v1"
)

// JindraPipelineRunConfigs keeps pod configurations for the pipeline run
type JindraPipelineRunConfigs []map[string]*core.Pod

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

func getInResourcesNames(p core.Pod) []string {
	resourceNames := []string{}
	if inResources := p.Annotations[inResourceAnnotationKey]; inResources != "" {
		resourceNames = append(resourceNames, strings.Split(inResources, ",")...)
	}

	return resourceNames
}

func getOutResourcesNames(p core.Pod) []string {
	resourceNames := []string{}
	if inResources := p.Annotations[outResourceAnnotationKey]; inResources != "" {
		resourceNames = append(resourceNames, strings.Split(inResources, ",")...)
	}

	return resourceNames
}

func getResourceNames(p core.Pod) []string {
	nameMarker := map[string]bool{}
	names := []string{}

	for _, name := range append(getInResourcesNames(p), getOutResourcesNames(p)...) {
		if _, ok := nameMarker[name]; !ok {
			names = append(names, name)
			nameMarker[name] = true
		}
	}

	return names
}

func volumes(resources []string) []core.Volume {
	volumes := []core.Volume{}
	emptyDirVolumes := []string{"jindra-tools", "jindra-semaphores"}
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

func getWaitFor(p core.Pod) []string {
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

	return waitFor
}

func setDefaults(p *core.Pod) {
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

	p.Spec.RestartPolicy = core.RestartPolicyNever

	for i, c := range p.Spec.Containers {
		p.Spec.Containers[i].VolumeMounts = append(p.Spec.Containers[i].VolumeMounts, getVolumeMounts(c, getResourceNames(*p))...)
	}

}

func getVolumeMounts(c core.Container, resources []string) []core.VolumeMount {
	mounts := []core.VolumeMount{}

	for _, r := range resources {
		mounts = append(mounts, core.VolumeMount{
			Name:      resourceVolumePrefix + r,
			MountPath: path.Join(resourcesPrefixPath, r),
		})
	}

	return mounts
}

func getContainerNames(p core.Pod) []string {
	names := []string{}
	for _, c := range p.Spec.Containers {
		names = append(names, c.Name)
	}

	return names
}

func getInitContainerNames(p core.Pod) []string {
	names := []string{}
	for _, c := range p.Spec.InitContainers {
		names = append(names, c.Name)
	}

	return names
}

func jindraWatcherContainer(stageName, waitFor string, semaphoreMount core.VolumeMount) core.Container {
	return core.Container{
		Name:  "jindra-watcher",
		Image: "alpine",
		Args: []string{"sh", "-c", fmt.Sprintf(`printf "waiting for steps to finish "
while ! wget -qO- ${MY_IP}:8080/pod/${MY_NAME}.%s?containers=%s|grep Completed &>/dev/null
do
  printf "."
  sleep 3
done
echo
rm %s
`, stageName, waitFor, path.Join(semaphoresPrefixPath, "steps-running")),
		},
		Env: []core.EnvVar{
			{Name: "JOB_IP", Value: "${MY_IP}"},
		},
		VolumeMounts: []core.VolumeMount{
			semaphoreMount,
		},
	}
}

func jindraDebugContainer(toolsMount, semaphoreMount core.VolumeMount, resourceNames []string) core.Container {
	c := core.Container{
		Name:  "jindra-debug-container",
		Image: "alpine",
		Args:  []string{"sh", "-c", "sleep 600"},
		Env: []core.EnvVar{
			{Name: "JOB_IP", Value: "${MY_IP}"},
		},
		VolumeMounts: []core.VolumeMount{
			toolsMount,
			semaphoreMount,
		},
	}

	for _, resource := range resourceNames {
		c.VolumeMounts = append(c.VolumeMounts,
			core.VolumeMount{Name: resourceVolumePrefix + resource, MountPath: path.Join(resourcesPrefixPath, resource)},
		)
	}

	return c
}

func jindraContainers(p core.Pod, stageName string, waitFor string, ppl JindraPipeline) []core.Container {
	toolsMount := core.VolumeMount{Name: "jindra-tools", MountPath: toolsPrefixPath, ReadOnly: true}
	semaphoreMount := core.VolumeMount{Name: "jindra-semaphores", MountPath: semaphoresPrefixPath}

	containers := append(p.Spec.Containers, jindraWatcherContainer(stageName, waitFor, semaphoreMount))

	if p.Annotations[debugContainerAnnotationKey] == "enable" {
		containers = append(containers, jindraDebugContainer(toolsMount, semaphoreMount, getResourceNames(p)))
	}

	semaphoreMount.ReadOnly = true

	for _, outName := range getOutResourcesNames(p) {
		c, err := getResource(ppl, outName)
		if err != nil {
			// TODO: use logger
			fmt.Fprintf(os.Stderr, "error creating init container: %s", err)
			continue
		}
		c.VolumeMounts = append(c.VolumeMounts, []core.VolumeMount{
			{Name: resourceVolumePrefix + outName, MountPath: path.Join(resourcesPrefixPath, outName)},
			toolsMount,
			semaphoreMount,
		}...)
		c.Name = outResourceContainerNamePrefix + c.Name
		c.Args = []string{
			path.Join(toolsPrefixPath, "env-to-json"),
			"-prefix",
			outName,
			"-semaphore-file",
			path.Join(semaphoresPrefixPath, "steps-running"),
			"/opt/resource/out",
			path.Join(resourcesPrefixPath, outName),
		}
		containers = append(containers, c)
	}

	return containers
}

func getResource(ppl JindraPipeline, name string) (core.Container, error) {
	for _, c := range ppl.Spec.Resources.Containers {
		if c.Name == name {
			return c, nil
		}
	}

	secretName := fmt.Sprintf("jindra.%s-%02d.rsync-keys", ppl.ObjectMeta.Name, ppl.Status.BuildNo)

	if name == "transit" {
		return core.Container{
			Name:  "transit",
			Image: "mrsixw/concourse-rsync-resource",
			Env: []core.EnvVar{
				{Name: "transit.source.server", Value: "${MY_IP}"},
				{Name: "transit.source.base_dir", Value: "/tmp"},
				{Name: "transit.source.user", Value: "root"},
				{Name: "transit.source.disable_version_path", Value: "true"},
				{Name: "transit.version", Value: `{"ref":"tmp"}`},
				{Name: "transit.source.private_key", ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						Key:                  "priv",
						LocalObjectReference: core.LocalObjectReference{Name: secretName},
					},
				}},
			},
		}, nil
	}

	return core.Container{}, fmt.Errorf("there is no resource with name %s", name)
}

func jindraInitContainers(p core.Pod, ppl JindraPipeline) []core.Container {
	createLocksSrc := []string{
		"touch " + path.Join(semaphoresPrefixPath, "steps-running"),
		"touch " + path.Join(semaphoresPrefixPath, "outputs-running"),
	}
	for _, name := range getContainerNames(p) {
		createLocksSrc = append(createLocksSrc, "touch "+path.Join(semaphoresPrefixPath, "container-"+name))
	}
	for _, name := range getInitContainerNames(p) {
		createLocksSrc = append(createLocksSrc, "touch "+path.Join(semaphoresPrefixPath, "init-container-"+name))
	}

	toolsMount := core.VolumeMount{Name: "jindra-tools", MountPath: toolsPrefixPath}

	initContainers := []core.Container{
		{
			Name:            "get-jindra-tools",
			Image:           "jindra/tools",
			ImagePullPolicy: "Always",
			VolumeMounts: []core.VolumeMount{
				{Name: "jindra-semaphores", MountPath: semaphoresPrefixPath},
				toolsMount,
			},
			Command: []string{"sh", "-xc", `cp /jindra/contrib/* ` + toolsPrefixPath + `

# create a few semaphores which can be used to block outputs
# until main steps are finished
` + strings.Join(createLocksSrc, "\n")},
		},
	}

	toolsMount.ReadOnly = true

	for _, inName := range getInResourcesNames(p) {
		c, err := getResource(ppl, inName)
		if err != nil {
			// TODO: use logger
			fmt.Fprintf(os.Stderr, "error creating init container: %s", err)
			continue
		}
		c.VolumeMounts = append(c.VolumeMounts, []core.VolumeMount{
			{Name: resourceVolumePrefix + inName, MountPath: path.Join(resourcesPrefixPath, inName)},
			toolsMount,
		}...)
		c.Name = inResourceContainerNamePrefix + c.Name
		c.Command = []string{
			path.Join(toolsPrefixPath, "env-to-json"),
			"-prefix",
			inName,
			"-semaphore-file",
			path.Join(semaphoresPrefixPath, "setting-up-pod"),
			"/opt/resource/in",
			path.Join(resourcesPrefixPath, inName),
		}
		initContainers = append(initContainers, c)
	}

	return initContainers
}

func pipelineConfigs(ppl JindraPipeline, buildNo int) (JindraPipelineRunConfigs, error) {
	config := JindraPipelineRunConfigs{}
	ppl.Status.BuildNo = buildNo

	for i, stage := range ppl.Spec.Stages {
		setDefaults(&stage)
		stage.Annotations[waitForAnnotationKey] = strings.Join(getWaitFor(stage), ",")

		stageName := fmt.Sprintf("%02d-%s", i+1, stage.GetName())
		name := fmt.Sprintf("jindra.%s-%02d.%s", ppl.ObjectMeta.Name, buildNo, stageName)
		stage.SetName(name)

		stage.Spec.Containers = jindraContainers(stage, stageName, strings.Join(getWaitFor(stage), ","), ppl)
		stage.Spec.InitContainers = jindraInitContainers(stage, ppl)

		stage.Spec.Affinity = &core.Affinity{NodeAffinity: &nodeAffinity}

		defaultMode := int32(256)
		stage.Spec.Volumes = append(volumes(getResourceNames(stage)), core.Volume{
			Name: "jindra-rsync-ssh-keys",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName:  fmt.Sprintf("jindra.%s-%02d.rsync-keys", ppl.ObjectMeta.Name, buildNo),
					DefaultMode: &defaultMode,
					Items: []core.KeyToPath{
						core.KeyToPath{Key: "priv", Path: "./jindra"},
					},
				},
			},
		})

		config = append(config, map[string]*core.Pod{name + ".yaml": stage.DeepCopy()})
	}

	return config, nil
}
