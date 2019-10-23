package jindra

const (
	inResourceAnnotationKey     = "jindra.io/inputs"
	outResourceAnnotationKey    = "jindra.io/outputs"
	servicesAnnotationKey       = "jindra.io/services"
	waitForAnnotationKey        = "jindra.io/wait-for"
	debugContainerAnnotationKey = "jindra.io/debug-container"
	inResourceEnvAnnotationKey  = "jindra.io/inputs-envs"
	outResourceEnvAnnotationKey = "jindra.io/outputs-envs"

	runLabelKey      = "jindra.io/run"
	pipelineLabelKey = "jindra.io/pipeline"

	resourcesPrefixPath  = "/jindra/resources"
	semaphoresPrefixPath = "/var/lock/jindra"
	toolsPrefixPath      = "/opt/jindra/bin"

	inResourceContainerNamePrefix  = "jindra-resource-in-"
	outResourceContainerNamePrefix = "jindra-resource-out-"
	resourceVolumePrefix           = "jindra-resource-"

	nameFormatString        = "jindra.%s.%02d."
	rsyncSecretFormatString = nameFormatString + "rsync-keys"
	configMapFormatString   = nameFormatString + "stages"
)
