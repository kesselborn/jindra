package jindra

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	jindra "github.com/kesselborn/jindra/pkg/apis/jindra/v1alpha1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// pipelineRunConfigs keeps pod configurations for the pipeline run
type pipelineRunConfigs map[string]core.Pod

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

	if p.Annotations == nil {
		p.Annotations = map[string]string{}
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

func jindraContainers(p core.Pod, stageName string, waitFor string, ppl jindra.JindraPipeline) []core.Container {
	toolsMount := core.VolumeMount{Name: "jindra-tools", MountPath: toolsPrefixPath, ReadOnly: true}
	semaphoreMount := core.VolumeMount{Name: "jindra-semaphores", MountPath: semaphoresPrefixPath}

	containers := append(p.Spec.Containers, jindraWatcherContainer(stageName, waitFor, semaphoreMount))

	if p.Annotations[debugContainerAnnotationKey] == "enable" {
		containers = append(containers, jindraDebugContainer(toolsMount, semaphoreMount, getResourceNames(p)))
	}

	semaphoreMount.ReadOnly = true

	outResourceEnvs := map[string][]core.EnvVar{}
	annotation, ok := p.Annotations[outResourceEnvAnnotationKey]

	if ok {
		outResourceEnvs = annotationToEnv(annotation)
	}

	for _, outName := range getOutResourcesNames(p) {
		c, err := getResource(ppl, outName)
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
		c.Args = []string{
			path.Join(toolsPrefixPath, "env-to-json"),
			"-prefix=" + outName,
			"-semaphore-file=" + path.Join(semaphoresPrefixPath, "steps-running"),
			"/opt/resource/out",
			path.Join(resourcesPrefixPath, outName),
		}
		containers = append(containers, c)
	}

	return containers
}

func secretName(name string, buildNo int) string {
	return fmt.Sprintf("jindra.%s.%02d.rsync-keys", name, buildNo)
}

func getResource(ppl jindra.JindraPipeline, name string) (core.Container, error) {
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
				{Name: "transit.source.server", Value: "${MY_IP}"},
				{Name: "transit.source.base_dir", Value: "/tmp"},
				{Name: "transit.source.user", Value: "root"},
				{Name: "transit.source.disable_version_path", Value: "true"},
				{Name: "transit.version", Value: `{"ref":"tmp"}`},
				{Name: "transit.source.private_key", ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						Key:                  "priv",
						LocalObjectReference: core.LocalObjectReference{Name: secretName(ppl.ObjectMeta.Name, ppl.Status.BuildNo)},
					},
				}},
			},
		}, nil
	}

	return core.Container{}, fmt.Errorf("there is no resource with name %s", name)
}

func jindraInitContainers(p core.Pod, ppl jindra.JindraPipeline) []core.Container {
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

	inResourceEnvs := map[string][]core.EnvVar{}
	annotation, ok := p.Annotations[inResourceEnvAnnotationKey]

	if ok {
		inResourceEnvs = annotationToEnv(annotation)
	}

	for _, inName := range getInResourcesNames(p) {
		c, err := getResource(ppl, inName)
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
		c.Command = []string{
			path.Join(toolsPrefixPath, "env-to-json"),
			"-prefix=" + inName,
			"-semaphore-file=" + path.Join(semaphoresPrefixPath, "setting-up-pod"),
			"/opt/resource/in",
			path.Join(resourcesPrefixPath, inName),
		}
		initContainers = append(initContainers, c)
	}

	return initContainers
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

func PipelineRunConfigMap(ppl jindra.JindraPipeline, buildNo int) (core.ConfigMap, error) {
	configs, err := pipelineConfigs(ppl, buildNo)
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
			Name: fmt.Sprintf("jindra.%s.%d.stages", ppl.Name, buildNo),
			Labels: map[string]string{
				"jindra.io/pipeline": ppl.Name,
				"jindra.io/run":      fmt.Sprintf("%d", buildNo),
			},
		},
		Data: cmData,
	}, nil
}

func NewJindraPipeline(yamlData []byte) (jindra.JindraPipeline, error) {
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

func pipelineConfigs(ppl jindra.JindraPipeline, buildNo int) (pipelineRunConfigs, error) {
	config := pipelineRunConfigs{}
	ppl.Status.BuildNo = buildNo

	for i, stage := range append(ppl.Spec.Stages, ppl.Spec.OnSuccess, ppl.Spec.OnError, ppl.Spec.Final) {
		if stage.Name == "" && stage.Annotations == nil {
			continue
		}
		setDefaults(&stage)
		stage.Annotations[waitForAnnotationKey] = strings.Join(getWaitFor(stage), ",")

		stageName := fmt.Sprintf("%02d-%s", i+1, stage.GetName())
		name := fmt.Sprintf("${MY_NAME}.%s", stageName)
		stage.SetName(name)

		stage.Spec.Containers = jindraContainers(stage, stageName, strings.Join(getWaitFor(stage), ","), ppl)
		stage.Spec.InitContainers = jindraInitContainers(stage, ppl)

		stage.Spec.Affinity = &core.Affinity{NodeAffinity: &nodeAffinity}

		defaultMode := int32(256)
		stage.Spec.Volumes = append(volumes(getResourceNames(stage)), core.Volume{
			Name: "jindra-rsync-ssh-keys",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName:  secretName(ppl.ObjectMeta.Name, ppl.Status.BuildNo),
					DefaultMode: &defaultMode,
					Items: []core.KeyToPath{
						core.KeyToPath{Key: "priv", Path: "./jindra"},
					},
				},
			},
		})

		config[stageName+".yaml"] = stage
	}

	return config, nil
}
