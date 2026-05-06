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

	appsv1k8s "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	studycafev1 "github.com/kuuku123/studycafe-operator/api/v1"
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
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *StudyCafeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Starting reconciliation", "request", req.NamespacedName)

	studycafe := &studycafev1.StudyCafe{}
	err := r.Get(ctx, req.NamespacedName, studycafe)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 1. Reconcile Secrets first (from environment variables)
	if err := r.reconcileSecrets(ctx, studycafe); err != nil {
		log.Error(err, "Failed to reconcile Secrets")
		return ctrl.Result{}, err
	}

	infras := getInfras()

	for _, infra := range infras {
		if infra.pvcName != "" {
			if err := r.reconcilePVC(ctx, studycafe, infra.pvcName, infra.pvcSize); err != nil {
				log.Error(err, "Failed to reconcile PVC", "PVC.Name", infra.pvcName)
				return ctrl.Result{}, err
			}
		}
		if err := r.reconcileInfraDeployment(ctx, studycafe, infra); err != nil {
			log.Error(err, "Failed to reconcile Infra Deployment", "Deployment.Name", infra.name)
			return ctrl.Result{}, err
		}
		if err := r.reconcileInfraService(ctx, studycafe, infra); err != nil {
			log.Error(err, "Failed to reconcile Infra Service", "Service.Name", infra.name)
			return ctrl.Result{}, err
		}
	}

	apps := getApps()

	for _, app := range apps {
		if err := r.reconcileDeployment(ctx, studycafe, app); err != nil {
			log.Error(err, "Failed to reconcile Deployment", "Deployment.Name", app.name)
			return ctrl.Result{}, err
		}
		if err := r.reconcileService(ctx, studycafe, app.name, app.port); err != nil {
			log.Error(err, "Failed to reconcile Service", "Service.Name", app.name)
			return ctrl.Result{}, err
		}
	}

	// Keep existing ConfigMap logic
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      studycafe.Name + "-info",
			Namespace: studycafe.Namespace,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		cm.Data["message"] = studycafe.Spec.Message
		cm.Data["owner"] = studycafe.Name
		return controllerutil.SetControllerReference(studycafe, cm, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled StudyCafe", "name", studycafe.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StudyCafeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&studycafev1.StudyCafe{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1k8s.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Named("studycafe").
		Complete(r)
}
