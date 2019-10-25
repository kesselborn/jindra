package jindra

const (
	inResourceAnnotationKey     = "jindra.io/inputs"
	outResourceAnnotationKey    = "jindra.io/outputs"
	servicesAnnotationKey       = "jindra.io/services"
	waitForAnnotationKey        = "jindra.io/wait-for"
	debugContainerAnnotationKey = "jindra.io/debug-container"
	inResourceEnvAnnotationKey  = "jindra.io/inputs-envs"
	outResourceEnvAnnotationKey = "jindra.io/outputs-envs"

	debugContainerName = "jindra-debug-container"

	runLabelKey      = "jindra.io/run"
	pipelineLabelKey = "jindra.io/pipeline"

	resourcesPrefixPath  = "/jindra/resources"
	semaphoresPrefixPath = "/var/lock/jindra"
	toolsPrefixPath      = "/opt/jindra/bin"

	inResourceContainerNamePrefix  = "jindra-resource-in-"
	outResourceContainerNamePrefix = "jindra-resource-out-"
	resourceVolumePrefix           = "jindra-resource-"

	nameFormatString        = "jindra.%s.%d"
	rsyncSecretFormatString = nameFormatString + ".rsync-keys"
	configMapFormatString   = nameFormatString + ".stages"

	sempahoresMountName = "jindra-semaphores"
	toolsMountName      = "jindra-tools"

	// job containers
	runnerImage                = "jindra/jindra-runner:latest"
	runnerContainerName        = "runner"
	podwatcherImage            = "jindra/pod-watcher:latest"
	podwatcherContainerName    = "pod-watcher"
	rsyncImage                 = "jindra/rsync-server:latest"
	rsyncContainerName         = "rsync"
	setSempahoresContainerName = "set-semaphores"
	setSemaphoresImage         = "alpine"

	rsyncSecretPubKey     = "pub"
	rsyncSecretPrivateKey = "priv"

	runnerServiceAccount = "jindra-runner"

	jindraStagesMountPath  = "/jindra/stages"
	stagesRunningSemaphore = "stages-running"
)
