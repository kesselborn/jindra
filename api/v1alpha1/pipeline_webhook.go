/*

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var webhLog = logf.Log.WithName("pipeline-webook")

// SetupWebhookWithManager registeres this webhook with a manager
func (ppl *Pipeline) SetupWebhookWithManager(mgr ctrl.Manager) error {
	webhLog.Info("Setting up Webhook")
	return ctrl.NewWebhookManagedBy(mgr).
		For(ppl).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-ci-jindra-io-v1alpha1-pipeline,mutating=true,failurePolicy=fail,groups=ci.jindra.io,resources=pipelines,verbs=create;update,versions=v1alpha1,name=defaulter.jindra.io

var _ webhook.Defaulter = &Pipeline{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (ppl *Pipeline) Default() {
	webhLog.Info("calling defaulter for ", "pipeline", ppl.Name)

	ppl.SetDefaults()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-ci-jindra-io-v1alpha1-pipeline,mutating=false,failurePolicy=fail,groups=ci.jindra.io,resources=pipelines,versions=v1alpha1,name=validator.jindra.io

var _ webhook.Validator = &Pipeline{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (ppl *Pipeline) ValidateCreate() error {
	webhLog.Info("calling create validator for ", "pipeline", ppl.Name)
	return ppl.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (ppl *Pipeline) ValidateUpdate(old runtime.Object) error {
	webhLog.Info("calling update validator for ", "pipeline", ppl.Name)
	return ppl.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (ppl *Pipeline) ValidateDelete() error {
	// webhLog.Info("calling delete validator for ", "pipeline", ppl.Name)
	webhLog.Info("delete validator not implemented", "pipeline", ppl.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
