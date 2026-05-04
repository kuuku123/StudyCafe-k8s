/*
Copyright 2026.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "github.com/kuuku123/studycafe-operator/api/v1"
)

// StudyCafeReconciler reconciles a StudyCafe object
type StudyCafeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.studycafe.io,resources=studycafes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.studycafe.io,resources=studycafes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.studycafe.io,resources=studycafes/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *StudyCafeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Starting reconciliation", "request", req.NamespacedName)

	// 1. Fetch the StudyCafe instance
	studycafe := &appsv1.StudyCafe{}
	err := r.Get(ctx, req.NamespacedName, studycafe)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// 2. Define the desired ConfigMap
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      studycafe.Name + "-info",
			Namespace: studycafe.Namespace,
		},
	}

	// 3. Create or Update the ConfigMap
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		cm.Data["message"] = studycafe.Spec.Message
		cm.Data["owner"] = studycafe.Name

		// Set owner reference so it's deleted when StudyCafe is deleted
		return controllerutil.SetControllerReference(studycafe, cm, r.Scheme)
	})

	if err != nil {
		log.Error(err, "unable to create or update ConfigMap")
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		log.Info("ConfigMap reconciled", "operation", op, "name", cm.Name)
	}

	log.Info("Successfully reconciled StudyCafe", "name", studycafe.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StudyCafeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.StudyCafe{}).
		Owns(&corev1.ConfigMap{}).
		Named("studycafe").
		Complete(r)
}
