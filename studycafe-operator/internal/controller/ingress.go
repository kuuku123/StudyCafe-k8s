package controller

import (
	"context"

	studycafev1 "github.com/kuuku123/studycafe-operator/api/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *StudyCafeReconciler) reconcileIngress(ctx context.Context, studycafe *studycafev1.StudyCafe) error {
	ingressClassName := "nginx"
	pathTypePrefix := networkingv1.PathTypePrefix
	pathTypeImplSpecific := networkingv1.PathTypeImplementationSpecific

	// 1. React Frontend Ingress (HTTP)
	reactIngress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      studycafe.Name + "-ingress", // Keeping the original name for the HTTP one
			Namespace: studycafe.Namespace,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, reactIngress, func() error {
		if reactIngress.Labels == nil {
			reactIngress.Labels = make(map[string]string)
		}
		reactIngress.Labels["app"] = studycafe.Name

		reactIngress.Spec.IngressClassName = &ingressClassName
		reactIngress.Spec.Rules = []networkingv1.IngressRule{
			{
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathTypePrefix,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: "react-frontend",
										Port: networkingv1.ServiceBackendPort{
											Number: 80,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		return controllerutil.SetControllerReference(studycafe, reactIngress, r.Scheme)
	})
	if err != nil {
		return err
	}

	// 2. API Gateway Ingress (HTTP)
	apiIngress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      studycafe.Name + "-api-ingress",
			Namespace: studycafe.Namespace,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, apiIngress, func() error {
		if apiIngress.Labels == nil {
			apiIngress.Labels = make(map[string]string)
		}
		apiIngress.Labels["app"] = studycafe.Name

		if apiIngress.Annotations == nil {
			apiIngress.Annotations = make(map[string]string)
		}
		// We use plain HTTP for internal traffic between Ingress and api-gateway, avoiding TLS mismatch
		// apiIngress.Annotations["nginx.ingress.kubernetes.io/backend-protocol"] = "HTTPS"
		apiIngress.Annotations["nginx.ingress.kubernetes.io/use-regex"] = "true"
		apiIngress.Annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/$2"

		apiIngress.Spec.IngressClassName = &ingressClassName
		apiIngress.Spec.Rules = []networkingv1.IngressRule{
			{
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/api(/|$)(.*)",
								PathType: &pathTypeImplSpecific,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: "api-gateway",
										Port: networkingv1.ServiceBackendPort{
											Number: 8083,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		return controllerutil.SetControllerReference(studycafe, apiIngress, r.Scheme)
	})

	return err
}
