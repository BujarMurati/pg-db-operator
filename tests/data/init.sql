/*
* For our test setup we create a situation simular to the one described in:
* https://docs.microsoft.com/en-us/azure/postgresql/howto-create-users
* We assume a high power user who does is not quite a superuser but can create users, databases, etc.
*/

CREATE USER server_admin WITH LOGIN NOSUPERUSER INHERIT CREATEDB CREATEROLE REPLICATION ENCRYPTED PASSWORD 'test';
