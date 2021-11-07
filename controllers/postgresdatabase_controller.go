/*
Copyright 2021.

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
	"encoding/base64"
	"fmt"
	"reflect"

	pgx "github.com/jackc/pgx/v4"
	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dbv1beta1 "github.com/bujarmurati/pg-db-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DatabaseStateReconciler interface {
	GetConfig() (config *pgx.ConnConfig)
	ReconcileDatabaseState(userName string, databaseName string, password string) (err error)
}

// PostgresDatabaseReconciler reconciles a PostgresDatabase object
type PostgresDatabaseReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	Database          DatabaseStateReconciler
	PasswordGenerator password.PasswordGenerator
}

func b64decode(b []byte) (result string, err error) {
	decoded, err := base64.StdEncoding.DecodeString(string(b))
	return string(decoded), err
}

func b64encode(s string) []byte {
	src := []byte(s)
	encoded := base64.StdEncoding.EncodeToString(src)
	return []byte(encoded)
}

var gvk schema.GroupVersionKind = dbv1beta1.GroupVersion.WithKind(reflect.TypeOf(dbv1beta1.PostgresDatabase{}).Name())

//+kubebuilder:rbac:groups=db.bujarmurati.com,resources=postgresdatabases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=db.bujarmurati.com,resources=postgresdatabases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=db.bujarmurati.com,resources=postgresdatabases/finalizers,verbs=update
//+kubebuilder:rbac:groups=corev1,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=corev1,resources=secrets/status,verbs=get

func (r *PostgresDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var postgresDatabase dbv1beta1.PostgresDatabase
	if err := r.Get(ctx, req.NamespacedName, &postgresDatabase); err != nil {
		log.Error(err, "unable to fetch PostgresDatabase")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// TODO: configurable password policy
	password, err := r.PasswordGenerator.Generate(30, 5, 5, false, false)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	config := r.Database.GetConfig().Config
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      postgresDatabase.Spec.SecretName,
			Namespace: postgresDatabase.Namespace,
		},
		Data: map[string][]byte{
			"PGPASSWORD": b64encode(password),
			"PGHOST":     b64encode(config.Host),
			"PGPORT":     b64encode(fmt.Sprint(config.Port)),
			"PGUSER":     b64encode(postgresDatabase.Spec.DatabaseName + postgresDatabase.Spec.UserNamePostFix),
		},
	}
	ownerRef := metav1.NewControllerRef(&postgresDatabase, gvk)
	secret.SetOwnerReferences([]metav1.OwnerReference{*ownerRef})
	err = r.Create(ctx, secret)
	if err != nil {
		log.Error(err, "unable to create Secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgresDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbv1beta1.PostgresDatabase{}).
		Complete(r)
}
