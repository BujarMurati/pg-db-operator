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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	dbv1beta1 "github.com/bujarmurati/pg-db-operator/api/v1beta1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
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
	packageName := "controllers"
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

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
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
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = dbv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&PostgresDatabaseReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())

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
