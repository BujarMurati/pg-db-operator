package database

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DatabaseServer", func() {

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
	Describe("CheckUserExists", func() {
		It("Returns false if the user does not exist", func() {
			userName := "user_does_not_exist"
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			exists, err := db.CheckUserExists(userName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})
		It("Returns true if the user already does exist", func() {
			userName := "user_exists"
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			exists, err := db.CheckUserExists(userName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
	})
	Describe("CreateUserIfNotExists", func() {
		It("Creates a new user if it doesn't exist", func() {
			userName := "new_user"
			password := "test"
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			exists, err := db.CheckUserExists(userName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())

			Expect(db.CreateUserOrUpdatePassword(userName, password)).NotTo(HaveOccurred())
			exists, err = db.CheckUserExists(userName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
		It("Skips creating a new database if it doesn't exist", func() {
			userName := "user_exists"
			password := "test"
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			exists, err := db.CheckUserExists(userName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			Expect(db.CreateUserOrUpdatePassword(userName, password)).NotTo(HaveOccurred())
			exists, err = db.CheckUserExists(userName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
	})
})
