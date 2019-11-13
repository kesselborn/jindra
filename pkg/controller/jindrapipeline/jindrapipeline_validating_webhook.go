/*
Copyright 2018 The Kubernetes Authors.

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

package jindrapipeline

import (
	"context"
	"net/http"

	jindra "github.com/kesselborn/jindra/pkg/apis/jindra/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var valLog = logf.Log.WithName("jindra-validator")

// +kubebuilder:webhook:path=/validate-v1alpha1-jindrapipeline,mutating=false,failurePolicy=fail,groups="",resources=jindrapipelines,verbs=create;update,versions=v1alpha1,name=vjppl.jindra.io

// PipelineValidator validates pipeline objects
type PipelineValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// Handle handles incoming requests
// note to self: https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/types.go#L506 <- dry run pods
func (v *PipelineValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	valLog.Info("validating pipeline")

	pipeline := &jindra.JindraPipeline{}

	err := v.decoder.Decode(req, pipeline)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	allowed, reason, err := v.validatePipelineFn(ctx, pipeline)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if !allowed {
		return admission.Denied(reason)
	}

	allowed, reason, err = v.ValidateStagesFn(ctx, pipeline)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if !allowed {
		return admission.Denied(reason)
	}

	return admission.Allowed("")
}

func (v *PipelineValidator) ValidateStagesFn(ctx context.Context, pipeline *jindra.JindraPipeline) (bool, string, error) {
	return true, "", nil
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

// InjectClient injects the client.
func (v *PipelineValidator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// podValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *PipelineValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
