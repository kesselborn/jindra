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
	"strings"

	core "k8s.io/api/core/v1"
)

func (ppl Pipeline) getJindraToolsContainer(toolsMount core.VolumeMount, createLocksSrc []string) core.Container {
	return core.Container{
		Name:            toolsContainerName,
		Image:           toolsImage,
		ImagePullPolicy: ppl.imagePullPolicy(),
		VolumeMounts: []core.VolumeMount{
			{Name: sempahoresMountName, MountPath: semaphoresPrefixPath},
			toolsMount,
		},
		Command: []string{"sh", "-xc", `cp /jindra/contrib/* ` + toolsPrefixPath + `

# create a few semaphores which can be used to block outputs
# until main steps are finished
` + strings.Join(createLocksSrc, "\n")},
	}
}
