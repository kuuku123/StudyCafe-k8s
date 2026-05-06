package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	studycafev1 "github.com/kuuku123/studycafe-operator/api/v1"
)

func (r *StudyCafeReconciler) reconcileService(ctx context.Context, sc *studycafev1.StudyCafe, name string, port int32) error {
	log := logf.FromContext(ctx)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: sc.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		labels := map[string]string{"app": name}

		if svc.Labels == nil {
			svc.Labels = labels
		} else {
			svc.Labels["app"] = name
		}

		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Selector = labels

		// Important: we need to update ports carefully because kubernetes services append
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Port:       port,
				TargetPort: intstr.FromInt32(port),
				Protocol:   corev1.ProtocolTCP,
			},
		}

		return controllerutil.SetControllerReference(sc, svc, r.Scheme)
	})

	if err != nil {
		return err
	}

	if op != controllerutil.OperationResultNone {
		log.Info("Service reconciled", "operation", op, "name", svc.Name)
	}

	return nil
}

func (r *StudyCafeReconciler) reconcileInfraService(ctx context.Context, sc *studycafev1.StudyCafe, infra infraDef) error {
	log := logf.FromContext(ctx)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infra.name,
			Namespace: sc.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		labels := map[string]string{"app": infra.name}

		if svc.Labels == nil {
			svc.Labels = labels
		} else {
			svc.Labels["app"] = infra.name
		}

		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Selector = labels

		var svcPorts []corev1.ServicePort
		for i, port := range infra.ports {
			name := ""
			if len(infra.ports) > 1 {
				if i == 0 {
					name = "external"
				} else if i == 1 {
					name = "internal"
				}
			}
			svcPorts = append(svcPorts, corev1.ServicePort{
				Name:       name,
				Port:       port,
				TargetPort: intstr.FromInt32(port),
				Protocol:   corev1.ProtocolTCP,
			})
		}
		svc.Spec.Ports = svcPorts

		return controllerutil.SetControllerReference(sc, svc, r.Scheme)
	})

	if err != nil {
		return err
	}
	if op != controllerutil.OperationResultNone {
		log.Info("Infra Service reconciled", "operation", op, "name", svc.Name)
	}

	return nil
}
