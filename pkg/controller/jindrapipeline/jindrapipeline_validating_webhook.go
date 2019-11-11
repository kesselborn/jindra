package jindrapipeline

// https://github.com/kubernetes-sigs/controller-runtime/blob/f1eaba5087d69cebb154c6a48193e6667f5b512c/example/main.go

// +kubebuilder:webhook:verbs=create,path=/validate-batch-tutorial-kubebuilder-io-v1-cronjob,mutating=false,failurePolicy=fail,groups=jindra.io,resources=jindrapipelines,versions=v1alpha1,name=jindrapipeline.jindra.io

import (
	"context"
	"fmt"
	"net/http"

	jindra "github.com/kesselborn/jindra/pkg/apis/jindra/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/apis/core"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

var valLog = logf.Log.WithName("jindra-validator")

// PipelineValidator validates pipeline objects
type PipelineValidator struct {
	client  client.Client
	decoder types.Decoder
}

var _ admission.Handler = &PipelineValidator{}

// Handle handles incoming requests
// note to self: https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/types.go#L506 <- dry run pods
func (v *PipelineValidator) Handle(ctx context.Context, req types.Request) types.Response {
	valLog.Info("validating pipeline")

	pipeline := &jindra.JindraPipeline{}

	err := v.decoder.Decode(req, pipeline)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	allowed, reason, err := v.validatePipelineFn(ctx, pipeline)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}

	allowed, reason, err = v.ValidateStagesFn(ctx, pipeline)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}

	return admission.ValidationResponse(allowed, reason)
}

var _client *client.Client

func dryRunClient() (*client.Client, error) {
	if _client == nil {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("error getting in-cluster-config: %s", err.Error())
		}

		client, err := client.New(config, client.Options{})
		if err != nil {
			return nil, fmt.Errorf("error creating client: %s", err.Error())
		}
		_client = &client
	}

	return _client, nil
}

func (v *PipelineValidator) validatePod(ctx context.Context, pod core.Pod) error {

	// use generated name instead of name in order to get a unique name
	name := pod.Name
	pod.Name = ""
	pod.GenerateName = name

	v.client.
		v.client.Create(ctx, pod)
}

func (v *PipelineValidator) ValidateStagesFn(ctx context.Context, pipeline *jindra.JindraPipeline) (bool, string, error) {

}

func (v *PipelineValidator) validatePipelineFn(ctx context.Context, pipeline *jindra.JindraPipeline) (bool, string, error) {
	valLog.Info("validating pipeline", "pipeline", pipeline.Name)
	reason := pipeline.Validate()

	if reason == nil {
		return true, "", nil
	}

	return false, reason.Error(), nil
}

// podValidator implements inject.Client.
// A client will be automatically injected.
var _ inject.Client = &PipelineValidator{}

// InjectClient injects the client.
func (v *PipelineValidator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// podValidator implements inject.Decoder.
// A decoder will be automatically injected.
var _ inject.Decoder = &PipelineValidator{}

// InjectDecoder injects the decoder.
func (v *PipelineValidator) InjectDecoder(d types.Decoder) error {
	v.decoder = d
	return nil
}
