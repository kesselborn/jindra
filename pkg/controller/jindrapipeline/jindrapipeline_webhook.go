package jindrapipeline

// https://github.com/kubernetes-sigs/controller-runtime/blob/f1eaba5087d69cebb154c6a48193e6667f5b512c/example/main.go

// +kubebuilder:webhook:verbs=create,path=/validate-batch-tutorial-kubebuilder-io-v1-cronjob,mutating=false,failurePolicy=fail,groups=jindra.io,resources=jindrapipelines,versions=v1alpha1,name=jindrapipeline.jindra.io

import (
	"context"
	"fmt"
	"net/http"

	jindrav1alpha1 "github.com/kesselborn/jindra/pkg/apis/jindra/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

var webhooklog = logf.Log.WithName("jindrapipeline-validator")

type PipelineValidator struct {
	client  client.Client
	decoder types.Decoder
}

var _ admission.Handler = &PipelineValidator{}

// https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/types.go#L506 <- dry run pods
func (v *PipelineValidator) Handle(ctx context.Context, req types.Request) types.Response {
	webhooklog.Info("validating pipeline")

	pipeline := &jindrav1alpha1.JindraPipeline{}

	err := v.decoder.Decode(req, pipeline)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	allowed, reason, err := v.validatePipelineFn(ctx, pipeline)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}
	return admission.ValidationResponse(allowed, reason)
}

func (v *PipelineValidator) validatePipelineFn(ctx context.Context, pipeline *jindrav1alpha1.JindraPipeline) (bool, string, error) {
	key := "jindra-webhook-test"
	anno, found := pipeline.Annotations[key]
	switch {
	case !found:
		return found, fmt.Sprintf("failed to find annotation with key: %q", key), nil
	case found && anno == "foo":
		return found, "", nil
	case found && anno != "foo":
		return false,
			fmt.Sprintf("the value associate with key %q is expected to be %q, but got %q", key, "foo", anno), nil
	}
	return false, "", nil
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
