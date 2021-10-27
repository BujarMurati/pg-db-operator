package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	dbv1beta1 "github.com/bujarmurati/pg-db-operator/api/v1beta1"
)

var _ = Describe("PostgresDatabase controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		PostgresDatabaseName      = "db"
		PostgresDatabaseNamespace = "default"
		DatabaseName              = "db"
		SecretName                = "secret-db"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating a PostgresDatabase", func() {
		It("Should update the secret", func() {
			ctx := context.Background()
			postgresDatabase := &dbv1beta1.PostgresDatabase{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "db.bujarmurati.com/v1beta1",
					Kind:       "PostgresDatabase",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      PostgresDatabaseName,
					Namespace: PostgresDatabaseNamespace,
				},
				Spec: dbv1beta1.PostgresDatabaseSpec{
					DatabaseName: "1 * * * *",
				},
			}
			Expect(k8sClient.Create(ctx, postgresDatabase)).Should(Succeed())
			pgdbLookupKey := types.NamespacedName{Name: PostgresDatabaseName, Namespace: PostgresDatabaseNamespace}
			createdPostgresDatabase := &dbv1beta1.PostgresDatabase{}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, pgdbLookupKey, createdPostgresDatabase)
				return err == nil
			}, timeout, interval).Should(BeTrue())
		})
	})

})
