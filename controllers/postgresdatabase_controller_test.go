package controllers

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	dbv1beta1 "github.com/bujarmurati/pg-db-operator/api/v1beta1"
)

var _ = Describe("PostgresDatabase controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		PostgresDatabaseName = "db"
		DatabaseName         = "db"
		SecretName           = "secret-db"
		timeout              = time.Second * 10
		duration             = time.Second * 10
		interval             = time.Millisecond * 250
	)
	var PostgresDatabaseNamespace string
	BeforeEach(func() {
		currentTest := CurrentGinkgoTestDescription().TestText
		PostgresDatabaseNamespace = strings.ReplaceAll(strings.ToLower(currentTest), " ", "-")
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: PostgresDatabaseNamespace,
			},
		}
		_, err := clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		if !SkipNamespaceCleanup {
			Expect(clientset.CoreV1().Namespaces().Delete(context.Background(), PostgresDatabaseNamespace, metav1.DeleteOptions{})).NotTo(HaveOccurred())
		}
	})

	Context("When creating a PostgresDatabase", func() {
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
					DatabaseName: "db_" + PostgresDatabaseNamespace,
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
