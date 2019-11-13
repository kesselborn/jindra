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
	"encoding/json"
	"net/http"

	jindra "github.com/kesselborn/jindra/pkg/apis/jindra/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var mutLog = logf.Log.WithName("jindra-mutator")

// +kubebuilder:webhook:path=/mutate-v1alpha1-jindrapipeline,mutating=true,failurePolicy=fail,groups="",resources=jindrapipeline,verbs=create;update,versions=v1alpha1,name=mjppl.jindra.io

// PipelineMutator sets default values for pipeline objects
type PipelineMutator struct {
	client  client.Client
	decoder *admission.Decoder
}

// Implement admission.Handler so the controller can handle admission request.
func (a *PipelineMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pipeline := &jindra.JindraPipeline{}

	err := a.decoder.Decode(req, pipeline)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	mutLog.Info("mutating pipeline", "pipeline", pipeline.Name)
	copy := pipeline.DeepCopy()

	err = a.mutatePipelineFn(ctx, copy)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	marshaledPpl, err := json.Marshal(pipeline)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPpl)
}

// mutatePipelineFn add an annotation to the given pod
func (a *PipelineMutator) mutatePipelineFn(ctx context.Context, ppl *jindra.JindraPipeline) error {
	logf.Log.Info("setting default values", "pipeline", ppl.Name)
	/*
		modifiedItems := ppl.SetDefaults()
		if len(modifiedItems) > 0 {
			logf.Log.Info("modified pipeline", modifiedItems...)
		}
	*/
	return nil
}

// podAnnotator implements inject.Client.
// A client will be automatically injected.

// InjectClient injects the client.
func (a *PipelineMutator) InjectClient(c client.Client) error {
	a.client = c
	return nil
}

// podAnnotator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (a *PipelineMutator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
