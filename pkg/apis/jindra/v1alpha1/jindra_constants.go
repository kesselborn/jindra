package v1alpha1

const (
	inResourceAnnotationKey     = "jindra.io/inputs"
	outResourceAnnotationKey    = "jindra.io/outputs"
	servicesAnnotationKey       = "jindra.io/services"
	waitForAnnotationKey        = "jindra.io/wait-for"
	debugContainerAnnotationKey = "jindra.io/debug-container"

	resourcesPrefixPath  = "/jindra/resources"
	semaphoresPrefixPath = "/var/lock/jindra"
	toolsPrefixPath      = "/opt/jindra/bin"

	inResourceContainerNamePrefix  = "jindra-resource-in-"
	outResourceContainerNamePrefix = "jindra-resource-out-"
	resourceVolumePrefix           = "jindra-resource-"
)
