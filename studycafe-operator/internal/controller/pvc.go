package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	studycafev1 "github.com/kuuku123/studycafe-operator/api/v1"
)

func (r *StudyCafeReconciler) reconcilePVC(ctx context.Context, sc *studycafev1.StudyCafe, name, size string) error {
	log := logf.FromContext(ctx)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: sc.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}

		if pvc.Spec.Resources.Requests == nil {
			pvc.Spec.Resources.Requests = corev1.ResourceList{}
		}

		// Only set storage if it's not set (PVCs can't be easily resized like this)
		if _, exists := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; !exists {
			pvc.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(size)
		}

		return controllerutil.SetControllerReference(sc, pvc, r.Scheme)
	})

	if err != nil {
		return err
	}
	if op != controllerutil.OperationResultNone {
		log.Info("PVC reconciled", "operation", op, "name", pvc.Name)
	}
	return nil
}
