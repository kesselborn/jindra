package v1alpha1

const (
	InResourceAnnotationKey     = "jindra.io/inputs"
	OutResourceAnnotationKey    = "jindra.io/outputs"
	ServicesAnnotationKey       = "jindra.io/services"
	WaitForAnnotationKey        = "jindra.io/wait-for"
	DebugContainerAnnotationKey = "jindra.io/debug-container"
	InResourceEnvAnnotationKey  = "jindra.io/inputs-envs"
	OutResourceEnvAnnotationKey = "jindra.io/outputs-envs"
	FirstInitContainers         = "jindra.io/first-init-containers"
	DebugResourcesAnnotationKey = "jindra.io/debug-resources"
	BuildNoOffsetAnnotationKey  = "jindra.io/build-no-offset"
)
