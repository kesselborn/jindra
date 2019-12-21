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

// annotation keys
const (
	buildNoOffsetAnnotationKey   = "jindra.io/build-no-offset"
	debugContainerAnnotationKey  = "jindra.io/debug-container"
	debugResourcesAnnotationKey  = "jindra.io/debug-resources"
	firstInitContainers          = "jindra.io/first-init-containers"
	inResourceAnnotationKey      = "jindra.io/inputs"
	inResourceEnvAnnotationKey   = "jindra.io/inputs-envs"
	outResourceAnnotationKey     = "jindra.io/outputs"
	outResourceEnvAnnotationKey  = "jindra.io/outputs-envs"
	servicesAnnotationKey        = "jindra.io/services"
	waitForAnnotationKey         = "jindra.io/wait-for"
	imagePullPolicyAnnotationKey = "jindra.io/image-pull-policy"
)

// container image names
const (
	runnerContainerName = "runner"
	runnerImage         = "jindra/jindra-runner:latest"

	podwatcherContainerName = "pod-watcher"
	podwatcherImage         = "jindra/pod-watcher:latest"

	rsyncContainerName = "rsync"
	rsyncImage         = "jindra/rsync-server:latest"

	setSempahoresContainerName = "set-semaphores"
	setSemaphoresImage         = "alpine"

	toolsContainerName = "get-jindra-tools"
	toolsImage         = "jindra/tools"

	transitContainerName = "transit"
	transitImage         = "mrsixw/concourse-rsync-resource"

	watcherContainerName = "jindra-watcher"
	watcherImage         = "alpine"
)

const (
	debugContainerName = "jindra-debug-container"

	pipelineLabelKey = "jindra.io/pipeline"
	runLabelKey      = "jindra.io/run"

	resourcesPrefixPath   = "/jindra/resources"
	semaphoresPrefixPath  = "/var/lock/jindra"
	toolsPrefixPath       = "/opt/jindra/bin"
	resourceEnvFile       = ".jindra.resource.env"
	inResourceStdoutFile  = ".jindra.in-resource.stdout"
	inResourceStderrFile  = ".jindra.in-resource.stderr"
	outResourceStdoutFile = ".jindra.out-resource.stdout"
	outResourceStderrFile = ".jindra.out-resource.stderr"

	inResourceContainerNamePrefix  = "jindra-resource-in-"
	outResourceContainerNamePrefix = "jindra-resource-out-"
	resourceVolumePrefix           = "jindra-resource-"

	nameFormatString        = "jindra.%s.%d"
	rsyncSecretFormatString = nameFormatString + ".rsync-keys"
	configMapFormatString   = nameFormatString + ".stages"

	sempahoresMountName = "jindra-semaphores"
	toolsMountName      = "jindra-tools"

	rsyncSecretPubKey     = "pub"
	rsyncSecretPrivateKey = "priv"

	runnerServiceAccount = "jindra-runner"

	jindraStagesMountPath  = "/jindra/stages"
	stagesRunningSemaphore = "stages-running"

	outResourceEnvFileName = ".jindra.env"
)
