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
	"path"

	core "k8s.io/api/core/v1"
)

func (ppl Pipeline) debugContainer(toolsMount, semaphoreMount core.VolumeMount, resourceNames []string) core.Container {
	c := core.Container{
		Name:            debugContainerName,
		Image:           "alpine",
		ImagePullPolicy: ppl.imagePullPolicy(),
		Args: []string{"sh", "-c", `touch /DELETE_ME_TO_STOP_DEBUG_CONTAINER
echo "waiting for /DELETE_ME_TO_STOP_DEBUG_CONTAINER to be deleted "
while test -f /DELETE_ME_TO_STOP_DEBUG_CONTAINER
do
  sleep 1
  printf "."
done`},
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
