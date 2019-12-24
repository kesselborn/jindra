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

func (ppl Pipeline) watcherContainer(stageName, waitFor string, semaphoreMount core.VolumeMount) core.Container {
	return core.Container{
		Name:            watcherContainerName,
		Image:           watcherImage,
		ImagePullPolicy: ppl.imagePullPolicy(),
		Args: []string{"sh", "-c", fmt.Sprintf(`printf "waiting for steps to finish "
containers=$(echo "%s"|sed "s/[,]*%s//g")
while ! wget -qO- ${MY_IP}:8080/pod/${MY_NAME}.%s?containers=${containers}|grep Completed &>/dev/null
do
  printf "."
  sleep 3
done
echo
rm %s
`, waitFor, debugContainerName, stageName, path.Join(semaphoresPrefixPath, "steps-running")),
		},
		Env: []core.EnvVar{
			{Name: "JOB_IP", Value: "${MY_IP}"},
		},
		VolumeMounts: []core.VolumeMount{
			semaphoreMount,
		},
	}
}
