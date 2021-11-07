package database

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DatabaseServer", func() {

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
	Describe("CheckUserHasAllPrivileges", func() {
		It("Returns false if the user does not have all privileges on the database", func() {
			userName := "has_no_privileges"
			databaseName := "has_no_privileges"
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			exists, err := db.CheckUserHasAllPrivileges(userName, databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})
		It("Returns true if the user has all privileges on the database", func() {
			userName := "everything_exists"
			databaseName := "everything_exists"
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			exists, err := db.CheckUserHasAllPrivileges(userName, databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
	})
	Describe("EnsureUserHasAllPrivileges", func() {
		It("Grants all privileges on a database to a user if not yet granted", func() {
			userName := "new_privileges"
			databaseName := "new_privileges"
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())

			Expect(db.CreateDatabaseIfNotExists(databaseName)).NotTo(HaveOccurred())
			Expect(db.CreateUserOrUpdatePassword(userName, "test")).NotTo(HaveOccurred())

			exists, err := db.CheckUserHasAllPrivileges(userName, databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())

			Expect(db.EnsureUserHasAllPrivileges(userName, databaseName)).NotTo(HaveOccurred())
			exists, err = db.CheckUserHasAllPrivileges(userName, databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
		It("Skips granting privileges that already exist", func() {
			userName := "everything_exists"
			databaseName := "everything_exists"
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			exists, err := db.CheckUserHasAllPrivileges(userName, databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			Expect(db.EnsureUserHasAllPrivileges(userName, databaseName)).NotTo(HaveOccurred())
			exists, err = db.CheckUserHasAllPrivileges(userName, databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
	})
	Describe("ReconcileDatabaseState", func() {
		It("Creates database, user and privileges", func() {
			userName := "desired_user"
			databaseName := "desired_database"
			password := "test"
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())

			Expect(db.ReconcileDatabaseState(userName, databaseName, password))

			exists, err := db.CheckDatabaseExists(databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			exists, err = db.CheckUserExists(userName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			exists, err = db.CheckUserHasAllPrivileges(userName, databaseName)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
	})
	Describe("CheckConnection", func() {
		It("Fails ewith incorrect connection details", func() {
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			password := "wrong-password"
			user := "wrong-user"
			database := "non-existing-database"
			config := *db.GetConfig()
			config.Config.Password = password
			config.Config.User = user
			config.Config.Database = database
			Expect(CheckConnection(config)).NotTo(Succeed())
		})
		It("Succeeds with correct connection details", func() {
			db, err := NewDatabaseServerFromEnvironment()
			Expect(err).NotTo(HaveOccurred())
			password := "correct_password"
			user := "correct_user"
			database := "newly_created_database"
			Expect(db.ReconcileDatabaseState(user, database, password)).To(Succeed())
			config := *db.GetConfig()
			config.Config.Password = password
			config.Config.User = user
			config.Config.Database = database
			Expect(CheckConnection(config)).To(Succeed())
		})
	})
})
