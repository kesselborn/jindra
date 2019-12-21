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
	"fmt"
	"path"

	core "k8s.io/api/core/v1"
)

func jindraRunnerContainer(ppl Pipeline, buildNo int) core.Container {
	return core.Container{
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
	}
}
