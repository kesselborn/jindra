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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	core "k8s.io/api/core/v1"
)

var (
	jindraEventLabels = map[string]string{"app": "jindra", "type": "trigger"}
)

type JindraEventReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete

func isJindraEvent(e core.Event) bool {
	labels := e.ObjectMeta.Labels

	for k, v := range jindraEventLabels {
		if labels[k] != v {
			return false
		}
	}

	return true
}

func (r *JindraEventReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("event", req.NamespacedName, "ctx", ctx)

	var event core.Event
	if err := r.Get(ctx, req.NamespacedName, &event); err != nil {
		// log.Error(err, "unable to fetch Event", "req", fmt.Sprintf("%#v", req), "err", fmt.Sprintf("%#v", err))
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, ignoreNotFound(err)
	}
	if !isJindraEvent(event) {
		return ctrl.Result{}, nil
	}

	log.Info("got event", "event", event)

	// r.Delete(ctx, &event)
	event.ObjectMeta.Annotations["processed"] = "true"
	if err := r.Update(ctx, &event); err != nil {
		log.Error(err, "error setting event 'processed' annotation to 'done'")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *JindraEventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&core.Event{}).
		Complete(r)
}
