package controller

import (
	"context"
	"fmt"

	appsv1k8s "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	studycafev1 "github.com/kuuku123/studycafe-operator/api/v1"
)

func (r *StudyCafeReconciler) reconcileDeployment(ctx context.Context, sc *studycafev1.StudyCafe, app appDef) error {
	log := logf.FromContext(ctx)

	dep := &appsv1k8s.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.name,
			Namespace: sc.Namespace,
		},
	}

	replicas := sc.Spec.Replicas
	if replicas == 0 {
		replicas = 1
	}

	imageTag := sc.Spec.ImageTag
	if imageTag == "" {
		imageTag = "latest"
	}
	fullImage := fmt.Sprintf("%s:%s", app.image, imageTag)

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, dep, func() error {
		labels := map[string]string{"app": app.name}

		if dep.Labels == nil {
			dep.Labels = labels
		} else {
			dep.Labels["app"] = app.name
		}

		dep.Spec.Replicas = &replicas
		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: labels,
		}

		dep.Spec.Template.ObjectMeta.Labels = labels

		// Define container
		container := corev1.Container{
			Name:            app.name,
			Image:           fullImage,
			ImagePullPolicy: corev1.PullNever,
			Ports: []corev1.ContainerPort{
				{ContainerPort: app.port},
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse(app.memoryRequest),
					corev1.ResourceCPU:    resource.MustParse(app.cpuRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse(app.memoryLimit),
					corev1.ResourceCPU:    resource.MustParse(app.cpuLimit),
				},
			},
		}

		if app.command != nil {
			container.Command = app.command
		}

		// Environment variables
		if app.needsJasypt {
			container.Env = append(container.Env, corev1.EnvVar{
				Name: "JASYPT_ENCRYPTOR_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "jasypt-secret"},
						Key:                  "JASYPT_ENCRYPTOR_PASSWORD",
					},
				},
			})
		}

		if app.needsJwt {
			container.Env = append(container.Env, corev1.EnvVar{
				Name: "JWT_PRIVATE_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "jwt-secret"},
						Key:                  "JWT_PRIVATE_KEY",
					},
				},
			})
			container.Env = append(container.Env, corev1.EnvVar{
				Name: "JWT_PUBLIC_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "jwt-secret"},
						Key:                  "JWT_PUBLIC_KEY",
					},
				},
			})
		}

		if app.name == "api-gateway" || app.name == "notification-service" || app.name == "auth-service" || app.name == "study-service" {
			envVal := sc.Spec.Environment
			if envVal == "" {
				envVal = "kube"
			}
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  "SPRING_PROFILES_ACTIVE",
				Value: envVal,
			})
		}

		// Probes
		readinessProbe := &corev1.Probe{
			InitialDelaySeconds: 40,
			PeriodSeconds:       10,
			TimeoutSeconds:      5,
			FailureThreshold:    10,
		}
		livenessProbe := &corev1.Probe{
			InitialDelaySeconds: 180,
			PeriodSeconds:       15,
			TimeoutSeconds:      5,
		}

		if app.isHttpProbe {
			readinessProbe.ProbeHandler = corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: app.probePath,
					Port: intstr.FromInt32(app.port),
				},
			}
			livenessProbe.ProbeHandler = corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: app.probePath,
					Port: intstr.FromInt32(app.port),
				},
			}
		} else {
			readinessProbe.ProbeHandler = corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt32(app.port),
				},
			}
			livenessProbe.ProbeHandler = corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt32(app.port),
				},
			}
		}

		if app.name == "auth-service" || app.name == "study-service" {
			readinessProbe.InitialDelaySeconds = 60
		} else if app.name == "react-frontend" {
			readinessProbe.InitialDelaySeconds = 10
			readinessProbe.PeriodSeconds = 5
			readinessProbe.TimeoutSeconds = 3
			livenessProbe.PeriodSeconds = 10
			livenessProbe.TimeoutSeconds = 3
		}

		container.ReadinessProbe = readinessProbe
		container.LivenessProbe = livenessProbe

		dep.Spec.Template.Spec.Containers = []corev1.Container{container}

		return controllerutil.SetControllerReference(sc, dep, r.Scheme)
	})

	if err != nil {
		return err
	}

	if op != controllerutil.OperationResultNone {
		log.Info("Deployment reconciled", "operation", op, "name", dep.Name)
	}

	return nil
}

func (r *StudyCafeReconciler) reconcileInfraDeployment(ctx context.Context, sc *studycafev1.StudyCafe, infra infraDef) error {
	log := logf.FromContext(ctx)

	dep := &appsv1k8s.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infra.name,
			Namespace: sc.Namespace,
		},
	}

	replicas := int32(1)

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, dep, func() error {
		labels := map[string]string{"app": infra.name}

		if dep.Labels == nil {
			dep.Labels = labels
		} else {
			dep.Labels["app"] = infra.name
		}

		dep.Spec.Replicas = &replicas
		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: labels,
		}

		if infra.isStateful {
			dep.Spec.Strategy = appsv1k8s.DeploymentStrategy{
				Type: appsv1k8s.RecreateDeploymentStrategyType,
			}
		}

		dep.Spec.Template.ObjectMeta.Labels = labels
		if infra.hostname != "" {
			dep.Spec.Template.Spec.Hostname = infra.hostname
		}

		container := corev1.Container{
			Name:            infra.name,
			Image:           infra.image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse(infra.memoryRequest),
					corev1.ResourceCPU:    resource.MustParse(infra.cpuRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse(infra.memoryLimit),
					corev1.ResourceCPU:    resource.MustParse(infra.cpuLimit),
				},
			},
			Env: infra.env,
		}

		if infra.command != nil {
			container.Command = infra.command
		}
		if infra.args != nil {
			container.Args = infra.args
		}

		var containerPorts []corev1.ContainerPort
		for _, port := range infra.ports {
			containerPorts = append(containerPorts, corev1.ContainerPort{ContainerPort: port})
		}
		container.Ports = containerPorts

		if infra.pvcName != "" && infra.mountPath != "" {
			container.VolumeMounts = []corev1.VolumeMount{
				{
					Name:      infra.name + "-data",
					MountPath: infra.mountPath,
				},
			}
			dep.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: infra.name + "-data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: infra.pvcName,
						},
					},
				},
			}
		}

		if infra.readinessCmd != nil {
			container.ReadinessProbe = &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{Command: infra.readinessCmd},
				},
				InitialDelaySeconds: infra.readinessDelay,
				PeriodSeconds:       10,
				TimeoutSeconds:      5,
			}
		} else if infra.tcpProbePort != 0 {
			container.ReadinessProbe = &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(infra.tcpProbePort)},
				},
				InitialDelaySeconds: infra.readinessDelay,
				PeriodSeconds:       10,
				TimeoutSeconds:      5,
				FailureThreshold:    3,
			}
		}

		if infra.livenessCmd != nil {
			container.LivenessProbe = &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{Command: infra.livenessCmd},
				},
				InitialDelaySeconds: infra.livenessDelay,
				PeriodSeconds:       15,
				TimeoutSeconds:      5,
			}
		}

		dep.Spec.Template.Spec.Containers = []corev1.Container{container}

		return controllerutil.SetControllerReference(sc, dep, r.Scheme)
	})

	if err != nil {
		return err
	}
	if op != controllerutil.OperationResultNone {
		log.Info("Infra Deployment reconciled", "operation", op, "name", dep.Name)
	}

	return nil
}
