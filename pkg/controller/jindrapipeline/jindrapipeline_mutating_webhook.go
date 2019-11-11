package jindrapipeline

import (
	"context"
	"net/http"

	jindra "github.com/kesselborn/jindra/pkg/apis/jindra/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

var mutLog = logf.Log.WithName("jindra-mutator")

// PipelineMutator sets default values for pipeline objects
type PipelineMutator struct {
	client  client.Client
	decoder types.Decoder
}

// Implement admission.Handler so the controller can handle admission request.
var _ admission.Handler = &PipelineMutator{}

// Handle handles incoming requests
func (a *PipelineMutator) Handle(ctx context.Context, req types.Request) types.Response {
	pipeline := &jindra.JindraPipeline{}

	err := a.decoder.Decode(req, pipeline)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	mutLog.Info("mutating pipeline", "pipeline", pipeline.Name)
	copy := pipeline.DeepCopy()

	err = a.mutatePipelineFn(ctx, copy)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}
	return admission.PatchResponse(pipeline, copy)
}

// mutatePipelineFn add an annotation to the given pod
func (a *PipelineMutator) mutatePipelineFn(ctx context.Context, ppl *jindra.JindraPipeline) error {
	logf.Log.Info("setting default values", "pipeline", ppl.Name)
	modifiedItems := ppl.SetDefaults()
	if len(modifiedItems) > 0 {
		logf.Log.Info("modified pipeline", modifiedItems...)
	}
	return nil
}

// podAnnotator implements inject.Client.
// A client will be automatically injected.
var _ inject.Client = &PipelineMutator{}

// InjectClient injects the client.
func (v *PipelineMutator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// podAnnotator implements inject.Decoder.
// A decoder will be automatically injected.
var _ inject.Decoder = &PipelineMutator{}

// InjectDecoder injects the decoder.
func (v *PipelineMutator) InjectDecoder(d types.Decoder) error {
	v.decoder = d
	return nil
}
