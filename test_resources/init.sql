/*
* For our test setup we create a situation simular to the one described in:
* https://docs.microsoft.com/en-us/azure/postgresql/howto-create-users
* We assume a high power user who does is not quite a superuser but can create users, databases, etc.
*/

CREATE USER server_admin WITH LOGIN NOSUPERUSER INHERIT CREATEDB CREATEROLE REPLICATION ENCRYPTED PASSWORD 'test';

/*
* Setup for a test case in which both the database and user already exist
*/
CREATE DATABASE everything_exists;
CREATE USER everything_exists WITH ENCRYPTED PASSWORD 'everything_exists';
GRANT ALL PRIVILEGES ON DATABASE everything_exists TO everything_exists;


/*
* Setup for a test case in which the database already exists but the user is missing
*/
CREATE DATABASE database_exists;

/*
* Setup for a test case in which the user already exists but the database is missing
*/
CREATE USER user_exists;


/*
* Setup for test cases in which both user and database exist but the user has no privileges on the database
*/
CREATE USER has_no_privileges;
CREATE DATABASE has_no_privileges;

/*
* Setup for a test case in which both the database and user already exist but the
* user has no privileges for the database
*/
CREATE DATABASE user_and_db_but_no_privileges;
CREATE USER user_and_db_but_no_privileges WITH ENCRYPTED PASSWORD 'user_and_db_but_no_privileges';