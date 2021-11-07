package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	dbv1beta1 "github.com/bujarmurati/pg-db-operator/api/v1beta1"
	"github.com/bujarmurati/pg-db-operator/database"
)

const (
	PostgresDatabaseName = "db"
	timeout              = time.Second * 10
	duration             = time.Second * 10
	interval             = time.Millisecond * 250
)

func assertDatumInSecret(key, value string, data map[string][]byte) {
	Expect(data).Should(HaveKey(key))
	actualValue, err := b64decode(data[key])
	Expect(err).NotTo(HaveOccurred())
	Expect(actualValue).To(Equal(value))
}

var DefaultPostgresDatabase = &dbv1beta1.PostgresDatabase{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "db.bujarmurati.com/v1beta1",
		Kind:       "PostgresDatabase",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: PostgresDatabaseName,
	},
}

func defaultPostgresDatabase(namespace string) *dbv1beta1.PostgresDatabase {
	return &dbv1beta1.PostgresDatabase{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "db.bujarmurati.com/v1beta1",
			Kind:       "PostgresDatabase",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      PostgresDatabaseName,
			Namespace: namespace,
		},
		Spec: dbv1beta1.PostgresDatabaseSpec{
			DatabaseName: "db_" + strings.ReplaceAll(namespace, "-", "_"),
			SecretName:   "secret-" + namespace,
		},
	}
}

func waitForPostgresDatabaseResourceCreation(ctx context.Context, namespace string, pgdb *dbv1beta1.PostgresDatabase) *dbv1beta1.PostgresDatabase {

	Expect(k8sClient.Create(ctx, pgdb)).Should(Succeed())
	pgdbLookupKey := types.NamespacedName{Name: pgdb.Name, Namespace: namespace}
	createdPostgresDatabase := &dbv1beta1.PostgresDatabase{}

	Eventually(func() bool {
		err := k8sClient.Get(ctx, pgdbLookupKey, createdPostgresDatabase)
		return err == nil
	}, timeout, interval).Should(BeTrue())
	return createdPostgresDatabase
}

var _ = Describe("createConnectionConfigFromSpec", func() {
	It("creates a correct config", func() {
		result := createConnectionConfigFromSpec(ExpectedPassword, dbv1beta1.PostgresDatabaseSpec{}, connConfig)
		Expect(result.Config.Password).To(Equal(ExpectedPassword))
		connString := createConnectionString(result)
		Expect(connString).To(ContainSubstring(ExpectedPassword))

	})
})

var _ = Describe("PostgresDatabase controller", func() {

	kind := reflect.TypeOf(dbv1beta1.PostgresDatabase{}).Name()
	gvk := dbv1beta1.GroupVersion.WithKind(kind)
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
		if !skipNamespaceCleanup {
			Expect(clientset.CoreV1().Namespaces().Delete(context.Background(), PostgresDatabaseNamespace, metav1.DeleteOptions{})).NotTo(HaveOccurred())
		}
	})

	Context("When creating a PostgresDatabase", func() {
		Context("and no matching secret exists", func() {
			It("Should create a new secret", func() {
				ctx := context.Background()
				postgresDatabase := defaultPostgresDatabase(PostgresDatabaseNamespace)
				postgresDatabase.Spec.UserNamePostFix = "@" + connConfig.Config.Host
				createdPostgresDatabase := waitForPostgresDatabaseResourceCreation(ctx, PostgresDatabaseNamespace, postgresDatabase)
				By("Creating a new opaque secret")
				var createdSecret *corev1.Secret
				Eventually(func() (err error) {
					createdSecret, err = clientset.CoreV1().Secrets(PostgresDatabaseNamespace).Get(context.Background(), postgresDatabase.Spec.SecretName, metav1.GetOptions{})
					return err
				}).Should(Succeed())
				Expect(string(createdSecret.Type)).To(Equal("Opaque"))
				targetConfig := createConnectionConfigFromSpec(ExpectedPassword, postgresDatabase.Spec, connConfig)
				By("Creating the secret with a correct data")
				assertDatumInSecret("PG_CONNECTION_STRING", createConnectionString(targetConfig), createdSecret.Data)
				connString, err := b64decode(createdSecret.Data["PG_CONNECTION_STRING"])
				Expect(err).NotTo(HaveOccurred())
				Expect(strings.Contains(connString, ExpectedPassword)).To(BeTrue())
				assertDatumInSecret("PGPASSWORD", ExpectedPassword, createdSecret.Data)
				assertDatumInSecret("PGUSER", postgresDatabase.Spec.DatabaseName+postgresDatabase.Spec.UserNamePostFix, createdSecret.Data)
				assertDatumInSecret("PGHOST", connConfig.Config.Host, createdSecret.Data)
				assertDatumInSecret("PGPORT", fmt.Sprint(connConfig.Config.Port), createdSecret.Data)

				By("Setting an OwnerReference on the secret")
				expectedOwnerReference := metav1.NewControllerRef(createdPostgresDatabase, gvk)
				Expect(createdSecret.ObjectMeta.OwnerReferences).To(ContainElement(*expectedOwnerReference))
			})
		})
		Context("With the new secret's connection string", func() {
			It("Should be possible to connect to the database", func() {
				ctx := context.Background()
				postgresDatabase := defaultPostgresDatabase(PostgresDatabaseNamespace)
				waitForPostgresDatabaseResourceCreation(ctx, PostgresDatabaseNamespace, postgresDatabase)

				var createdSecret *corev1.Secret
				Eventually(func() (err error) {
					createdSecret, err = clientset.CoreV1().Secrets(PostgresDatabaseNamespace).Get(context.Background(), postgresDatabase.Spec.SecretName, metav1.GetOptions{})
					return err
				}, time.Second*5).Should(Succeed())
				connString, err := b64decode(createdSecret.Data["PG_CONNECTION_STRING"])
				Expect(err).NotTo(HaveOccurred())
				config, err := pgx.ParseConfig(connString)
				Expect(err).NotTo(HaveOccurred())
				Eventually(func() error {
					return database.CheckConnection(*config)
				}, time.Second*5).Should(Succeed())
			})
		})

	})

})
