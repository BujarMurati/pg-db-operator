package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func setServerAdminCredentials(port string, host string) {
	os.Setenv("PGPASSWORD", "test")
	os.Setenv("PGUSER", "server_admin")
	os.Setenv("PGHOST", host)
	os.Setenv("PGPORT", port)
	os.Setenv("PGDATABASE", "postgres")
}

func unsetServerAdminCredentials() {
	os.Unsetenv("PGPASSWORD")
	os.Unsetenv("PGUSER")
	os.Unsetenv("PGHOST")
	os.Unsetenv("PGPORT")
	os.Unsetenv("PGDATABASE")
}

func CreatePostgresContainer() (postgresContainer testcontainers.Container, err error) {
	packageName := "database"
	workingDir, _ := os.Getwd()
	rootDir := strings.Replace(workingDir, packageName, "", 1)
	mountFrom := fmt.Sprintf("%s/test_resources/init.sql", rootDir)
	mountTo := "/docker-entrypoint-initdb.d/init.sql"
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:14",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForAll(wait.ForListeningPort("5432/tcp"), wait.ForLog("init.sql")),
		BindMounts:   map[string]string{mountFrom: mountTo},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_DATABASE": "postgres",
		},
	}
	postgresContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	return postgresContainer, err
}

var _ = Describe("DatabaseServer", func() {
	var postgresContainer testcontainers.Container
	const postgresStartupTimeout = time.Second * 10

	BeforeEach(func() {
		Eventually(func() error {
			var err error
			postgresContainer, err = CreatePostgresContainer()
			return err
		}, postgresStartupTimeout).ShouldNot(HaveOccurred())
		mappedPort, err := postgresContainer.MappedPort(context.Background(), "5432")
		if err != nil {
			Fail("Failed to get mapped port")
		}
		ip, err := postgresContainer.Host(context.Background())
		if err != nil {
			Fail("Failed to get container host")
		}
		setServerAdminCredentials(mappedPort.Port(), ip)
	})

	Describe("NewDatabaseServerFromEnvironment", func() {
		It("Can use libpq environment variables (a.k.a. PGHOST, etc.)", func() {
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			Expect(db.ConnectionPool.Ping(context.Background())).NotTo(HaveOccurred())
		})
	})
	Describe("CheckDatabaseExists", func() {
		It("Returns false if the database does not exist", func() {
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			exists, err := db.CheckDatabaseExists("database_does_not_exist")
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})
		It("Returns true if the database already does exist", func() {
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			exists, err := db.CheckDatabaseExists("database_exists")
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
	})
	Describe("CreateDatabaseIfNotExists", func() {
		It("Creates a new database if it doesn't exist", func() {
			databaseName := "new_database"
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			exists, err := db.CheckDatabaseExists(databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())

			Expect(db.CreateDatabaseIfNotExists(databaseName)).NotTo(HaveOccurred())
			exists, err = db.CheckDatabaseExists(databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
		It("Skips creating a new database if it doesn't exist", func() {
			databaseName := "database_exists"
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			exists, err := db.CheckDatabaseExists(databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			Expect(db.CreateDatabaseIfNotExists(databaseName)).NotTo(HaveOccurred())
			exists, err = db.CheckDatabaseExists(databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
	})

	AfterEach(func() {
		unsetServerAdminCredentials()
		if postgresContainer != nil {
			defer postgresContainer.Terminate(context.Background())
			var logs []byte
			logReader, err := postgresContainer.Logs(context.Background())
			if err != nil {
				Fail("Failed to get log reader from postgres container")
				return
			}
			_, err = logReader.Read(logs)
			if err != nil {
				Fail("Failed to read logs from postgres container")
				return
			}
			GinkgoWriter.Write(logs)
		}
	})
})
