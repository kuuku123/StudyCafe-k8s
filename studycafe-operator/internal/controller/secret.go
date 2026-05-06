package controller

import (
	"context"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	studycafev1 "github.com/kuuku123/studycafe-operator/api/v1"
)

func (r *StudyCafeReconciler) reconcileSecrets(ctx context.Context, sc *studycafev1.StudyCafe) error {
	log := logf.FromContext(ctx)

	// 1. Jasypt Secret
	jasyptPass := os.Getenv("JASYPT_ENCRYPTOR_PASSWORD")
	if jasyptPass != "" {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "jasypt-secret",
				Namespace: sc.Namespace,
			},
		}
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
			if secret.StringData == nil {
				secret.StringData = make(map[string]string)
			}
			secret.StringData["JASYPT_ENCRYPTOR_PASSWORD"] = jasyptPass
			return controllerutil.SetControllerReference(sc, secret, r.Scheme)
		})
		if err != nil {
			return err
		}
		log.Info("Reconciled jasypt-secret from environment variable")
	}

	// 2. JWT Secret
	jwtPrivate := os.Getenv("JWT_PRIVATE_KEY")
	jwtPublic := os.Getenv("JWT_PUBLIC_KEY")
	if jwtPrivate != "" && jwtPublic != "" {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "jwt-secret",
				Namespace: sc.Namespace,
			},
		}
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
			if secret.StringData == nil {
				secret.StringData = make(map[string]string)
			}
			secret.StringData["JWT_PRIVATE_KEY"] = jwtPrivate
			secret.StringData["JWT_PUBLIC_KEY"] = jwtPublic
			return controllerutil.SetControllerReference(sc, secret, r.Scheme)
		})
		if err != nil {
			return err
		}
		log.Info("Reconciled jwt-secret from environment variables")
	}

	// 3. MySQL Secret
	mysqlPass := os.Getenv("MYSQL_ROOT_PASSWORD")
	if mysqlPass != "" {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysql-secret",
				Namespace: sc.Namespace,
			},
		}
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
			if secret.StringData == nil {
				secret.StringData = make(map[string]string)
			}
			secret.StringData["MYSQL_ROOT_PASSWORD"] = mysqlPass
			return controllerutil.SetControllerReference(sc, secret, r.Scheme)
		})
		if err != nil {
			return err
		}
		log.Info("Reconciled mysql-secret from environment variable")
	}

	return nil
}
