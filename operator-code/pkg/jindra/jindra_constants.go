package jindra

const (
	debugContainerName = "jindra-debug-container"

	runLabelKey      = "jindra.io/run"
	pipelineLabelKey = "jindra.io/pipeline"

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

	outResourceEnvFileName = ".jindra.env"
)