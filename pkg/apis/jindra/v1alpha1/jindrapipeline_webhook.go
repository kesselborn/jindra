package v1alpha1

import (
	"fmt"

	core "k8s.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var log = logf.Log.WithName("jindra-resource")

func (r *JindraPipeline) SetupWebhookWithManager(mgr ctrl.Manager) error {
	log.Info("SetupWebhookWithManager called")
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Defaulter = &JindraPipeline{}

// +kubebuilder:webhook:path=/mutate-v1alpha1-jindrapipeline,mutating=true,failurePolicy=fail,groups="",resources=jindrapipeline,verbs=create;update,versions=v1alpha1,name=mjppl.jindra.io
func (ppl *JindraPipeline) Default() {
	modifiedItems := []interface{}{}
	// only set name if final is actually set (annotations or containers are set)
	if (len(ppl.Spec.Final.Annotations) > 0 || len(ppl.Spec.Final.Spec.Containers) > 0) && ppl.Spec.Final.Name == "" {
		modifiedItems = append(modifiedItems, "final/name", "final")
		ppl.Spec.Final.Name = "final"
	}

	// only set name if on-error is actually set (annotations or containers are set)
	if (len(ppl.Spec.OnError.Annotations) > 0 || len(ppl.Spec.OnError.Spec.Containers) > 0) && ppl.Spec.OnError.Name == "" {
		modifiedItems = append(modifiedItems, "on-error/name", "on-error")
		ppl.Spec.OnError.Name = "on-error"
	}

	// only set name if on-success is actually set (annotations or containers are set)
	if (len(ppl.Spec.OnSuccess.Annotations) > 0 || len(ppl.Spec.OnSuccess.Spec.Containers) > 0) && ppl.Spec.OnSuccess.Name == "" {
		modifiedItems = append(modifiedItems, "on-success/name", "on-success")
		ppl.Spec.OnSuccess.Name = "on-success"
	}

	for i := 0; i < len(ppl.Spec.Resources.Triggers); i++ {
		if ppl.Spec.Resources.Triggers[i].Schedule == "" {
			modifiedItems = append(modifiedItems, "trigger/"+ppl.Spec.Resources.Triggers[i].Name+"/name/schedule", "/5 * * * *")
			ppl.Spec.Resources.Triggers[i].Schedule = "/5 * * * *"
		}
	}

	if ppl.Annotations[BuildNoOffsetAnnotationKey] == "" {
		modifiedItems = append(modifiedItems, "/build-offset", "0")
		ppl.Annotations[BuildNoOffsetAnnotationKey] = "0"
	}

	for i := 0; i < len(ppl.Spec.Stages); i++ {
		if ppl.Spec.Stages[i].Spec.RestartPolicy == core.RestartPolicy("") {
			modifiedItems = append(modifiedItems, "/stage/"+ppl.Spec.Stages[i].Name+"/restart-policy", "never")
			ppl.Spec.Stages[i].Spec.RestartPolicy = core.RestartPolicyNever
		}

	}
	for _, pod := range []*core.Pod{&ppl.Spec.OnSuccess, &ppl.Spec.OnError, &ppl.Spec.Final} {
		if (len(pod.Spec.Containers) > 0) && pod.Spec.RestartPolicy == core.RestartPolicy("") {
			modifiedItems = append(modifiedItems, "/stage/"+pod.Name+"/restart-policy", "never")
			pod.Spec.RestartPolicy = core.RestartPolicyNever
		}
	}

	fmt.Printf("%#v", modifiedItems)
}
