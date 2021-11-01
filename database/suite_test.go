package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var postgresContainer testcontainers.Container

func SetServerAdminCredentials(port string, host string) {
	os.Setenv("PGPASSWORD", "test")
	os.Setenv("PGUSER", "server_admin")
	os.Setenv("PGHOST", host)
	os.Setenv("PGPORT", port)
	os.Setenv("PGDATABASE", "postgres")
}

func UnsetServerAdminCredentials() {
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
func TestDatabase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Database Suite")
}

var _ = BeforeSuite(func() {
	const postgresStartupTimeout = time.Second * 10

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
	SetServerAdminCredentials(mappedPort.Port(), ip)
})

var _ = AfterSuite(func() {
	UnsetServerAdminCredentials()
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
